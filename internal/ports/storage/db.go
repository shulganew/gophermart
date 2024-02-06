package storage

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Repo struct {
	master *sqlx.DB
}

func NewRepo(ctx context.Context, master *sqlx.DB) (*Repo, error) {
	db := Repo{master: master}
	err := db.Start(ctx)
	return &db, err
}

func (base *Repo) Start(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	err := base.master.PingContext(ctx)
	defer cancel()
	return err
}
