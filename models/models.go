package models

import (
	"errors"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"io"
	"strconv"
	"time"
)

// структура описывает пост
type Post struct {
	ID            int       `json:"id"`
	Title         string    `json:"title"`
	Author        string    `json:"author"`
	Content       string    `json:"content"`
	AllowComments bool      `json:"AllowComments"`
	CreatedAt     time.Time `json:"createdAt"`
	Comments      []Comment
}

// структура описывает комментарии под постом
type Comment struct {
	ID         int       `json:"id"`
	PostID     int       `json:"post_id"`
	ParentID   *int      `json:"parent_id"`
	Author     string    `json:"author"`
	Text       string    `json:"text"`
	CreatedAt  time.Time `json:"createdAt"`
	HasReplies bool      `json:"hasReplies"`
}

func MarshalID(id int) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.Quote(fmt.Sprintf("%d", id)))
	})
}

func UnmarshalID(v interface{}) (int, error) {
	id, ok := v.(string)
	if !ok {
		return 0, fmt.Errorf("ids must be strings")
	}
	i, err := strconv.Atoi(id)
	if err != nil {
		return 0, errors.New("error occured while parsing id")
	}
	return int(i), err
}

// маршалер скалярного пользовательского типа Timestamp
func MarshalTimestamp(t time.Time) graphql.Marshaler {

	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, "\""+t.Format("02.01.2006 15:04:05 MST")+"\"")
	})
}

// анмаршалер скалярного пользовательского типа Timestamp
func UnmarshalTimestamp(v interface{}) (time.Time, error) {
	if tmpStr, ok := v.(int); ok {
		return time.Unix(int64(tmpStr), 0), nil
	}
	return time.Time{}, errors.New("wrong timestamp")
}
