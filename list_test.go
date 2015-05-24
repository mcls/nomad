package nomad

import (
	"errors"
	"testing"
)

func TestSort(t *testing.T) {
	l := NewList(NewMemVersionStore())
	l.Add(&Migration{Version: "B"})
	l.Add(&Migration{Version: "A"})
	l.Add(&Migration{Version: "C"})
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

	l := NewList(NewMemVersionStore())
	for _, v := range []string{"A", "B"} {
		if l.HasVersion(v) {
			t.Fatalf("Can't have version '%s'", v)
		}
	}

	l.Add(&Migration{
		Version: "A",
		Up: func(ctx interface{}) error {
			x += 1
			return nil
		},
	})
	l.Add(&Migration{
		Version: "B",
		Up: func(ctx interface{}) error {
			x += 2
			return nil
		},
	})

	l.Run(nil)

	if x != 3 {
		t.Fatal("Didn't run migrations properly")
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
	l := NewList(NewMemVersionStore())
	l.Add(&Migration{
		Version: "A",
		Up: func(ctx interface{}) error {
			return errors.New("Oh no")
		},
	})
	l.Add(&Migration{
		Version: "B",
		Up: func(ctx interface{}) error {
			x += 1
			return nil
		},
	})
	err := l.Run(nil)

	if x != 0 {
		t.Fatal("Didn't run migrations properly")
	}

	if err.Error() != "Oh no" {
		t.Fatalf("Wrong error returned: '%s'", err)
	}
}

func TestDoesntRunMigrationTwice(t *testing.T) {
	x := 0

	l := NewList(NewMemVersionStore())
	l.AddVersion("A")
	l.Add(&Migration{
		Version: "A",
		Up: func(ctx interface{}) error {
			x += 5
			return nil
		},
	})
	l.Add(&Migration{
		Version: "B",
		Up: func(ctx interface{}) error {
			x += 1
			return nil
		},
	})

	l.Run(nil)
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