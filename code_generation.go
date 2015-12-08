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
	nomadpg "{{.NomadPackage}}/pg"
	// Setup postgres driver
	_ "github.com/lib/pq"
)

var Migrations *nomad.List

func init() {
	Migrations = nomad.NewList()
}

func Runner() *nomad.Runner {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	return nomadpg.NewRunner(db, Migrations)
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

// CodeGenerator generates migration files
type CodeGenerator struct {
	Dir          string        // Where migrations will be stored
	NewVersion   func() string // Generates the Migration's version
	NomadPackage string
}

func NewCodeGenerator(dir string) *CodeGenerator {
	return &CodeGenerator{
		Dir:          dir,
		NewVersion:   generateTimestamp,
		NomadPackage: "github.com/mcls/nomad",
	}
}

// Create creates a new migration
func (cg *CodeGenerator) Create(name string) error {
	err := os.MkdirAll(cg.Dir, 0755)
	if err != nil {
		return err
	}

	err = cg.createSetupFile()
	if err != nil {
		return err
	}

	version := cg.NewVersion()
	f, err := cg.createFile(name, version)
	if err != nil {
		return err
	}
	defer f.Close()
	return cg.WriteMigration(f, version)
}

func (cg *CodeGenerator) createSetupFile() error {
	// Use 000_ prefix so it's init function gets called first and can do setup
	// on which the other migrations can rely
	full := path.Join(cg.Dir, "000_setup_migrations.go")
	f, err := os.Create(full)
	if err != nil {
		return err
	}
	t := template.Must(template.New("default").Parse(tplSetup))
	err = t.Execute(f, map[string]string{
		"NomadPackage": cg.NomadPackage,
	})
	return err
}

func (cg *CodeGenerator) createFile(name, version string) (*os.File, error) {
	name = fmt.Sprintf("%s_%s.go", version, name)
	full := path.Join(cg.Dir, name)
	fmt.Printf("Creating migration: '%s'\n", full)
	return os.Create(full)
}

// WriteMigration writes boilerplate migration go code to the writer
func (cg *CodeGenerator) WriteMigration(w io.Writer, version string) error {
	t := template.Must(template.New("migration").Parse(tplMigration))
	err := t.Execute(w, map[string]string{
		"NomadPackage": cg.NomadPackage,
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
