package nomad

// VersionStore checks whether versions are up to date
type VersionStore interface {
	AddVersion(v string) error
	RemoveVersion(v string) error
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
