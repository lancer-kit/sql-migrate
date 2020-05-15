package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/rubenv/sql-migrate"
)

type SkipCommand struct {
}

func (c *SkipCommand) Help() string {
	helpText := `
Usage: sql-migrate skip [options] ...

  Set the database level to the most recent version available, without actually running the migrations.

Options:

  -config=dbconfig.yml   Configuration file to use.
  -env="development"     Environment.
  -limit=0               Limit the number of migrations (0 = unlimited).
  -enablePatch           Enable patch versions

`
	return strings.TrimSpace(helpText)
}

func (c *SkipCommand) Synopsis() string {
	return "Sets the database level to the most recent version available, without running the migrations"
}

func (c *SkipCommand) Run(args []string) int {
	var limit int
	var dryrun bool
	var enablePatch bool

	cmdFlags := flag.NewFlagSet("up", flag.ContinueOnError)
	cmdFlags.Usage = func() { ui.Output(c.Help()) }
	cmdFlags.IntVar(&limit, "limit", 0, "Max number of migrations to skip.")
	cmdFlags.BoolVar(&enablePatch, "enablePatch", false, "Enable patch versions.")
	ConfigFlags(cmdFlags)

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	migrate.EnablePatchMode(enablePatch)

	err := SkipMigrations(migrate.Up, dryrun, enablePatch, limit)
	if err != nil {
		ui.Error(err.Error())
		return 1
	}

	return 0
}

func SkipMigrations(dir migrate.MigrationDirection, dryrun, enablePatch bool, limit int) error {
	env, err := GetEnvironment()
	if err != nil {
		return fmt.Errorf("Could not parse config: %s", err)
	}

	db, dialect, err := GetConnection(env)
	if err != nil {
		return err
	}

	source := migrate.FileMigrationSource{
		Dir: env.Dir,
	}

	var n int
	if enablePatch {
		n, err = migrate.SkipMaxPatch(db, dialect, source, dir, limit)
	} else {
		n, err = migrate.SkipMax(db, dialect, source, dir, limit)
	}
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
