package nomad

import (
	"log"

	"github.com/spf13/cobra"
)

func NewMigrationCmd(runner *Runner, migrationDirectory string) *cobra.Command {
	cmdRoot := &cobra.Command{
		Use:   "migration",
		Short: "migration subcommands",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}

	cmdNew := &cobra.Command{
		Use:   "new",
		Short: "create migration",
		Run: func(cmd *cobra.Command, args []string) {
			cg := NewCodeGenerator(migrationDirectory)
			if err := cg.Create(args[0]); err != nil {
				log.Fatal(err)
			}
		},
	}

	cmdRun := &cobra.Command{
		Use:   "run",
		Short: "run all pending migrations",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runner.Run(); err != nil {
				log.Fatal(err)
			}
		},
	}

	cmdRollback := &cobra.Command{
		Use:   "rollback",
		Short: "rollback the most recent migration",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runner.Rollback(); err != nil {
				log.Fatal(err)
			}
		},
	}

	cmdRoot.AddCommand(cmdNew)
	cmdRoot.AddCommand(cmdRun)
	cmdRoot.AddCommand(cmdRollback)

	return cmdRoot
}
