package main

import (
	"flag"
	"strings"

	migrate "github.com/rubenv/sql-migrate"
)

type DownCommand struct {
}

func (c *DownCommand) Help() string {
	helpText := `
Usage: evedbtool down [options] ...
  Undo a database migration.
Options:
  -limit=1               Limit the number of migrations (0 = unlimited).
  -dryrun                Don't apply migrations, just print them.
`
	return strings.TrimSpace(helpText)
}

func (c *DownCommand) Synopsis() string {
	return "Undo a database migration"
}

func (c *DownCommand) Run(args []string) int {
	var limit int
	var dryrun bool

	migrate.SetTable("migrations")

	cmdFlags := flag.NewFlagSet("down", flag.ContinueOnError)
	cmdFlags.Usage = func() { ui.Output(c.Help()) }
	cmdFlags.IntVar(&limit, "limit", 1, "Max number of migrations to apply.")
	cmdFlags.BoolVar(&dryrun, "dryrun", false, "Don't apply migrations, just print them.")

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	err := ApplyMigrations(migrate.Down, dryrun, limit)
	if err != nil {
		ui.Error(err.Error())
		return 1
	}

	return 0
}
