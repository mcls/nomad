package nomad

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
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

func TestCodeGenerator_OutputsValidSyntax(t *testing.T) {
	cg := NewCodeGenerator(migrationDir)
	w := new(bytes.Buffer)
	cg.WriteMigration(w, "abc")
	// Use go/format.Source to detect syntax errors
	_, err := format.Source(w.Bytes())
	if err != nil {
		t.Log("Couldn't build migrations")
		t.Fatal(err)
	}
}

func TestGeneratedMigrationsCanBeBuilt(t *testing.T) {
	clearMigrationDir(t, migrationDir)
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
		t.Log("Couldn't build migrations")
		t.Log(errout.String())
		t.Log(out.String())
		t.Fatal(err)
	}
}

func clearMigrationDir(t *testing.T, dir string) {
	err := os.RemoveAll(migrationDir)
	if err != nil {
		t.Log("Couldn't clean migrations dir")
		t.Log(err)
	}
}

func TestCodeGenerator_DoesntOverrideExistingSetupFile(t *testing.T) {
	var err error
	clearMigrationDir(t, migrationDir)
	cg := NewCodeGenerator(migrationDir)
	cg.NomadPackage = "github.com/mcls/nomad"
	cg.Create("create_users")

	setupPath := path.Join(cg.Dir, "000_setup_migrations.go")

	err = ioutil.WriteFile(setupPath, []byte("// Random Comment"), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	original, err := ioutil.ReadFile(setupPath)
	if err != nil {
		t.Fatal(err)
	}

	cg.Create("create_posts")

	afterMigrate, err := ioutil.ReadFile(setupPath)
	if err != nil {
		t.Fatal(err)
	}

	if string(original) != string(afterMigrate) {
		t.Fatal("Setup file contents shouldn't have changed")
	}
}
