package nomad

import (
	"fmt"
	"log"
	"sort"
)

// List is a list of migrations
type List struct {
	VersionStore
	Context    interface{}
	hooks      *Hooks
	migrations []*Migration
}

type Hooks struct {
	Before  func(interface{}) error        // Before is called before running or rolling back a migration
	After   func(interface{}) error        // After is called after running or rolling back a migration
	OnError func(interface{}, error) error // OnError is called if anything goes wrong during a migration
}

func NewList(versionStore VersionStore, context interface{}, hooks *Hooks) *List {
	if hooks == nil {
		hooks = &Hooks{}
	}
	return &List{
		Context:      context,
		VersionStore: versionStore,
		hooks:        hooks,
		migrations:   []*Migration{},
	}
}

func (m *List) Add(migration *Migration) {
	m.migrations = append(m.migrations, migration)
}

func (m *List) Get(i int) *Migration {
	return m.migrations[i]
}

func (m *List) Len() int {
	return len(m.migrations)
}

func (m *List) Less(i, j int) bool {
	a := m.migrations[i]
	b := m.migrations[j]
	return a.Version < b.Version
}

func (m *List) Swap(i, j int) {
	a := m.migrations[i]
	b := m.migrations[j]
	m.migrations[i] = b
	m.migrations[j] = a
}

func (m *List) Sort() {
	sort.Sort(m)
}

func (m *List) Run() error {
	if err := m.setup(); err != nil {
		return err
	}
	for _, x := range m.migrations {
		if m.HasVersion(x.Version) {
			continue
		}
		log.Printf("Running migration %q\n", x.Version)
		if err := m.runWithHooks(x, m.migrateUp); err != nil {
			return err
		}

	}
	return nil
}

// setup setup the version store and sorts the migrations according to their
// version
func (m *List) setup() error {
	if err := m.SetupVersionStore(); err == nil {
		m.Sort()
		return nil
	} else {
		return err
	}
}

// Rollback reverts the last migration
func (m *List) Rollback() error {
	if err := m.setup(); err != nil {
		return err
	}
	sort.Sort(sort.Reverse(m))
	for _, x := range m.migrations {
		if !m.HasVersion(x.Version) {
			continue
		}
		log.Printf("Rolling back migration %q\n", x.Version)
		// Stop after one rollback
		return m.runWithHooks(x, m.migrateDown)
	}
	return nil
}

func (m *List) runWithHooks(migration *Migration, fn func(*Migration) error) error {
	if fn == nil {
		return fmt.Errorf("No function for migration")
	}

	if m.hooks.Before != nil {
		if err := m.hooks.Before(m.Context); err != nil {
			return err
		}
	}

	if err := fn(migration); err != nil {
		if m.hooks.OnError != nil {
			if err2 := m.hooks.OnError(m.Context, err); err2 != nil {
				return err2
			}
		}
		return err
	}

	if m.hooks.After != nil {
		if err := m.hooks.After(m.Context); err != nil {
			return err
		}
	}

	return nil
}

func (m *List) migrateUp(migration *Migration) error {
	if migration.Up == nil {
		return fmt.Errorf("No Up() function for migration %q", migration.Version)
	}
	if err := migration.Up(m.Context); err != nil {
		return err
	}
	return m.AddVersion(migration.Version)
}

func (m *List) migrateDown(migration *Migration) error {
	if migration.Down == nil {
		return fmt.Errorf("No Down() function for migration %q", migration.Version)
	}

	if err := migration.Down(m.Context); err != nil {
		return err
	}

	return m.RemoveVersion(migration.Version)
}
