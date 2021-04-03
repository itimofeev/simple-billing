package repository

import (
	"context"
	"fmt"

	"github.com/go-pg/pg/v10"
)

type Store struct {
	db *pg.DB
}

func New(url string) *Store {
	opts, err := pg.ParseURL(url)
	if err != nil {
		panic(err)
	}
	db := pg.Connect(opts)
	db.AddQueryHook(dbLogger{})
	if err := db.Ping(context.Background()); err != nil {
		panic(err)
	}

	if err := doMigrationIfNeeded(db); err != nil {
		panic(err)
	}

	return &Store{
		db: db,
	}
}

type dbLogger struct{}

func (d dbLogger) BeforeQuery(ctx context.Context, q *pg.QueryEvent) (context.Context, error) {
	return ctx, nil
}

func (d dbLogger) AfterQuery(cctx context.Context, q *pg.QueryEvent) error {
	query, _ := q.FormattedQuery()
	fmt.Println(string(query))
	return nil
}
