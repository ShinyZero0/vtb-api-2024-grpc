package main

import (
	"context"
	"database/sql"
	"errors"
	"time"

	models "codeberg.org/shinyzero0/vtb-api-2024-grpc/server-models"
	"github.com/volatiletech/sqlboiler/v4/boil"
	_ "modernc.org/sqlite"
)

type storage struct {
	r boil.ContextExecutor
	w boil.ContextExecutor
}

// GetHistory implements Storage.
func (s storage) GetHistory(ctx context.Context, until int64, amount int64) ([]msgxuser, error) {
	msgs, err := models.Messages().All(ctx, s.r)
	if err != nil {
		return nil, err
	}
	result := make([]msgxuser, len(msgs))
	for k, v := range msgs {
		result[k] = msgxuser{
			msg: msg{
				Message:   v.Content,
				timestamp: time.Unix(v.Timestamp, 0),
			},
			CN: v.UserID,
			ID: v.MessageID,
		}
	}
	return result, nil
}

// AppendMessage implements Storage.
func (s storage) AppendMessage(ctx context.Context, m msg, sid string) error {
	mm := models.Message{
		UserID:    sid,
		Content:   m.Message,
		Timestamp: m.timestamp.Unix(),
	}
	return mm.Insert(ctx, s.w, boil.Infer())
}

func NewStorage(dsn string) (Storage, error) {
	w, err1 := sql.Open("sqlite", dsn)
	w.SetMaxOpenConns(1)
	r, err2 := sql.Open("sqlite", dsn)
	// db.SetMaxOpenConns()
	return storage{r: r, w: w}, errors.Join(err1, err2)
}
