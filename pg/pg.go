// Helps with the integration of nomad with postgres
package pg

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type VersionStore struct {
	DB *sql.DB
}

func NewVersionStore(db *sql.DB) *VersionStore {
	return &VersionStore{DB: db}
}

func (vs *VersionStore) HasVersion(v string) bool {
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

func (vs *VersionStore) AddVersion(v string) {
	_, err := vs.DB.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", v)
	if err != nil {
		panic(err)
	}
}

// SetupVersionStore creates the schema_migrations table to store the versions
func (vs *VersionStore) SetupVersionStore() error {
	_, err := vs.DB.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
  version text NOT NULL UNIQUE
)`)
	return err
}
