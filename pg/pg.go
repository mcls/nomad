// Helps with the integration of nomad with postgres
package pg

import (
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/mcls/nomad"
)

// NewRunner creates a nomad.Runner for postgres migrations
func NewRunner(db *sql.DB, list *nomad.List) *nomad.Runner {
	return nomad.NewRunner(
		NewVersionStore(db),
		list,
		NewContext(db),
		NewHooks(),
	)
}

type Context struct {
	DB *sql.DB
	Tx *sql.Tx
}

func NewContext(db *sql.DB) *Context {
	return &Context{
		DB: db,
		Tx: nil,
	}
}

func beforeHook(ctx interface{}) error {
	c := ctx.(*Context)
	if tx, err := c.DB.Begin(); err == nil {
		c.Tx = tx
		return nil
	} else {
		return err
	}
}
func afterHook(ctx interface{}) error {
	c := ctx.(*Context)
	if err := c.Tx.Commit(); err != nil {
		c.Tx = nil
		return err
	}
	return nil
}

func onError(ctx interface{}, origErr error) error {
	c := ctx.(*Context)
	if err := c.Tx.Rollback(); err != nil {
		return err
	}
	return origErr
}

func NewHooks() *nomad.Hooks {
	return &nomad.Hooks{
		Before:  beforeHook,
		After:   afterHook,
		OnError: onError,
	}
}

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

func (vs *VersionStore) AddVersion(v string) error {
	_, err := vs.DB.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", v)
	return err
}

func (vs *VersionStore) RemoveVersion(v string) error {
	_, err := vs.DB.Exec("DELETE FROM schema_migrations WHERE version = $1", v)
	return err
}

// SetupVersionStore creates the schema_migrations table to store the versions
func (vs *VersionStore) SetupVersionStore() error {
	_, err := vs.DB.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
  version text NOT NULL UNIQUE
)`)
	return err
}
