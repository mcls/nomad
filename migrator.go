package nomad

import (
	"fmt"
	"io"
	"os"
	"path"
	"text/template"
	"time"
)

var tplSetup string = `package migrations

import (
	"database/sql"
	"log"
	"os"

	"{{.NomadPackage}}"
	"{{.NomadPackage}}/pg"
	// Setup postgres driver
	_ "github.com/lib/pq"
)

var Migrations *nomad.List

func init() {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	Migrations = pg.NewList(db)
}
`

var tplMigration string = `package migrations

import (
	"fmt"

	"{{.NomadPackage}}"
	"{{.NomadPackage}}/pg"
)

func init() {
	migration := &nomad.Migration{
		Version: "{{.Version}}",
		Up: func(ctx interface{}) error {
			c := ctx.(*pg.Context)
			fmt.Println("Up")
			fmt.Println(c)
			_, err := c.Tx.Exec("CREATE TABLE ...")
			return err
		},
		Down: func(ctx interface{}) error {
			c := ctx.(*pg.Context)
			fmt.Println("Down")
			fmt.Println(c)
			return nil
		},
	}
	Migrations.Add(migration)
}
`

// Migrator generates migration files
type Migrator struct {
	Dir          string        // Where migrations will be stored
	NewVersion   func() string // Generates the Migration's version
	NomadPackage string
}

func NewMigrator(dir string) *Migrator {
	return &Migrator{
		Dir:          dir,
		NewVersion:   generateTimestamp,
		NomadPackage: "github.com/mcls/nomad",
	}
}

// Create creates a new migration
func (m *Migrator) Create(name string) error {
	err := os.MkdirAll(m.Dir, 0755)
	if err != nil {
		return err
	}

	err = m.createSetupFile()
	if err != nil {
		return err
	}

	version := m.NewVersion()
	f, err := m.createFile(name, version)
	if err != nil {
		return err
	}
	defer f.Close()
	return m.WriteMigration(f, version)
}

func (m *Migrator) createSetupFile() error {
	// Use 000_ prefix so it's init function gets called first and can do setup
	// on which the other migrations can rely
	full := path.Join(m.Dir, "000_setup_migrations.go")
	f, err := os.Create(full)
	if err != nil {
		return err
	}
	t := template.Must(template.New("default").Parse(tplSetup))
	err = t.Execute(f, map[string]string{
		"NomadPackage": m.NomadPackage,
	})
	return err
}

func (m *Migrator) createFile(name, version string) (*os.File, error) {
	name = fmt.Sprintf("%s_%s.go", version, name)
	full := path.Join(m.Dir, name)
	fmt.Printf("Creating migration: '%s'\n", full)
	return os.Create(full)
}

// WriteMigration writes boilerplate migration go code to the writer
func (m *Migrator) WriteMigration(w io.Writer, version string) error {
	t := template.Must(template.New("migration").Parse(tplMigration))
	err := t.Execute(w, map[string]string{
		"NomadPackage": m.NomadPackage,
		"Version":      version,
	})
	return err
}

func generateTimestamp() string {
	now := time.Now()
	return fmt.Sprintf(
		"%d-%02d-%02d_%02d:%02d:%02d",
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute(),
		now.Second(),
	)
}