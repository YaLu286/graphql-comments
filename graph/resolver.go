package graph

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

import (
	"context"
	"graphql-comments/graph/model"
	"graphql-comments/models"
	"graphql-comments/storage"
	"sync"
)

// структура, которая будет содержать наше хранилище и канал для подписок на комменты
type Resolver struct {
	DB               storage.Storager
	CommentObservers map[int][]chan *models.Comment
	mu               sync.Mutex
}

// Конструктор ресолвера
func NewResolver(db storage.Storager) *Resolver {
	return &Resolver{
		DB:               db,
		CommentObservers: make(map[int][]chan *models.Comment),
	}
}

func (r *mutationResolver) CreatePost(ctx context.Context, input model.NewPost) (*models.Post, error) {
	post := &models.Post{
		Title:         input.Title,
		Author:        input.Author,
		Content:       input.Content,
		AllowComments: input.AllowComments,
	}

	createdPost, err := r.DB.CreatePost(ctx, *post)
	if err != nil {
		return nil, err
	}
	return &createdPost, nil
}

// CreateComment is the resolver for the createComment field.
func (r *mutationResolver) CreateComment(ctx context.Context, input model.NewComment) (*models.Comment, error) {

	comment := &models.Comment{
		PostID:   input.PostID,
		Author:   input.Author,
		ParentID: input.ParentID,
		Text:     input.Text,
	}

	createdComment, err := r.DB.CreateComment(ctx, *comment, comment.ParentID)
	if err != nil {
		return nil, err
	}

	// Уведомляем подписчиков
	for _, observer := range r.CommentObservers[createdComment.PostID] {
		observer <- &createdComment
	}

	return &createdComment, nil
}

func (r *postResolver) Replies(ctx context.Context, obj *models.Post, limit int, afterID int) ([]*models.Comment, error) {

	replies, err := r.DB.GetComments(ctx, obj.ID, nil, limit, afterID)
	if err != nil {
		return nil, err
	}

	return replies, nil
}

func (r *queryResolver) Posts(ctx context.Context) ([]*models.Post, error) {
	posts, err := r.DB.GetPosts(ctx)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func (r *queryResolver) Post(ctx context.Context, id int) (*models.Post, error) {
	post, err := r.DB.GetPost(ctx, id)
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (r *queryResolver) Comments(ctx context.Context, postID int, parentID *int, limit int, afterID int) ([]*models.Comment, error) {
	replies, err := r.DB.GetComments(ctx, postID, parentID, limit, afterID)
	if err != nil {
		return nil, err
	}

	return replies, nil
}

// подписка на новые комментарии
func (r *subscriptionResolver) NewComment(ctx context.Context, postId int) (<-chan *models.Comment, error) {
	commentChan := make(chan *models.Comment)

	r.mu.Lock()
	if _, ok := r.CommentObservers[postId]; !ok {
		r.CommentObservers[postId] = []chan *models.Comment{}
	}
	r.CommentObservers[postId] = append(r.CommentObservers[postId], commentChan)
	r.mu.Unlock()

	go func() {
		<-ctx.Done()
		r.removeSubscriber(postId, commentChan)
		close(commentChan)
	}()

	return commentChan, nil
}

func (r *Resolver) removeSubscriber(postId int, observer chan *models.Comment) {
	r.mu.Lock()
	defer r.mu.Unlock()

	subscribers := r.CommentObservers[postId]
	for i, sub := range subscribers {
		if sub == observer {
			r.CommentObservers[postId] = append(subscribers[:i], subscribers[i+1:]...)
			return
		}
	}
}

func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

func (r *Resolver) Post() PostResolver { return &postResolver{r} }

func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

func (r *Resolver) Subscription() SubscriptionResolver { return &subscriptionResolver{r} }

type mutationResolver struct{ *Resolver }
type postResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
