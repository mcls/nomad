package nomad

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
)

const migrationDir = "dummy_migrations"

func TestCodeGenerator_NewVersion(t *testing.T) {
	cg := NewCodeGenerator(migrationDir)
	v := cg.NewVersion()
	re := regexp.MustCompile("^\\d{4}-\\d{2}-\\d{2}_\\d{2}:\\d{2}:\\d{2}")
	if !re.MatchString(v) {
		t.Fatal(fmt.Sprintf("Version '%s' didn't match regexp %s", v, re.String()))
	}
}

func TestMigrationFileHasValidSyntax(t *testing.T) {
	cg := NewCodeGenerator(migrationDir)
	w := new(bytes.Buffer)
	cg.WriteMigration(w, "abc")
	// Use go/format.Source to detect syntax errors
	_, err := format.Source(w.Bytes())
	if err != nil {
		fmt.Println("Couldn't build migrations")
		t.Fatal(err)
	}
}

func TestGeneratedMigrationsCanBeBuilt(t *testing.T) {
	err := os.RemoveAll(migrationDir)
	if err != nil {
		fmt.Println("Couldn't clean migrations dir")
		fmt.Println(err)
	}
	cg := NewCodeGenerator(migrationDir)
	cg.NomadPackage = "github.com/mcls/nomad"
	cg.Create("blah_blah")

	cmd := exec.Command("go", "build")
	full, err := filepath.Abs(migrationDir)
	if err != nil {
		t.Fatal(err)
	}
	cmd.Dir = full

	var out bytes.Buffer
	cmd.Stdout = &out
	var errout bytes.Buffer
	cmd.Stderr = &errout

	err = cmd.Run()
	if err != nil {
		fmt.Println("Couldn't build migrations")
		fmt.Println(errout.String())
		t.Fatal(err)
	}

	fmt.Println(out.String())
}
