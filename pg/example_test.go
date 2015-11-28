package pg_test

import (
	"database/sql"
	"log"

	"github.com/mcls/nomad"
	nomadpg "github.com/mcls/nomad/pg"
)

// Example of migration with postgres database. The context object passed to
// the migrations gives you access to the current database transaction, which
// will be rolled back if anything goes wrong
func Example() {
	db, err := sql.Open("postgres", "dbname=nomad_db_test sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	migrations := nomadpg.NewList(db)

	m1 := &nomad.Migration{
		Version: "2015-11-26_19:00:00",
		Up: func(ctx interface{}) error {
			c := ctx.(*nomadpg.Context)
			_, err := c.Tx.Exec(`
				CREATE TABLE users (
					id serial PRIMARY KEY,
					username text
				);`)
			return err
		},
		Down: func(ctx interface{}) error {
			c := ctx.(*nomadpg.Context)
			_, err := c.Tx.Exec(`DROP TABLE users`)
			return err
		},
	}
	m2 := &nomad.Migration{
		Version: "2015-11-26_19:30:00",
		Up: func(ctx interface{}) error {
			c := ctx.(*nomadpg.Context)
			_, err := c.Tx.Exec(`
				CREATE TABLE posts (
					id serial PRIMARY KEY,
					title text
					content text
				);`)
			return err
		},
		Down: func(ctx interface{}) error {
			c := ctx.(*nomadpg.Context)
			_, err := c.Tx.Exec(`DROP TABLE posts`)
			return err
		},
	}
	migrations.Add(m1)
	migrations.Add(m2)

	// Run all pending migrations
	migrations.Run()

	// Rollback the latest migration
	migrations.Rollback()
}
