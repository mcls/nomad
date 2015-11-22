package nomad

type Migration struct {
	Version string                      // Unique version
	Up      func(ctx interface{}) error // Ran when migrating
	Down    func(ctx interface{}) error // Ran when rolling back
}
