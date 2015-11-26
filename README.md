# Nomad

***UNDER DEVELOPMENT***

Nomad is a migration library for Go. Its goal is to be minimal and flexible.  
You can use it with an ORM, or use plain SQL.

## Example

Here's an example of a migration definition.

```go
&nomad.Migration{
    Version: "2015-11-22_18:07:05",
    Up: func(ctx interface{}) error {
        c := ctx.(*Context)
        _, err := c.DB.Exec(`
        CREATE TABLE posts (
            id serial PRIMARY KEY,
            title text NOT NULL CHECK(length(title) < 200),
            content text NOT NULL,
            created_at timestamp with time zone DEFAULT(current_timestamp)
        )`)
        return err
    },
    Down: func(ctx interface{}) error {
        c := ctx.(*Context)
        _, err := c.DB.Exec("DROP TABLE posts")
        return err
    },
}
```

For a more complete example, take a look at
[example_test.go](https://github.com/mcls/nomad/blob/master/example_test.go).


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

