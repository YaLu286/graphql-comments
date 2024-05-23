package storage

import (
	"context"
	"errors"
	"graphql-comments/config"
	"graphql-comments/models"
	"graphql-comments/storage/inmemory"
	"graphql-comments/storage/postgres"
	"io"
)

const (
	StorageTypePostgres string = "postgres"
	StorageTypeInmemory string = "inmemory"
)

type Storager interface {
	// Сохраняет пост в хранилище, возвращает созданный пост или ошибку
	CreatePost(ctx context.Context, p models.Post) (models.Post, error)

	// Сохраняет комментарии в хранилище, возвращает созданный комментарии или ошибку
	CreateComment(ctx context.Context, c models.Comment, parentID *int) (models.Comment, error)

	// Находит пост в хранилище по id
	GetPost(ctx context.Context, id int) (*models.Post, error)

	// Получает слайс всех постов из хранилища
	GetPosts(ctx context.Context) ([]*models.Post, error)

	// Получает сплайс комментариев в треде под постом с id = postID или под комментарием с id = parenID.
	// Поддерживается keyset пагинация
	GetComments(ctx context.Context, postID int, parentID *int, limit, afterID int) ([]*models.Comment, error)

	// Великий закрыватор
	io.Closer
}

// Конструктор хранилища, выбирает реализацию на основании конфигурации
func New(cfg *config.Config) (Storager, error) {
	switch cfg.StorageType {
	case StorageTypePostgres:
		return postgres.NewPostgresStorage(cfg)
	case StorageTypeInmemory:
		return inmemory.NewMemoryStorage()
	default:
		return nil, errors.New("unknown storage type")
	}
}
