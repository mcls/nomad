package nomad

import (
	"sort"
)

type Migration struct {
	Version string                      // Unique version
	Up      func(ctx interface{}) error // Ran when migrating
	Down    func(ctx interface{}) error // Ran when rolling back
}

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
	err := m.SetupVersions()
	if err != nil {
		return err
	}
	m.Sort()
	for _, x := range m.migrations {
		if m.HasVersion(x.Version) {
			continue
		}
		if err := x.Up(context); err != nil {
			return err
		} else {
			m.AddVersion(x.Version)
		}
	}
	return nil
}
