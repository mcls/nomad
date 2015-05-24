// Helps with the integration of nomad with postgres
package pg

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type PgVersioner struct {
	DB *sql.DB
}

func NewPgVersioner(db *sql.DB) *PgVersioner {
	return &PgVersioner{DB: db}
}

func (vs *PgVersioner) HasVersion(v string) bool {
	var found string
	err := vs.DB.QueryRow("SELECT * FROM schema_migrations WHERE version = $1", v).Scan(&found)
	switch {
	case err == sql.ErrNoRows:
		return false
	case err != nil:
		panic(err)
	default:
		return true
	}
}

func (vs *PgVersioner) AddVersion(v string) {
	_, err := vs.DB.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", v)
	if err != nil {
		panic(err)
	}
}

// SetupVersions creates the schema_migrations table to store the versions
func (vs *PgVersioner) SetupVersions() error {
	_, err := vs.DB.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
  version text NOT NULL UNIQUE
)`)
	return err
}
