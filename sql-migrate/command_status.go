package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"

	"github.com/rubenv/sql-migrate"
)

type StatusCommand struct {
}

func (c *StatusCommand) Help() string {
	helpText := `
Usage: sql-migrate status [options] ...

  Show migration status.

Options:

  -config=dbconfig.yml   Configuration file to use.
  -env="development"     Environment.

`
	return strings.TrimSpace(helpText)
}

func (c *StatusCommand) Synopsis() string {
	return "Show migration status"
}

func (c *StatusCommand) Run(args []string) int {
	cmdFlags := flag.NewFlagSet("status", flag.ContinueOnError)
	cmdFlags.Usage = func() { ui.Output(c.Help()) }
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
	migrations, err := source.FindMigrations()
	if err != nil {
		ui.Error(err.Error())
		return 1
	}

	records, err := migrate.GetMigrationRecords(db, dialect)
	if err != nil {
		ui.Error(err.Error())
		return 1
	}

	var existingMigrations []*statusRow
	for _, migrationRecord := range records {
		em := &migrate.Migration{
			Name:  migrationRecord.Name,
			Ver:   migrationRecord.Ver,
			Patch: migrationRecord.Patch,
		}

		if err := em.ParseName(); err != nil {
			ui.Error(err.Error())
			return 1
		}
		existingMigrations = append(existingMigrations, &statusRow{
			Migration: em,
			AppliedAt: migrationRecord.CreatedAt,
		})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Migration", "Applied"})
	table.SetColWidth(60)

	for _, m := range migrations {
		var existMigration *statusRow
		for _, existing := range existingMigrations {
			if existing.VerInt == m.VerInt && existing.PatchInt >= m.PatchInt {
				existMigration = existing
				break
			}
		}

		if existMigration != nil {
			table.Append([]string{
				m.Name,
				existMigration.AppliedAt.String(),
			})
		} else {
			table.Append([]string{
				m.Name,
				"no",
			})
		}
	}

	table.Render()

	return 0
}

type statusRow struct {
	*migrate.Migration
	AppliedAt time.Time
}
