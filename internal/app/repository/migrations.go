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
	return []string{
		`
DROP TABLE IF EXISTS balances;

CREATE TABLE balances
(
    id      BIGSERIAL NOT NULL PRIMARY KEY,
    balance BIGINT NOT NULL,
    CHECK ( balance >= 0 )
);
`,

		`DROP TABLE IF EXISTS events;

CREATE TABLE events
(
    id              BIGSERIAL PRIMARY KEY      NOT NULL,

    type            VARCHAR(32)                NOT NULL,

    from_user_id    BIGINT REFERENCES balances NOT NULL,
    to_user_id      BIGINT REFERENCES balances,
    amount          BIGINT,

    created_time    timestamptz                NOT NULL,

    queue_id        VARCHAR(128) UNIQUE        NOT NULL,
    queue_sent_time timestamptz,

    CHECK (type IN ('open', 'deposit', 'withdraw', 'transfer')),
    CHECK (type = 'transfer' OR to_user_id IS NULL),
    CHECK (type <> 'transfer' OR to_user_id IS NOT NULL),
    CHECK (type = 'open' OR amount IS NOT NULL),
    CHECK (type <> 'open' OR amount IS NULL)
);

CREATE UNIQUE INDEX events__queue_id__idx ON events (queue_id);
`,
	}
}
