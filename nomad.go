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

// List is a list of migrations
type List struct {
	migrations []*Migration
}

func NewList() *List {
	return &List{[]*Migration{}}
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

// Runner runs pending migrations, or rolls back existing ones
type Runner struct {
	VersionStore
	Context interface{}
	list    *List
	hooks   *Hooks
}

func NewRunner(versionStore VersionStore, list *List, context interface{}, hooks ...*Hooks) *Runner {
	runner := &Runner{
		VersionStore: versionStore,
		Context:      context,
		list:         list,
		hooks:        &Hooks{},
	}
	if len(hooks) > 0 {
		runner.hooks = hooks[0]
	}
	return runner
}

func (r *Runner) Run() error {
	if err := r.setup(); err != nil {
		return err
	}
	for _, x := range r.list.migrations {
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
		r.list.Sort()
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
	sort.Sort(sort.Reverse(r.list))
	for _, x := range r.list.migrations {
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
