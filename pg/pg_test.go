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
	`)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestPostgresVersionStoreWorks(t *testing.T) {
	db := setupDatabase(t)
	versioner := NewPgVersioner(db)

	err := versioner.SetupVersions()
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
	versioner := NewPgVersioner(db)

	err := versioner.SetupVersions()
	if err != nil {
		log.Fatal(err)
	}

	l := nomad.NewList(versioner)
	l.Add(&nomad.Migration{
		Version: "A",
		Up: func(ctx interface{}) error {
			db := ctx.(*sql.DB)
			_, err := db.Exec("CREATE TABLE users (id serial PRIMARY KEY, username text);")
			if err != nil {
				log.Println(err)
				return err
			}
			_, err = db.Exec("INSERT INTO users (username) VALUES ('mcls')")
			return err
		},
	})
	l.Run(db)

	if !versioner.HasVersion("A") {
		t.Fatal("Should have version A")
	}

	username := ""
	err = db.QueryRow("SELECT username FROM users").Scan(&username)
	if err != nil {
		t.Fatal(err)
	}

	if username != "mcls" {
		t.Fatalf("Expected username 'mcls' was '%s'", username)
	}
}
