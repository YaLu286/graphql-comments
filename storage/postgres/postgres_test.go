package postgres

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"graphql-comments/config"
	"graphql-comments/models"
	"testing"
	"time"
)

// Функция для очистки базы данных
func cleanDB(ctx context.Context, storage *PostgresStorage) {
	_, _ = storage.pool.Exec(ctx, `TRUNCATE comments, comment_hierarchy, posts RESTART IDENTITY CASCADE`)
}

// Инициализация хранилища для тестов
// Перед запускам поднимаем тестовую бд в контейнере и запускам на ней скрипт создания таблиц
// docker run -d -p 5432:5432 -e "TZ=Europe/Moscow" -e POSTGRES_USER=user -e POSTGRES_DB=gqldb -e POSTGRES_PASSWORD=password postgres
func setupStorage(t *testing.T) *PostgresStorage {
	cfg := &config.Config{
		DatabaseURL:     "postgres://user:password@localhost:5432/gqldb?sslmode=disable",
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		PostgresMaxConn: 10,
	}

	storage, err := NewPostgresStorage(cfg)
	assert.NoError(t, err)

	ctx := context.Background()
	cleanDB(ctx, storage)
	return storage
}

func TestCreatePost(t *testing.T) {
	storage := setupStorage(t)
	ctx := context.Background()

	post := models.Post{
		Title:         "Тест",
		Content:       "Что-нибудь",
		Author:        "Вася",
		AllowComments: true,
	}

	createdPost, err := storage.CreatePost(ctx, post)
	assert.NoError(t, err)
	assert.NotZero(t, createdPost.ID)
	assert.WithinDuration(t, time.Now(), createdPost.CreatedAt, time.Second)
	assert.Equal(t, "Тест", createdPost.Title)
	assert.Equal(t, "Что-нибудь", createdPost.Content)
	assert.Equal(t, "Вася", createdPost.Author)
	assert.True(t, createdPost.AllowComments)
}

func TestCreateComment(t *testing.T) {
	storage := setupStorage(t)
	ctx := context.Background()

	post := models.Post{
		Title:         "Тест",
		Content:       "Что-нибудь",
		Author:        "Гена",
		AllowComments: true,
	}

	createdPost, err := storage.CreatePost(ctx, post)
	assert.NoError(t, err)

	comment := models.Comment{
		PostID: createdPost.ID,
		Text:   "Баян",
		Author: "Вася",
	}

	createdComment, err := storage.CreateComment(ctx, comment, nil)
	assert.NoError(t, err)
	assert.NotZero(t, createdComment.ID)
	assert.WithinDuration(t, time.Now(), createdComment.CreatedAt, time.Second)
	assert.Equal(t, "Баян", createdComment.Text)
	assert.Equal(t, "Вася", createdComment.Author)
	assert.Equal(t, createdPost.ID, createdComment.PostID)
}

func TestCreateCommentForNonExistingPost(t *testing.T) {
	storage := setupStorage(t)
	ctx := context.Background()

	comment := models.Comment{
		PostID: 1337, // несуществующий айди поста
		Text:   "Огонь",
		Author: "Андрей",
	}

	_, err := storage.CreateComment(ctx, comment, nil)
	assert.Error(t, err)
	assert.Equal(t, ErrPostNotFound, err)
}

func TestCreateCommentWhenCommentsNotAllowed(t *testing.T) {
	storage := setupStorage(t)
	ctx := context.Background()

	post := models.Post{
		Title:         "Тест",
		Content:       "Простыня",
		Author:        "Вася",
		AllowComments: false,
	}

	createdPost, err := storage.CreatePost(ctx, post)
	assert.NoError(t, err)

	comment := models.Comment{
		PostID: createdPost.ID,
		Text:   "Много букв",
		Author: "Петя",
	}

	_, err = storage.CreateComment(ctx, comment, nil)
	assert.Error(t, err)
	assert.Equal(t, ErrCommentsAreNotAllowed, err)
}

func TestGetPost(t *testing.T) {
	storage := setupStorage(t)
	ctx := context.Background()

	post := models.Post{
		Title:         "Тест",
		Content:       "Пост",
		Author:        "Я",
		AllowComments: true,
	}

	createdPost, err := storage.CreatePost(ctx, post)
	assert.NoError(t, err)

	retrievedPost, err := storage.GetPost(ctx, createdPost.ID)
	assert.NoError(t, err)
	assert.Equal(t, createdPost, *retrievedPost)
}

func TestGetPostNotFound(t *testing.T) {
	storage := setupStorage(t)
	ctx := context.Background()

	_, err := storage.GetPost(ctx, 228) // несуществующий айди поста
	assert.Error(t, err)
	assert.Equal(t, ErrPostNotFound, err)
}

func TestGetPosts(t *testing.T) {
	storage := setupStorage(t)
	ctx := context.Background()

	post1 := models.Post{
		Title:         "Тест1",
		Content:       "Первый",
		Author:        "1",
		AllowComments: true,
	}
	post2 := models.Post{
		Title:         "Тест2",
		Content:       "Второй",
		Author:        "2",
		AllowComments: true,
	}

	createdPost1, err := storage.CreatePost(ctx, post1)
	assert.NoError(t, err)
	time.Sleep(time.Millisecond)
	createdPost2, err := storage.CreatePost(ctx, post2)
	assert.NoError(t, err)

	posts, err := storage.GetPosts(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(posts))
	assert.Equal(t, createdPost2.ID, posts[0].ID) // пост2 должен добавиться позже
	assert.Equal(t, createdPost1.ID, posts[1].ID)
}

func TestGetComments(t *testing.T) {
	storage := setupStorage(t)
	ctx := context.Background()

	post := models.Post{
		Title:         "Тест",
		Content:       "Пост",
		Author:        "Пушкин",
		AllowComments: true,
	}

	createdPost, err := storage.CreatePost(ctx, post)
	assert.NoError(t, err)

	comment1 := models.Comment{
		PostID: createdPost.ID,
		Text:   "1",
		Author: "Лермонтов",
	}
	comment2 := models.Comment{
		PostID: createdPost.ID,
		Text:   "2",
		Author: "Дантес",
	}

	createdComment1, err := storage.CreateComment(ctx, comment1, nil)
	assert.NoError(t, err)
	time.Sleep(time.Millisecond)
	createdComment2, err := storage.CreateComment(ctx, comment2, nil)
	assert.NoError(t, err)

	comments, err := storage.GetComments(ctx, createdPost.ID, nil, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(comments))
	assert.Equal(t, createdComment1.ID, comments[0].ID)
	assert.Equal(t, createdComment2.ID, comments[1].ID)
}

func TestGetCommentsWithPagination(t *testing.T) {
	storage := setupStorage(t)
	ctx := context.Background()

	post := models.Post{
		Title:         "Тест",
		Content:       "Пост",
		Author:        "Автор",
		AllowComments: true,
	}

	createdPost, err := storage.CreatePost(ctx, post)
	assert.NoError(t, err)

	for i := 0; i < 5; i++ {
		comment := models.Comment{
			PostID: createdPost.ID,
			Text:   "Коммент номер " + fmt.Sprint(i),
			Author: "Уткин",
		}
		_, err := storage.CreateComment(ctx, comment, nil)
		assert.NoError(t, err)
	}

	comments, err := storage.GetComments(ctx, createdPost.ID, nil, 2, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(comments))

	comments, err = storage.GetComments(ctx, createdPost.ID, nil, 2, 2)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(comments))

	comments, err = storage.GetComments(ctx, createdPost.ID, nil, 2, 4)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(comments))
}

func TestGetCommentsHierarchy(t *testing.T) {
	storage := setupStorage(t)
	ctx := context.Background()

	post := models.Post{
		Title:         "Тест",
		Content:       "Пост",
		Author:        "Автор",
		AllowComments: true,
	}

	createdPost, err := storage.CreatePost(ctx, post)
	assert.NoError(t, err)

	comment1 := models.Comment{
		PostID: createdPost.ID,
		Text:   "Первый",
		Author: "1",
	}
	createdComment1, err := storage.CreateComment(ctx, comment1, nil)
	assert.NoError(t, err)

	comment2 := models.Comment{
		PostID: createdPost.ID,
		Text:   "Ответ на первый",
		Author: "2",
	}
	createdComment2, err := storage.CreateComment(ctx, comment2, &createdComment1.ID)
	assert.NoError(t, err)

	comments, err := storage.GetComments(ctx, createdPost.ID, &createdComment1.ID, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(comments))
	assert.Equal(t, createdComment2, *comments[0])
}
