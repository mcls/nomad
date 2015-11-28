package inmem

import (
	"errors"
	"testing"

	"github.com/mcls/nomad"
)

func TestSort(t *testing.T) {
	l := nomad.NewList(NewMemVersionStore(), nil, nil)
	l.Add(&nomad.Migration{Version: "B"})
	l.Add(&nomad.Migration{Version: "A"})
	l.Add(&nomad.Migration{Version: "C"})
	l.Sort()

	for i, v := range []string{"A", "B", "C"} {
		x := l.Get(i).Version
		if x != v {
			t.Fatalf("Expected elem at %d to be '%s', but was '%s'", i, v, x)
		}
	}
}

func TestRun(t *testing.T) {
	x := 0

	l := nomad.NewList(NewMemVersionStore(), nil, nil)
	for _, v := range []string{"A", "B"} {
		if l.HasVersion(v) {
			t.Fatalf("Can't have version '%s'", v)
		}
	}

	l.Add(&nomad.Migration{
		Version: "A",
		Up: func(ctx interface{}) error {
			x += 1
			return nil
		},
	})
	l.Add(&nomad.Migration{
		Version: "B",
		Up: func(ctx interface{}) error {
			x += 2
			return nil
		},
	})

	if err := l.Run(); err != nil {
		t.Fatal(err)
	}

	if x != 3 {
		t.Fatalf("Didn't run migrations properly. x = %d\n", x)
	}

	// Check that all versions have been added
	for _, v := range []string{"A", "B"} {
		if !l.HasVersion(v) {
			t.Fatalf("Should have version '%s'", v)
		}
	}
}

func TestRunWithErrors(t *testing.T) {
	x := 0
	l := nomad.NewList(NewMemVersionStore(), nil, nil)
	l.Add(&nomad.Migration{
		Version: "A",
		Up: func(ctx interface{}) error {
			return errors.New("Oh no")
		},
	})
	l.Add(&nomad.Migration{
		Version: "B",
		Up: func(ctx interface{}) error {
			x += 1
			return nil
		},
	})
	err := l.Run()

	if err == nil || err.Error() != "Oh no" {
		t.Fatalf("Wrong error returned: '%s'", err)
	}

	if x != 0 {
		t.Fatal("Something went wrong while running the migrations")
	}
}

func TestDoesntRunMigrationTwice(t *testing.T) {
	x := 0

	l := nomad.NewList(NewMemVersionStore(), nil, nil)
	l.AddVersion("A")
	l.Add(&nomad.Migration{
		Version: "A",
		Up: func(ctx interface{}) error {
			x += 5
			return nil
		},
	})
	l.Add(&nomad.Migration{
		Version: "B",
		Up: func(ctx interface{}) error {
			x += 1
			return nil
		},
	})

	l.Run()

	if x != 1 {
		t.Fatal("Didn't run migrations properly")
	}

	// Check that all versions have been added
	for _, v := range []string{"A", "B"} {
		if !l.HasVersion(v) {
			t.Fatalf("Should have version '%s'", v)
		}
	}
}

func TestRollback(t *testing.T) {
	x := 0

	l := nomad.NewList(NewMemVersionStore(), nil, nil)
	l.AddVersion("A")
	l.AddVersion("B")

	l.Add(&nomad.Migration{
		Version: "A",
		Down: func(ctx interface{}) error {
			x = 50
			return nil
		},
	})
	l.Add(&nomad.Migration{
		Version: "B",
		Down: func(ctx interface{}) error {
			x = 100
			return nil
		},
	})

	if err := l.Rollback(); err != nil {
		t.Fatal(err)
	}

	if x != 100 {
		t.Fatalf("Didn't rollback properly. x = %d\n", x)
	}

	// Check that only the last migration was rolled back
	if !l.HasVersion("A") {
		t.Fatal("Should not have rolled back 'A'")
	}

	if l.HasVersion("B") {
		t.Fatal("Still has version 'B'")
	}
}

func TestRollbackWithErrors(t *testing.T) {
	l := nomad.NewList(NewMemVersionStore(), nil, nil)
	l.AddVersion("A")

	l.Add(&nomad.Migration{
		Version: "A",
		Down: func(ctx interface{}) error {
			return errors.New("No way back!")
		},
	})

	if err := l.Rollback(); err == nil {
		t.Fatal("Expected error")
	} else if err.Error() != "No way back!" {
		t.Fatalf("Expected different error than %q", err)
	}

	if !l.HasVersion("A") {
		t.Fatal("Shouldn't have removed 'A' after error")
	}
}