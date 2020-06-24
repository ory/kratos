package main

import (
	"database/sql"
	"os"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	pgu := os.Getenv("TEST_DATABASE_POSTGRESQL")
	if len(pgu) == 0 {
		panic("pg url not set")
	}

	c, err := sql.Open("pgx", pgu)
	check(err)
	_, err = c.Exec(`
CREATE TABLE IF NOT EXISTS "table_a" (
	"id" UUID NOT NULL,
	PRIMARY KEY("id")
);
CREATE TABLE IF NOT EXISTS "table_b" (
	"id" UUID NOT NULL,
	PRIMARY KEY("id"),
	"table_a_id" UUID NOT NULL,
	FOREIGN KEY ("table_a_id") REFERENCES "table_a" ("id") ON DELETE cascade
);
DELETE FROM table_a;
DELETE FROM table_b;
`)
	check(err)

	tx, err := c.Begin()
	check(err)

	var run = func() error {
		// works
		if _, err := tx.Exec("INSERT INTO table_a (id) VALUES ($1)", "0f820d3d-ff15-4fab-b1c9-3ed87a0aaee8"); err != nil {
			return err
		}

		// works
		if _, err := tx.Exec("INSERT INTO table_b (id, table_a_id) VALUES ($1, $2)", "06b6266a-65aa-43f2-8ad3-0fb04be7b691", "0f820d3d-ff15-4fab-b1c9-3ed87a0aaee8"); err != nil {
			return err
		}

		_,_= tx.Exec("SELECT * FROM table_a WHERE id=$1", "c91e8117-b8c2-4972-ae7f-53cd6b3eeebc")
		// _,_= tx.Exec("INSERT INTO table_b (id, table_a_id) VALUES ($1, $2)", "88b2d8ee-7988-4040-b91e-36531930d9a5", "c91e8117-b8c2-4972-ae7f-53cd6b3eeebc")

		// fails
		if _, err := tx.Exec("INSERT INTO table_b (id, table_a_id) VALUES ($1, $2)", "88b2d8ee-7988-4040-b91e-36531930d9a5", "c91e8117-b8c2-4972-ae7f-53cd6b3eeebc"); err != nil {
			return err
		}

		return nil
	}

	err = run()
	if err != nil {
		check(tx.Rollback())
	} else {
		check(tx.Commit())
	}
	check(err)
}
