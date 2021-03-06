package pg

import (
	"database/sql"
	"log"
	"os/exec"
	"testing"

	_ "github.com/lib/pq"
	"github.com/mcls/nomad"
)

func setupDatabase(t *testing.T) *sql.DB {
	// Setup database
	createdb := exec.Command("createdb", "nomad_db_test")
	err := createdb.Run()
	if err != nil {
		log.Println("createdb already ran")
		log.Println(err)
	}

	db, err := sql.Open("postgres", "dbname=nomad_db_test sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`
	DROP TABLE IF EXISTS schema_migrations;
	DROP TABLE IF EXISTS users;
	DROP TABLE IF EXISTS blogs;
	`)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestImplementsVersionStoreInterface(t *testing.T) {
	var _ nomad.VersionStore = NewVersionStore(nil)
}

func TestPostgresVersionStoreWorks(t *testing.T) {
	db := setupDatabase(t)
	versioner := NewVersionStore(db)

	err := versioner.SetupVersionStore()
	if err != nil {
		log.Fatal(err)
	}

	if versioner.HasVersion("123") {
		t.Fatalf("Shouldn't have version %s", "123")
	}

	versioner.AddVersion("123")

	if !versioner.HasVersion("123") {
		t.Fatalf("Should have version %s", "123")
	}
}

func TestRunningMigrations(t *testing.T) {
	db := setupDatabase(t)

	l := nomad.NewList()
	l.Add(&nomad.Migration{
		Version: "A",
		Up: func(ctx interface{}) error {
			c := ctx.(*Context)
			_, err := c.Tx.Exec("CREATE TABLE users (id serial PRIMARY KEY, username text);")
			if err != nil {
				log.Println(err)
				return err
			}
			_, err = c.Tx.Exec("INSERT INTO users (username) VALUES ('mcls')")
			return err
		},
	})

	runner := NewRunner(db, l)
	if err := runner.Run(); err != nil {
		t.Fatal(err)
	}

	if !runner.HasVersion("A") {
		t.Fatal("Should have version A")
	}

	username := ""
	err := db.QueryRow("SELECT username FROM users").Scan(&username)
	if err != nil {
		t.Fatal(err)
	}

	if username != "mcls" {
		t.Fatalf("Expected username 'mcls' was '%s'", username)
	}
}

func TestRollingBackMigration(t *testing.T) {
	db := setupDatabase(t)
	l := nomad.NewList()
	l.Add(&nomad.Migration{
		Version: "A",
		Up: func(ctx interface{}) error {
			c := ctx.(*Context)
			_, err := c.Tx.Exec("CREATE TABLE users (id serial PRIMARY KEY, username text);")
			if err != nil {
				log.Println(err)
				return err
			}
			_, err = c.Tx.Exec("INSERT INTO users (username) VALUES ('mcls')")
			return err
		},
		Down: func(ctx interface{}) error {
			return nil
		},
	})
	l.Add(&nomad.Migration{
		Version: "B",
		Up: func(ctx interface{}) error {
			c := ctx.(*Context)
			_, err := c.Tx.Exec("CREATE TABLE blogs (id serial PRIMARY KEY, content text);")
			if err != nil {
				log.Println(err)
				return err
			}
			return nil
		},
		Down: func(ctx interface{}) error {
			c := ctx.(*Context)
			_, err := c.Tx.Exec("DROP TABLE blogs")
			if err != nil {
				log.Println(err)
				return err
			}
			return nil
		},
	})

	runner := NewRunner(db, l)
	if err := runner.Run(); err != nil {
		log.Fatal(err)
	}

	if !runner.HasVersion("B") {
		t.Fatal("Should have version B")
	}

	// Rollback the last migration
	if err := runner.Rollback(); err != nil {
		log.Fatal(err)
	}

	if !runner.HasVersion("A") {
		t.Fatal("Should have version A")
	}

	if runner.HasVersion("B") {
		t.Fatal("Should NOT have version B")
	}
}
