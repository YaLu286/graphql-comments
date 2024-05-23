package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"graphql-comments/config"
	"graphql-comments/models"
	"time"
)

var (
	ErrPostNotFound          = fmt.Errorf("post not found")
	ErrCommentsAreNotAllowed = fmt.Errorf("comments are not allowed for this post")
	ErrParentCommentNotFound = fmt.Errorf("parent comment not found")
)

type PostgresStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresStorage(cfg *config.Config) (*PostgresStorage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ReadTimeout)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	poolConfig.MaxConns = int32(cfg.PostgresMaxConn)

	pool, err := pgxpool.ConnectConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	return &PostgresStorage{pool: pool}, nil
}

func (s *PostgresStorage) CreatePost(ctx context.Context, p models.Post) (models.Post, error) {
	query := `INSERT INTO posts (title, author, content, allow_comments) VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	row := s.pool.QueryRow(ctx, query, p.Title, p.Author, p.Content, p.AllowComments)

	err := row.Scan(&p.ID, &p.CreatedAt)
	return p, err
}

func (s *PostgresStorage) CreateComment(ctx context.Context, c models.Comment, parentID *int) (models.Comment, error) {
	var allowComments bool
	query := `SELECT allow_comments FROM posts WHERE id=$1`
	err := s.pool.QueryRow(ctx, query, c.PostID).Scan(&allowComments)
	if err != nil {
		return c, ErrPostNotFound
	}

	if !allowComments {
		return c, ErrCommentsAreNotAllowed
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return c, err
	}
	defer tx.Rollback(ctx)

	if parentID != nil {
		// Проверяем существует ли родительский комментарий
		var parentExists bool
		err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM comments WHERE id=$1)`, *parentID).Scan(&parentExists)
		if err != nil {
			return c, err
		}
		if !parentExists {
			return c, ErrParentCommentNotFound
		}

		// Устанавливаем HasReplies равным true у родительского комментария
		_, err = tx.Exec(ctx, `UPDATE comments SET has_replies = true WHERE id = $1`, *parentID)
		if err != nil {
			return c, err
		}

	}

	query = `INSERT INTO comments (post_id, text, author, created_at) VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	row := tx.QueryRow(ctx, query, c.PostID, c.Text, c.Author, time.Now())
	err = row.Scan(&c.ID, &c.CreatedAt)
	if err != nil {
		return c, err
	}

	if parentID != nil {
		_, err = tx.Exec(ctx, `INSERT INTO comment_hierarchy (parent_id, child_id) VALUES ($1, $2)`, *parentID, c.ID)
		if err != nil {
			return c, err
		}
	}

	err = tx.Commit(ctx)
	return c, err

}

func (s *PostgresStorage) GetPost(ctx context.Context, id int) (*models.Post, error) {
	query := `SELECT id, title, author, content, created_at, allow_comments FROM posts WHERE id=$1`
	row := s.pool.QueryRow(ctx, query, id)

	var post models.Post
	err := row.Scan(&post.ID, &post.Title, &post.Author, &post.Content, &post.CreatedAt, &post.AllowComments)
	if err == pgx.ErrNoRows {
		return nil, ErrPostNotFound
	}
	return &post, err

}

func (s *PostgresStorage) GetPosts(ctx context.Context) ([]*models.Post, error) {
	query := `SELECT id, title, content, created_at, allow_comments FROM posts ORDER BY created_at DESC`
	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		var post models.Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.CreatedAt, &post.AllowComments); err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}

	return posts, nil
}

func (s *PostgresStorage) GetComments(ctx context.Context, postID int, parentID *int, limit, afterID int) ([]*models.Comment, error) {
	var rows pgx.Rows
	var err error

	if parentID == nil {
		query := `SELECT id, post_id, author, text, created_at, has_replies FROM comments WHERE post_id=$1 AND id NOT IN 
				(SELECT child_id FROM comment_hierarchy) AND id > $2 
				ORDER BY created_at LIMIT $3`
		rows, err = s.pool.Query(ctx, query, postID, afterID, limit)
	} else {
		query := `SELECT c.id, c.post_id, c.author, c.text, c.created_at, c.has_replies FROM comments c 
				JOIN comment_hierarchy ch ON c.id = ch.child_id 
				WHERE ch.parent_id = $1 AND c.id > $2 
				ORDER BY c.created_at LIMIT $3`
		rows, err = s.pool.Query(ctx, query, *parentID, afterID, limit)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*models.Comment
	for rows.Next() {
		var c models.Comment
		if err := rows.Scan(&c.ID, &c.PostID, &c.Author, &c.Text, &c.CreatedAt, &c.HasReplies); err != nil {
			return nil, err
		}
		comments = append(comments, &c)
	}

	return comments, nil
}

func (s *PostgresStorage) Close() error {
	s.pool.Close()
	return nil
}
