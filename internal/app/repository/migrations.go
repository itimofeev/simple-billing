package repository

import (
	"github.com/go-pg/migrations/v8"
	"github.com/go-pg/pg/v10"
)

func doMigrationIfNeeded(db *pg.DB) error {
	col := getMigrations()
	_, _, _ = col.Run(db, "init") // nolint:dogsled
	_, _, err := col.Run(db, "up")
	return err
}

func getMigrations() *migrations.Collection {
	sqls := getSQLs()

	migs := make([]*migrations.Migration, 0, len(sqls))

	for i, sql := range sqls {
		sql := sql
		migs = append(migs, &migrations.Migration{
			Version: int64(i + 1),
			UpTx:    true,
			Up: func(db migrations.DB) error {
				_, err := db.Exec(sql)
				return err
			},
		})
	}

	return migrations.NewCollection(migs...)
}

func getSQLs() []string {
	return []string{`
DROP TABLE IF EXISTS balances;

CREATE TABLE balances
(
    id      BIGSERIAL PRIMARY KEY,
    balance BIGINT NOT NULL
);
	`,
	}
}
