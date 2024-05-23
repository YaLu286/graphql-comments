package graph

import (
	"context"
	"graphql-comments/graph/model"
	"graphql-comments/models"
	"graphql-comments/storage/postgres"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockStorage struct {
	posts    []models.Post
	comments []models.Comment
}

func (m *mockStorage) CreatePost(ctx context.Context, post models.Post) (models.Post, error) {
	post.ID = len(m.posts) + 1
	post.CreatedAt = time.Now()
	m.posts = append(m.posts, post)
	return post, nil
}

func (m *mockStorage) CreateComment(ctx context.Context, comment models.Comment, parentID *int) (models.Comment, error) {
	comment.ID = len(m.comments) + 1
	comment.CreatedAt = time.Now()
	m.comments = append(m.comments, comment)
	return comment, nil
}

func (m *mockStorage) GetComments(ctx context.Context, postID int, parentID *int, limit int, afterID int) ([]*models.Comment, error) {
	var filteredComments []*models.Comment
	for _, comment := range m.comments {
		if comment.PostID == postID && (parentID == nil || comment.ParentID == parentID) && comment.ID > afterID {
			filteredComments = append(filteredComments, &comment)
		}
	}
	if limit > 0 && len(filteredComments) > limit {
		filteredComments = filteredComments[:limit]
	}
	return filteredComments, nil
}

func (m *mockStorage) GetPosts(ctx context.Context) ([]*models.Post, error) {
	var posts []*models.Post
	for i := range m.posts {
		posts = append(posts, &m.posts[i])
	}
	return posts, nil
}

func (m *mockStorage) GetPost(ctx context.Context, id int) (*models.Post, error) {
	for i := range m.posts {
		if m.posts[i].ID == id {
			return &m.posts[i], nil
		}
	}
	return nil, postgres.ErrPostNotFound
}

func (m *mockStorage) Close() error {
	return nil
}

func TestCreatePost(t *testing.T) {
	db := &mockStorage{}
	resolver := NewResolver(db)

	input := model.NewPost{
		Title:         "Тест",
		Author:        "Вася",
		Content:       "Что-нибудь",
		AllowComments: true,
	}

	ctx := context.Background()
	post, err := resolver.Mutation().CreatePost(ctx, input)
	assert.NoError(t, err)
	assert.Equal(t, "Тест", post.Title)
	assert.Equal(t, "Вася", post.Author)
	assert.Equal(t, "Что-нибудь", post.Content)
}

func TestCreateComment(t *testing.T) {
	db := &mockStorage{}
	resolver := NewResolver(db)

	postInput := model.NewPost{
		Title:         "Тест",
		Author:        "Гена",
		Content:       "Что-нибудь",
		AllowComments: true,
	}

	ctx := context.Background()
	post, err := resolver.Mutation().CreatePost(ctx, postInput)
	assert.NoError(t, err)

	commentInput := model.NewComment{
		PostID: post.ID,
		Author: "Вася",
		Text:   "Баян",
	}

	comment, err := resolver.Mutation().CreateComment(ctx, commentInput)
	assert.NoError(t, err)
	assert.Equal(t, post.ID, comment.PostID)
	assert.Equal(t, "Вася", comment.Author)
	assert.Equal(t, "Баян", comment.Text)
}

func TestGetPosts(t *testing.T) {
	db := &mockStorage{}
	resolver := NewResolver(db)

	_, err := resolver.Mutation().CreatePost(context.Background(), model.NewPost{
		Title:         "Тест1",
		Author:        "1",
		Content:       "Первый",
		AllowComments: true,
	})
	assert.NoError(t, err)

	_, err = resolver.Mutation().CreatePost(context.Background(), model.NewPost{
		Title:         "Тест2",
		Author:        "2",
		Content:       "Второй",
		AllowComments: true,
	})
	assert.NoError(t, err)

	posts, err := resolver.Query().Posts(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 2, len(posts))
}

func TestSubscriptionNewComment(t *testing.T) {
	db := &mockStorage{}
	resolver := NewResolver(db)

	postInput := model.NewPost{
		Title:         "Тест",
		Author:        "Автор",
		Content:       "Пост",
		AllowComments: true,
	}

	ctx := context.Background()
	post, err := resolver.Mutation().CreatePost(ctx, postInput)
	assert.NoError(t, err)

	commentChan, err := resolver.Subscription().NewComment(ctx, post.ID)
	assert.NoError(t, err)

	commentInput := model.NewComment{
		PostID: post.ID,
		Author: "Петя",
		Text:   "Много букв",
	}

	go func() {
		_, err := resolver.Mutation().CreateComment(ctx, commentInput)
		assert.NoError(t, err)
	}()

	select {
	case comment := <-commentChan:
		assert.Equal(t, post.ID, comment.PostID)
		assert.Equal(t, "Петя", comment.Author)
		assert.Equal(t, "Много букв", comment.Text)
	case <-time.After(2 * time.Second):
		t.Fatal("expected a comment but got none")
	}
}
