package main

import (
	"flag"
	"fmt"
	"strings"

	migrate "github.com/rubenv/sql-migrate"
	"github.com/spf13/viper"
)

type SkipCommand struct {
}

func (c *SkipCommand) Help() string {
	helpText := `
Usage: evedbtool skip [options] ...
  Set the database level to the most recent version available, without actually running the migrations.
Options:
  -limit=0               Limit the number of migrations (0 = unlimited).
`
	return strings.TrimSpace(helpText)
}

func (c *SkipCommand) Synopsis() string {
	return "Sets the database level to the most recent version available, without running the migrations"
}

func (c *SkipCommand) Run(args []string) int {
	var limit int
	var dryrun bool

	migrate.SetTable("migrations")

	cmdFlags := flag.NewFlagSet("up", flag.ContinueOnError)
	cmdFlags.Usage = func() { ui.Output(c.Help()) }
	cmdFlags.IntVar(&limit, "limit", 0, "Max number of migrations to skip.")

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	err := SkipMigrations(migrate.Up, dryrun, limit)
	if err != nil {
		ui.Error(err.Error())
		return 1
	}

	return 0
}

func SkipMigrations(dir migrate.MigrationDirection, dryrun bool, limit int) error {

	db := getDB()
	dialect := "mysql"

	source := migrate.FileMigrationSource{
		Dir: viper.GetString("migrations-dir"),
	}

	n, err := migrate.SkipMax(db, dialect, source, dir, limit)
	if err != nil {
		return fmt.Errorf("Migration failed: %s", err)
	}

	switch n {
	case 0:
		ui.Output("All migrations have already been applied")
	case 1:
		ui.Output("Skipped 1 migration")
	default:
		ui.Output(fmt.Sprintf("Skipped %d migrations", n))
	}

	return nil
}
