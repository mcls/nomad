package inmem_test

import (
	"fmt"

	"github.com/mcls/nomad"
	"github.com/mcls/nomad/inmem"
)

// Context will be available to each migration and should be used to provide
// access to the database. This example just updates an in-memory array of
// Strings.
type Context struct {
	Data []string
}

// Below is an example with an in-memory VersionStore. The VersionStore
// determines which migrations are pending.
// The context object can be of any type and provides the migration functions
// with access to the database or other resources.
func Example() {
	migrations := nomad.NewList()
	m1 := &nomad.Migration{
		Version: "2015-11-26_19:00:00",
		Up: func(ctx interface{}) error {
			c := ctx.(*Context)
			fmt.Println("Migrated Up: m1")
			c.Data = append(c.Data, "m1")
			return nil
		},
		Down: func(ctx interface{}) error {
			c := ctx.(*Context)
			fmt.Println("Migrated Down: m2")
			c.Data = c.Data[:len(c.Data)-1]
			return nil
		},
	}
	m2 := &nomad.Migration{
		Version: "2015-11-26_19:30:00",
		Up: func(ctx interface{}) error {
			c := ctx.(*Context)
			fmt.Println("Migrated Up: m2")
			c.Data = append(c.Data, "m2")
			return nil
		},
		Down: func(ctx interface{}) error {
			c := ctx.(*Context)
			fmt.Println("Migrated Down: m2")
			c.Data = c.Data[:len(c.Data)-1]
			return nil
		},
	}
	migrations.Add(m1)
	migrations.Add(m2)

	context := &Context{[]string{}}
	runner := inmem.NewRunner(migrations, context)
	runner.Run()
	fmt.Printf("context.Data: %q\n", context.Data)
	runner.Rollback()
	fmt.Printf("context.Data: %q\n", context.Data)
	// Output:
	// Migrated Up: m1
	// Migrated Up: m2
	// context.Data: ["m1" "m2"]
	// Migrated Down: m2
	// context.Data: ["m1"]
}
