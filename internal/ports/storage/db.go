package storage

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Repo struct {
	db *sqlx.DB
}

func NewRepo(ctx context.Context, master *sqlx.DB) (*Repo, error) {
	db := Repo{db: master}
	err := db.Start(ctx)
	return &db, err
}

func (r *Repo) Start(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	err := r.db.PingContext(ctx)
	defer cancel()
	return err
}

func (r *Repo) DB() *sqlx.DB {
	return r.db
}
