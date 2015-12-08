# Nomad

[![Build
Status](https://travis-ci.org/mcls/nomad.svg)](https://travis-ci.org/mcls/nomad)
| [GoDoc](https://godoc.org/github.com/mcls/nomad)

***UNDER DEVELOPMENT***

Nomad is a migration library for Go. Its goal is to be minimal and flexible.  
You can use it with an ORM, or use plain SQL.

## Example

Here's an example of a migration definition.

```go
import (
    "database/sql"
    "log"

    "github.com/mcls/nomad"
    nomadpg "github.com/mcls/nomad/pg"
)

migrations := nomad.NewList()

m1 := &nomad.Migration{
  Version: "2015-11-26_19:00:00",
  Up: func(ctx interface{}) error {
    c := ctx.(*nomadpg.Context)
    _, err := c.Tx.Exec(`
      CREATE TABLE users (
        id serial PRIMARY KEY,
        username text
      );`)
    return err
  },
  Down: func(ctx interface{}) error {
    c := ctx.(*nomadpg.Context)
    _, err := c.Tx.Exec(`DROP TABLE users`)
    return err
  },
}
m2 := &nomad.Migration{
  Version: "2015-11-26_19:30:00",
  Up: func(ctx interface{}) error {
    c := ctx.(*nomadpg.Context)
    _, err := c.Tx.Exec(`
      CREATE TABLE posts (
        id serial PRIMARY KEY,
        title text
        content text
      );`)
    return err
  },
  Down: func(ctx interface{}) error {
    c := ctx.(*nomadpg.Context)
    _, err := c.Tx.Exec(`DROP TABLE posts`)
    return err
  },
}
migrations.Add(m1)
migrations.Add(m2)

db, err := sql.Open("postgres", "dbname=nomad_db_test sslmode=disable")
if err != nil {
  log.Fatal(err)
}

runner := nomadpg.NewRunner(db, migrations)

// Run all pending migrations
runner.Run()

// Rollback the latest migration
runner.Rollback()
```

For more examples, take a look at:

* In-memory example: [example_test.go](https://github.com/mcls/nomad/blob/master/inmem/example_test.go).
* PostgreSQL example: [example_test.go](https://github.com/mcls/nomad/blob/master/pg/example_test.go).

## Install

To install the package and the command-line tool:

```bash
git clone https://github.com/mcls/nomad.git
make install
```

## Code generation

The initial migration can be created via:

```bash
# to create a migration
nomad new create_users
```

After that you can import [cobra](https://github.com/spf13/cobra) commands in
your own app:

```go
migrationCmd := nomad.NewMigrationCmd(
  migrations, // migration.List object, basically your migrations
  "./migrations", // the migrations directory
)
```

This creates a `migration` command with these subcommands:

* `run`: runs all pending migrations
* `rollback`: rolls back the latest migration
* `new`: creates a new migration

