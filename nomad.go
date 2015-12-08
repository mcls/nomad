package nomad

import (
	"fmt"
	"log"
	"sort"
)

type Migration struct {
	Version string                      // Unique version
	Up      func(ctx interface{}) error // Ran when migrating
	Down    func(ctx interface{}) error // Ran when rolling back
}

// List is a list of migrations
type List struct {
	migrations []*Migration
}

func NewList() *List {
	return &List{[]*Migration{}}
}

type Runner struct {
	VersionStore
	hooks   *Hooks
	Context interface{}
	List    *List
}

// VersionStore checks whether versions are up to date
type VersionStore interface {
	AddVersion(v string) error
	RemoveVersion(v string) error
	HasVersion(v string) bool
	SetupVersionStore() error
}

type Hooks struct {
	Before  func(interface{}) error        // Before is called before running or rolling back a migration
	After   func(interface{}) error        // After is called after running or rolling back a migration
	OnError func(interface{}, error) error // OnError is called if anything goes wrong during a migration
}

func NewRunner(versionStore VersionStore, hooks *Hooks, list *List, context interface{}) *Runner {
	if hooks == nil {
		hooks = &Hooks{}
	}
	runner := &Runner{
		VersionStore: versionStore,
		hooks:        hooks,
		Context:      context,
		List:         list,
	}
	return runner
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

func (r *Runner) Run() error {
	if err := r.setup(); err != nil {
		return err
	}
	for _, x := range r.List.migrations {
		if r.HasVersion(x.Version) {
			continue
		}
		log.Printf("Running migration %q\n", x.Version)
		if err := r.runWithHooks(x, r.migrateUp); err != nil {
			return err
		}

	}
	return nil
}

// setup setup the version store and sorts the migrations according to their
// version
func (r *Runner) setup() error {
	if err := r.SetupVersionStore(); err == nil {
		return nil
	} else {
		return err
	}
}

// Rollback reverts the last migration
func (r *Runner) Rollback() error {
	if err := r.setup(); err != nil {
		return err
	}
	sort.Sort(sort.Reverse(r.List))
	for _, x := range r.List.migrations {
		if !r.HasVersion(x.Version) {
			continue
		}
		log.Printf("Rolling back migration %q\n", x.Version)
		// Stop after one rollback
		return r.runWithHooks(x, r.migrateDown)
	}
	return nil
}

func (r *Runner) runWithHooks(migration *Migration, fn func(*Migration) error) error {
	if fn == nil {
		return fmt.Errorf("No function for migration")
	}

	if r.hooks.Before != nil {
		if err := r.hooks.Before(r.Context); err != nil {
			return err
		}
	}

	if err := fn(migration); err != nil {
		if r.hooks.OnError != nil {
			if err2 := r.hooks.OnError(r.Context, err); err2 != nil {
				return err2
			}
		}
		return err
	}

	if r.hooks.After != nil {
		if err := r.hooks.After(r.Context); err != nil {
			return err
		}
	}

	return nil
}

func (r *Runner) migrateUp(migration *Migration) error {
	if migration.Up == nil {
		return fmt.Errorf("No Up() function for migration %q", migration.Version)
	}
	if err := migration.Up(r.Context); err != nil {
		return err
	}
	return r.AddVersion(migration.Version)
}

func (r *Runner) migrateDown(migration *Migration) error {
	if migration.Down == nil {
		return fmt.Errorf("No Down() function for migration %q", migration.Version)
	}

	if err := migration.Down(r.Context); err != nil {
		return err
	}

	return r.RemoveVersion(migration.Version)
}
