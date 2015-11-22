package nomad

// VersionStore checks whether versions are up to date
type VersionStore interface {
	AddVersion(v string)
	RemoveVersion(v string)
	HasVersion(v string) bool
	SetupVersionStore() error
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
func (mv *MemVersionStore) AddVersion(v string) {
	mv.versions[v] = true
}

func (mv *MemVersionStore) RemoveVersion(v string) {
	mv.versions[v] = false
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
