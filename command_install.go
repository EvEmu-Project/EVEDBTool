package main

import (
	"flag"
	"strings"

	migrate "github.com/rubenv/sql-migrate"
)

type InstallCommand struct {
}

func (c *InstallCommand) Help() string {
	helpText := `
Usage: evedbtool install [options] ...
  Installs the base database and migrates to the most recent version available.
Options:
  -limit=0               Limit the number of migrations (0 = unlimited).
  -dryrun                Don't apply migrations, just print them.
`
	return strings.TrimSpace(helpText)
}

func (c *InstallCommand) Synopsis() string {
	return "Installs the base database and migrates to the most recent version available."
}

func (c *InstallCommand) Run(args []string) int {
	var limit int
	var dryrun bool

	cmdFlags := flag.NewFlagSet("up", flag.ContinueOnError)
	cmdFlags.Usage = func() { ui.Output(c.Help()) }
	cmdFlags.IntVar(&limit, "limit", 0, "Max number of migrations to apply.")
	cmdFlags.BoolVar(&dryrun, "dryrun", false, "Don't apply migrations, just print them.")

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	if !dryrun {
		tables := GetNumberOfTables()
		log.Info("Number of tables in DB: ", tables)
		if tables == 0 {
			log.Info("Database not initialized, installing...")
			InstallBase()
		} else {
			log.Info("Base database already installed. Won't overwrite.")
		}
	} else {
		log.Info("Dry run, not installing base.")
	}
	migrate.SetTable("migrations")
	err := ApplyMigrations(migrate.Up, dryrun, limit)
	if err != nil {
		ui.Error(err.Error())
		return 1
	}

	return 0
}
