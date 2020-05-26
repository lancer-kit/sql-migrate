package main

import (
	"flag"
	"fmt"
	"strings"

	migrate "github.com/lancer-kit/sql-migrate"
)

type RedoCommand struct {
}

func (c *RedoCommand) Help() string {
	helpText := `
Usage: sql-migrate redo [options] ...

  Reapply the last migration.

Options:

  -config=dbconfig.yml   Configuration file to use.
  -env="development"     Environment.
  -dryrun                Don't apply migrations, just print them.
  -enablePatch           Enable patch versions

`
	return strings.TrimSpace(helpText)
}

func (c *RedoCommand) Synopsis() string {
	return "Reapply the last migration"
}

func (c *RedoCommand) Run(args []string) int {
	var dryrun bool
	var enablePatch bool

	cmdFlags := flag.NewFlagSet("redo", flag.ContinueOnError)
	cmdFlags.Usage = func() { ui.Output(c.Help()) }
	cmdFlags.BoolVar(&dryrun, "dryrun", false, "Don't apply migrations, just print them.")
	cmdFlags.BoolVar(&enablePatch, "enablePatch", false, "Enable patch versions.")
	ConfigFlags(cmdFlags)

	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	env, err := GetEnvironment()
	if err != nil {
		ui.Error(fmt.Sprintf("Could not parse config: %s", err))
		return 1
	}

	db, dialect, err := GetConnection(env)
	if err != nil {
		ui.Error(err.Error())
		return 1
	}

	source := migrate.FileMigrationSource{
		Dir: env.Dir,
	}

	migrate.EnablePatchMode(enablePatch)

	var migrations []*migrate.PlannedMigration
	var migrationsPatch []*migrate.PlannedMigrationPatch
	if enablePatch {
		migrationsPatch, _, err = migrate.PlanMigrationPatch(db, dialect, source, migrate.Down, 1)
		if err != nil {
			ui.Error(fmt.Sprintf("Migration (redo) failed: %v", err))
			return 1
		} else if len(migrationsPatch) == 0 {
			ui.Output("Nothing to do!")
			return 0
		}
	} else {
		migrations, _, err = migrate.PlanMigration(db, dialect, source, migrate.Down, 1)
		if err != nil {
			ui.Error(fmt.Sprintf("Migration (redo) failed: %v", err))
			return 1
		} else if len(migrations) == 0 {
			ui.Output("Nothing to do!")
			return 0
		}
	}

	if dryrun {
		if migrationsPatch != nil {
			PrintMigrationPatch(migrationsPatch[0], migrate.Down)
			PrintMigrationPatch(migrationsPatch[0], migrate.Up)
		}
		if migrations != nil {
			PrintMigration(migrations[0], migrate.Down)
			PrintMigration(migrations[0], migrate.Up)
		}
	} else {
		_, err := migrate.ExecMax(db, dialect, source, migrate.Down, 1)
		if err != nil {
			ui.Error(fmt.Sprintf("Migration (down) failed: %s", err))
			return 1
		}

		_, err = migrate.ExecMax(db, dialect, source, migrate.Up, 1)
		if err != nil {
			ui.Error(fmt.Sprintf("Migration (up) failed: %s", err))
			return 1
		}

		if migrations != nil {
			ui.Output(fmt.Sprintf("Reapplied migration %s.", migrations[0].Id))
			return 0
		}

		if migrationsPatch != nil {
			ui.Output(fmt.Sprintf("Reapplied migration %s.", migrationsPatch[0].Name))
			return 0
		}
	}

	return 0
}
