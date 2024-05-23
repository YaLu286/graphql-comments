package inmemory

import (
	"context"
	"fmt"
	"graphql-comments/models"
	"sort"
	"sync"
	"time"
)

var (
	ErrPostNotFound          = fmt.Errorf("post not found")
	ErrCommentsAreNotAllowed = fmt.Errorf("comments are not allowed for this post")
	ErrParentCommentNotFound = fmt.Errorf("parent comment not found")
)

// структура описывает хранилище в памяти
type InMemoryStorage struct {
	posts            map[int]*models.Post      // хеш-таблица для хранения постов, где ключ это id поста
	comments         map[int][]*models.Comment // хеш-таблица для хранения коментариев первого уровня под постом, где ключ это id поста
	commentHierarchy map[int][]*models.Comment // хеш-таблица для хранения коментариев последующих уровней под постом, где ключ это id родительского комментария
	postCounter      int                       // cчетчик числа постов
	commentCounter   int                       // cчетчик числа комментариев
	postMu           sync.RWMutex
	commentMu        sync.RWMutex
	hierarchyMu      sync.RWMutex
}

// Конструктор inmemory хранилища
func NewMemoryStorage() (*InMemoryStorage, error) {
	return &InMemoryStorage{
		posts:            make(map[int]*models.Post),
		comments:         make(map[int][]*models.Comment),
		commentHierarchy: make(map[int][]*models.Comment),
	}, nil
}

// Сохраняет пост в памяти, возвращает созданный пост или ошибку
func (s *InMemoryStorage) CreatePost(ctx context.Context, p models.Post) (models.Post, error) {
	s.postMu.Lock()
	defer s.postMu.Unlock()

	s.postCounter++
	p.ID = s.postCounter
	p.CreatedAt = time.Now()
	s.posts[p.ID] = &p

	return p, nil
}

// Сохраняет комментарии в памяти, возвращает созданный комментарии или ошибку
func (s *InMemoryStorage) CreateComment(ctx context.Context, c models.Comment, parentID *int) (models.Comment, error) {
	s.commentMu.RLock()
	s.hierarchyMu.RLock()
	defer s.commentMu.RUnlock()
	defer s.hierarchyMu.RUnlock()

	s.postMu.RLock()
	post, exists := s.posts[c.PostID]
	s.postMu.RUnlock()
	if !exists {
		return c, ErrPostNotFound
	}

	if !post.AllowComments {
		return c, ErrCommentsAreNotAllowed
	}

	s.commentCounter++
	c.ID = s.commentCounter
	c.CreatedAt = time.Now()
	c.HasReplies = false

	if parentID != nil {
		// если указан id родительского коммента, то сначала находим его
		parentComment := s.findComment(c.PostID, *parentID)
		if parentComment == nil {
			return c, ErrParentCommentNotFound
		}
		// теперь у родительского коммента есть дрочерние, фиксируем это
		parentComment.HasReplies = true
		s.commentHierarchy[*parentID] = append(s.commentHierarchy[*parentID], &c)
	} else {
		s.comments[c.PostID] = append(s.comments[c.PostID], &c)
	}

	return c, nil
}

// Находит пост в памяти по id
func (s *InMemoryStorage) GetPost(ctx context.Context, id int) (*models.Post, error) {
	s.postMu.RLock()
	defer s.postMu.RUnlock()

	post, exists := s.posts[id]
	if !exists {
		return nil, ErrPostNotFound
	}

	return post, nil
}

// Получает слайс всех постов из памяти, посты сортированы от нового к старому
func (s *InMemoryStorage) GetPosts(ctx context.Context) ([]*models.Post, error) {
	s.postMu.RLock()
	defer s.postMu.RUnlock()

	posts := make([]*models.Post, 0, len(s.posts))
	for _, post := range s.posts {
		posts = append(posts, post)
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].CreatedAt.After(posts[j].CreatedAt)
	})

	return posts, nil
}

// Получает сплайс комментариев c id > afterID под постом с id = postID длинной limit
func (s *InMemoryStorage) GetComments(ctx context.Context, postID int, parentID *int, limit, afterID int) ([]*models.Comment, error) {
	s.commentMu.RLock()
	s.hierarchyMu.RLock()
	defer s.commentMu.RUnlock()
	defer s.hierarchyMu.RUnlock()

	s.postMu.RLock()
	_, postExists := s.posts[postID]
	s.postMu.RUnlock()
	if !postExists {
		return nil, ErrPostNotFound
	}

	var comments []*models.Comment
	if parentID == nil {
		comments = s.comments[postID]
	} else {

		parentComment := s.findComment(postID, *parentID)
		if parentComment == nil {
			return nil, ErrParentCommentNotFound
		}

		comments = s.commentHierarchy[*parentID]
	}

	sort.Slice(comments, func(i, j int) bool {
		return comments[i].CreatedAt.Before(comments[j].CreatedAt)
	})

	if afterID > 0 {
		var startIndex int
		for i, comment := range comments {
			if comment.ID == afterID {
				startIndex = i + 1
				break
			}
		}
		if startIndex >= len(comments) {
			return nil, nil
		}
		comments = comments[startIndex:]
	}

	end := limit
	if end > len(comments) {
		end = len(comments)
	} else if end < 0 {
		end = 0
	}

	return comments[:end], nil
}

// Не делаем ничего, но тем самым реализуем интерфейс Storager
func (s *InMemoryStorage) Close() error {
	return nil
}

// Вспомогательная функция находит комментарий с заданным id под постом с postID
func (s *InMemoryStorage) findComment(postID, commentID int) *models.Comment {

	for _, comment := range s.comments[postID] {
		if comment.ID == commentID {
			return comment
		}
	}

	for _, comments := range s.commentHierarchy {
		for _, comment := range comments {
			if comment.ID == commentID {
				return comment
			}
		}
	}

	return nil
}
