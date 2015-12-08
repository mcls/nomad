package inmem

import "github.com/mcls/nomad"

func NewRunner(list *nomad.List, ctx ...interface{}) *nomad.Runner {
	runner := nomad.NewRunner(NewMemVersionStore(), list, nil)
	if len(ctx) > 0 {
		runner.Context = ctx[0]
	}
	return runner
}

// MemVersionStore is a in-memory implementation of VersionStore,
// only used for tests
type MemVersionStore struct {
	versions map[string]bool
}

func NewMemVersionStore() *MemVersionStore {
	return &MemVersionStore{map[string]bool{}}
}

// AddVersion adds the version
func (mv *MemVersionStore) AddVersion(v string) error {
	mv.versions[v] = true
	return nil
}

func (mv *MemVersionStore) RemoveVersion(v string) error {
	mv.versions[v] = false
	return nil
}

// HasVersion checks if the version already exists
func (mv *MemVersionStore) HasVersion(v string) bool {
	return mv.versions[v]
}

// SetupVersionStore must be ran before checking versions
func (mv *MemVersionStore) SetupVersionStore() error {
	if mv.versions == nil {
		mv.versions = map[string]bool{}
	}
	return nil
}
