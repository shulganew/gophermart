package storage

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type Repo struct {
	master *pgx.Conn
}

func NewRepo(ctx context.Context, master *pgx.Conn) (*Repo, error) {
	db := Repo{master: master}
	err := db.Start(ctx)
	return &db, err
}

// Init Database
func InitDB(ctx context.Context, dsn string) (db *pgx.Conn, err error) {

	db, err = pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}

	return
}

func (base *Repo) Start(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	err := base.master.Ping(ctx)
	defer cancel()
	return err
}

func (base *Repo) GetOrder() {

}
