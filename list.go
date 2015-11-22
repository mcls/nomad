package nomad

import (
	"fmt"
	"log"
	"sort"
)

// List is a list of migrations
type List struct {
	VersionStore
	migrations []*Migration
}

func NewList(versionStore VersionStore) *List {
	return &List{
		migrations:   []*Migration{},
		VersionStore: versionStore,
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

func (m *List) Run(context interface{}) error {
	if err := m.setup(); err != nil {
		return err
	}
	for _, x := range m.migrations {
		if m.HasVersion(x.Version) {
			continue
		}
		log.Printf("Running migration %q\n", x.Version)
		if x.Up == nil {
			return fmt.Errorf("No Up() function for migration %q", x.Version)
		}
		if err := x.Up(context); err == nil {
			m.AddVersion(x.Version)
		} else {
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
func (m *List) Rollback(context interface{}) error {
	if err := m.setup(); err != nil {
		return err
	}
	sort.Sort(sort.Reverse(m))
	for _, x := range m.migrations {
		if !m.HasVersion(x.Version) {
			continue
		}
		log.Printf("Rolling back migration %q\n", x.Version)
		if x.Down == nil {
			return fmt.Errorf("No Down() function for migration %q", x.Version)
		}
		if err := x.Down(context); err == nil {
			m.RemoveVersion(x.Version)
			return nil // Stop after one rollback
		} else {
			return err
		}
	}
	return nil
}
