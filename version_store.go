package nomad

// VersionStore checks whether versions are up to date
type VersionStore interface {
	AddVersion(v string)
	HasVersion(v string) bool
	SetupVersions() error
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

// HasVersion checks if the version already exists
func (mv *MemVersionStore) HasVersion(v string) bool {
	return mv.versions[v]
}

// SetupVersions must be ran before checking versions
func (mv *MemVersionStore) SetupVersions() error {
	if mv.versions == nil {
		mv.versions = map[string]bool{}
	}
	return nil
}
