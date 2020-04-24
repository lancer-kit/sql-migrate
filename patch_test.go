package migrate

import (
	"net/http"

	"github.com/gobuffalo/packr/v2"
	_ "github.com/mattn/go-sqlite3"
	. "gopkg.in/check.v1"
)

var sqliteMigrationsPatch = []*MigrationPatch{{
	Name: "0123_00_test.sql",
	Up:   []string{"CREATE TABLE people (id int)"},
	Down: []string{"DROP TABLE people"},
}, {
	Name: "0124_00_test.sql",
	Up:   []string{"ALTER TABLE people ADD COLUMN first_name text"},
	Down: []string{"SELECT 0"}, // Not really supported
},
}

func (s *SqliteMigrateSuite) TestRunMigrationPatch(c *C) {
	EnablePatchMode(true)
	migrations := &MemoryMigrationSource{
		MigrationsPatch: sqliteMigrationsPatch[:1],
	}

	// Executes one migration
	n, err := Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 1)

	// Can use table now
	_, err = s.DbMap.Exec("SELECT * FROM people")
	c.Assert(err, IsNil)

	// Shouldn't apply migration again
	n, err = Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 0)
}

func (s *SqliteMigrateSuite) TestRunMigrationEscapeTablePatch(c *C) {
	EnablePatchMode(true)
	migrations := &MemoryMigrationSource{
		MigrationsPatch: sqliteMigrationsPatch[:1],
	}

	SetTable(`my migrations`)

	// Executes one migration
	n, err := Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 1)
}

func (s *SqliteMigrateSuite) TestMigrateMultiplePatch(c *C) {
	EnablePatchMode(true)
	migrations := &MemoryMigrationSource{
		MigrationsPatch: sqliteMigrationsPatch[:2],
	}

	// Executes two migrations
	n, err := Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 2)

	// Can use column now
	_, err = s.DbMap.Exec("SELECT first_name FROM people")
	c.Assert(err, IsNil)
}

func (s *SqliteMigrateSuite) TestMigrateIncrementalPatch(c *C) {
	EnablePatchMode(true)
	migrations := &MemoryMigrationSource{
		MigrationsPatch: sqliteMigrationsPatch[:1],
	}

	// Executes one migration
	n, err := Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 1)

	// Execute a new migration
	migrations = &MemoryMigrationSource{
		MigrationsPatch: sqliteMigrationsPatch[:2],
	}
	n, err = Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 1)

	// Can use column now
	_, err = s.DbMap.Exec("SELECT first_name FROM people")
	c.Assert(err, IsNil)
}

func (s *SqliteMigrateSuite) TestFileMigratePatch(c *C) {
	EnablePatchMode(true)
	migrations := &FileMigrationSource{
		Dir: "test-migrations-patch",
	}

	// Executes two migrations
	n, err := Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 5)

	// Has data
	id, err := s.DbMap.SelectInt("SELECT id FROM people")
	c.Assert(err, IsNil)
	c.Assert(id, Equals, int64(1))
}

func (s *SqliteMigrateSuite) TestHttpFileSystemMigratePatch(c *C) {
	EnablePatchMode(true)
	migrations := &HttpFileSystemMigrationSource{
		FileSystem: http.Dir("test-migrations-patch"),
	}

	// Executes two migrations
	n, err := Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 5)

	// Has data
	id, err := s.DbMap.SelectInt("SELECT id FROM people")
	c.Assert(err, IsNil)
	c.Assert(id, Equals, int64(1))
}

//go:generate go-bindata --ignore .+\.go$ -pkg migrate -o bindata_test.go ./...
func (s *SqliteMigrateSuite) TestAssetMigratePatch(c *C) {
	EnablePatchMode(true)
	migrations := &AssetMigrationSource{
		Asset:    Asset,
		AssetDir: AssetDir,
		Dir:      "test-migrations-patch",
	}

	// Executes two migrations
	n, err := Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 5)

	// Has data
	id, err := s.DbMap.SelectInt("SELECT id FROM people")
	c.Assert(err, IsNil)
	c.Assert(id, Equals, int64(1))
}

func (s *SqliteMigrateSuite) TestPackrMigratePatch(c *C) {
	EnablePatchMode(true)
	migrations := &PackrMigrationSource{
		Box: packr.New("migrations", "test-migrations-patch"),
	}

	// Executes two migrations
	n, err := Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 5)

	// Has data
	id, err := s.DbMap.SelectInt("SELECT id FROM people")
	c.Assert(err, IsNil)
	c.Assert(id, Equals, int64(1))
}

func (s *SqliteMigrateSuite) TestPackrMigrateDirPatch(c *C) {
	EnablePatchMode(true)
	migrations := &PackrMigrationSource{
		Box: packr.NewBox("."),
		Dir: "./test-migrations-patch/",
	}

	// Executes two migrations
	n, err := Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 5)

	// Has data
	id, err := s.DbMap.SelectInt("SELECT id FROM people")
	c.Assert(err, IsNil)
	c.Assert(id, Equals, int64(1))
}

func (s *SqliteMigrateSuite) TestMigrateMaxPatch(c *C) {
	EnablePatchMode(true)
	migrations := &FileMigrationSource{
		Dir: "test-migrations-patch",
	}

	// Executes one migration
	n, err := ExecMax(s.Db, "sqlite3", migrations, Up, 1)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 1)

	id, err := s.DbMap.SelectInt("SELECT COUNT(*) FROM people")
	c.Assert(err, IsNil)
	c.Assert(id, Equals, int64(0))
}

func (s *SqliteMigrateSuite) TestMigrateDownPatch(c *C) {
	EnablePatchMode(true)
	migrations := &FileMigrationSource{
		Dir: "test-migrations-patch",
	}

	n, err := Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 5)

	// Has data
	id, err := s.DbMap.SelectInt("SELECT id FROM people")
	c.Assert(err, IsNil)
	c.Assert(id, Equals, int64(1))

	// Undo the last one
	n, err = ExecMax(s.Db, "sqlite3", migrations, Down, 1)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 1)

	// No more data
	id, err = s.DbMap.SelectInt("SELECT COUNT(*) FROM people")
	c.Assert(err, IsNil)
	c.Assert(id, Equals, int64(0))

	// Remove the table.
	n, err = ExecMax(s.Db, "sqlite3", migrations, Down, 0)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 4)

	// Cannot query it anymore
	_, err = s.DbMap.SelectInt("SELECT COUNT(*) FROM people")
	c.Assert(err, Not(IsNil))

	// Nothing left to do.
	n, err = ExecMax(s.Db, "sqlite3", migrations, Down, 1)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 0)
}

func (s *SqliteMigrateSuite) TestMigrateDownFullPatch(c *C) {
	EnablePatchMode(true)
	migrations := &FileMigrationSource{
		Dir: "test-migrations-patch",
	}

	n, err := Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 5)

	// Has data
	id, err := s.DbMap.SelectInt("SELECT id FROM people")
	c.Assert(err, IsNil)
	c.Assert(id, Equals, int64(1))

	// Undo the last one
	n, err = Exec(s.Db, "sqlite3", migrations, Down)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 5)

	// Cannot query it anymore
	_, err = s.DbMap.SelectInt("SELECT COUNT(*) FROM people")
	c.Assert(err, Not(IsNil))

	// Nothing left to do.
	n, err = Exec(s.Db, "sqlite3", migrations, Down)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 0)
}

func (s *SqliteMigrateSuite) TestMigrateTransactionPatch(c *C) {
	EnablePatchMode(true)
	migrations := &MemoryMigrationSource{
		MigrationsPatch: []*MigrationPatch{
			sqliteMigrationsPatch[0],
			sqliteMigrationsPatch[1],
			{
				Name: "0125_00_test.sql",
				Up:   []string{"INSERT INTO people (id, first_name) VALUES (1, 'Test')", "SELECT fail"},
				Down: []string{}, // Not important here
			},
		},
	}

	// Should fail, transaction should roll back the INSERT.
	n, err := Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, Not(IsNil))
	c.Assert(n, Equals, 2)

	// INSERT should be rolled back
	count, err := s.DbMap.SelectInt("SELECT COUNT(*) FROM people")
	c.Assert(err, IsNil)
	c.Assert(count, Equals, int64(0))
}

func (s *SqliteMigrateSuite) TestPlanMigrationPatch(c *C) {
	EnablePatchMode(true)
	migrations := &MemoryMigrationSource{
		MigrationsPatch: []*MigrationPatch{
			{
				Name: "0001_00_create_table.sql",
				Up:   []string{"CREATE TABLE people (id int)"},
				Down: []string{"DROP TABLE people"},
			},
			{
				Name: "0002_00_alter_table.sql",
				Up:   []string{"ALTER TABLE people ADD COLUMN first_name text"},
				Down: []string{"SELECT 0"}, // Not really supported
			},
			{
				Name: "0010_00_add_last_name.sql",
				Up:   []string{"ALTER TABLE people ADD COLUMN last_name text"},
				Down: []string{"ALTER TABLE people DROP COLUMN last_name"},
			},
		},
	}
	n, err := Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 3)

	migrations.MigrationsPatch = append(migrations.MigrationsPatch, &MigrationPatch{
		Name: "0011_00_add_middle_name.sql",
		Up:   []string{"ALTER TABLE people ADD COLUMN middle_name text"},
		Down: []string{"ALTER TABLE people DROP COLUMN middle_name"},
	})

	plannedMigrations, _, err := PlanMigrationPatch(s.Db, "sqlite3", migrations, Up, 0)
	c.Assert(err, IsNil)
	c.Assert(plannedMigrations, HasLen, 1)
	c.Assert(plannedMigrations[0].MigrationPatch, Equals, migrations.MigrationsPatch[3])

	plannedMigrations, _, err = PlanMigrationPatch(s.Db, "sqlite3", migrations, Down, 0)
	c.Assert(err, IsNil)
	c.Assert(plannedMigrations, HasLen, 3)
	c.Assert(plannedMigrations[0].MigrationPatch, Equals, migrations.MigrationsPatch[2])
	c.Assert(plannedMigrations[1].MigrationPatch, Equals, migrations.MigrationsPatch[1])
	c.Assert(plannedMigrations[2].MigrationPatch, Equals, migrations.MigrationsPatch[0])
}

func (s *SqliteMigrateSuite) TestSkipMigrationPatch(c *C) {
	EnablePatchMode(true)
	migrations := &MemoryMigrationSource{
		MigrationsPatch: []*MigrationPatch{{
			Name: "0001_00_create_table.sql",
			Up:   []string{"CREATE TABLE people (id int)"},
			Down: []string{"DROP TABLE people"},
		}, {
			Name: "0002_00_alter_table.sql",
			Up:   []string{"ALTER TABLE people ADD COLUMN first_name text"},
			Down: []string{"SELECT 0"}, // Not really supported
		}, {
			Name: "0010_00_add_last_name.sql",
			Up:   []string{"ALTER TABLE people ADD COLUMN last_name text"},
			Down: []string{"ALTER TABLE people DROP COLUMN last_name"},
		},
		},
	}
	n, err := SkipMaxPatch(s.Db, "sqlite3", migrations, Up, 0)
	// there should be no errors
	c.Assert(err, IsNil)
	// we should have detected and skipped 3 migrations
	c.Assert(n, Equals, 3)
	// should not actually have the tables now since it was skipped
	// so this query should fail
	_, err = s.DbMap.Exec("SELECT * FROM people")
	c.Assert(err, NotNil)
	// run the migrations again, should execute none of them since we pegged the db level
	// in the skip command
	n2, err2 := Exec(s.Db, "sqlite3", migrations, Up)
	// there should be no errors
	c.Assert(err2, IsNil)
	// we should not have executed any migrations
	c.Assert(n2, Equals, 0)
}

func (s *SqliteMigrateSuite) TestPlanMigrationWithHolesPatch(c *C) {
	EnablePatchMode(true)
	up := "SELECT 0"
	down := "SELECT 1"
	migrations := &MemoryMigrationSource{
		MigrationsPatch: []*MigrationPatch{{
			Name: "0001_00_name.sql",
			Up:   []string{up},
			Down: []string{down},
		}, {
			Name: "0003_00_name.sql",
			Up:   []string{up},
			Down: []string{down},
		},
		},
	}
	n, err := Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 2)

	migrations.MigrationsPatch = append(migrations.MigrationsPatch, &MigrationPatch{
		Name: "0002_00_name.sql",
		Up:   []string{up},
		Down: []string{down},
	})

	migrations.MigrationsPatch = append(migrations.MigrationsPatch, &MigrationPatch{
		Name: "0004_00_name.sql",
		Up:   []string{up},
		Down: []string{down},
	})

	migrations.MigrationsPatch = append(migrations.MigrationsPatch, &MigrationPatch{
		Name: "0005_00_name.sql",
		Up:   []string{up},
		Down: []string{down},
	})

	// apply all the missing migrations
	plannedMigrations, _, err := PlanMigrationPatch(s.Db, "sqlite3", migrations, Up, 0)
	c.Assert(err, IsNil)
	c.Assert(plannedMigrations, HasLen, 3)
	c.Assert(plannedMigrations[0].MigrationPatch.Name, Equals, "0002_00_name.sql")
	c.Assert(plannedMigrations[0].Queries[0], Equals, up)
	c.Assert(plannedMigrations[1].MigrationPatch.Name, Equals, "0004_00_name.sql")
	c.Assert(plannedMigrations[1].Queries[0], Equals, up)
	c.Assert(plannedMigrations[2].MigrationPatch.Name, Equals, "0005_00_name.sql")
	c.Assert(plannedMigrations[2].Queries[0], Equals, up)

	// first catch up to current target state 123, then migrate down 1 step to 12
	plannedMigrations, _, err = PlanMigrationPatch(s.Db, "sqlite3", migrations, Down, 1)
	c.Assert(err, IsNil)
	c.Assert(plannedMigrations, HasLen, 2)
	c.Assert(plannedMigrations[0].MigrationPatch.Name, Equals, "0002_00_name.sql")
	c.Assert(plannedMigrations[0].Queries[0], Equals, up)
	c.Assert(plannedMigrations[1].MigrationPatch.Name, Equals, "0003_00_name.sql")
	c.Assert(plannedMigrations[1].Queries[0], Equals, down)

	// first catch up to current target state 123, then migrate down 2 steps to 1
	plannedMigrations, _, err = PlanMigrationPatch(s.Db, "sqlite3", migrations, Down, 2)
	c.Assert(err, IsNil)
	c.Assert(plannedMigrations, HasLen, 3)
	c.Assert(plannedMigrations[0].MigrationPatch.Name, Equals, "0002_00_name.sql")
	c.Assert(plannedMigrations[0].Queries[0], Equals, up)
	c.Assert(plannedMigrations[1].MigrationPatch.Name, Equals, "0003_00_name.sql")
	c.Assert(plannedMigrations[1].Queries[0], Equals, down)
	c.Assert(plannedMigrations[2].MigrationPatch.Name, Equals, "0002_00_name.sql")
	c.Assert(plannedMigrations[2].Queries[0], Equals, down)
}

func (s *SqliteMigrateSuite) TestLessPatch(c *C) {
	c.Assert((MigrationPatch{VerInt: 1, PatchInt: 0}).Less(&MigrationPatch{VerInt: 2, PatchInt: 0}), Equals, true) // 1 less than 2
	c.Assert((MigrationPatch{VerInt: 2}).Less(&MigrationPatch{VerInt: 2}), Equals, false)                          // 2 not less than 1
	c.Assert((MigrationPatch{VerInt: 1}).Less(&MigrationPatch{Name: "a"}), Equals, false)                          // a(0) less than 1
	c.Assert((MigrationPatch{Name: "a"}).Less(&MigrationPatch{Name: "1"}), Equals, false)                          // a not less than 1
	c.Assert((MigrationPatch{Name: "a"}).Less(&MigrationPatch{Name: "a"}), Equals, false)                          // a not less than a
	c.Assert((MigrationPatch{Name: "1-a"}).Less(&MigrationPatch{Name: "1-b"}), Equals, true)                       // 1-a less than 1-b
	c.Assert((MigrationPatch{Name: "1-b"}).Less(&MigrationPatch{Name: "1-a"}), Equals, false)                      // 1-b not less than 1-a
	c.Assert((MigrationPatch{Name: "1"}).Less(&MigrationPatch{Name: "10"}), Equals, true)                          // 1 less than 10
	c.Assert((MigrationPatch{Name: "10"}).Less(&MigrationPatch{Name: "1"}), Equals, false)                         // 10 not less than 1
	// 20160126_1100 less than 20160126_1200
	c.Assert((MigrationPatch{VerInt: 20160126, PatchInt: 1100}).
		Less(&MigrationPatch{VerInt: 20160126, PatchInt: 1200}), Equals, true)
	// 20160126_1200 not less than 20160126_1100
	c.Assert((MigrationPatch{VerInt: 20160126, PatchInt: 1200}).
		Less(&MigrationPatch{VerInt: 20160126, PatchInt: 1100}), Equals, false)
}

func (s *SqliteMigrateSuite) TestPlanMigrationWithUnknownDatabaseMigrationAppliedPatch(c *C) {
	EnablePatchMode(true)
	migrations := &MemoryMigrationSource{
		MigrationsPatch: []*MigrationPatch{{
			Name: "0001_00_create_table.sql",
			Up:   []string{"CREATE TABLE people (id int)"},
			Down: []string{"DROP TABLE people"},
		}, {
			Name: "0002_00_alter_table.sql",
			Up:   []string{"ALTER TABLE people ADD COLUMN first_name text"},
			Down: []string{"SELECT 0"}, // Not really supported
		}, {
			Name: "0010_00_add_last_name.sql",
			Up:   []string{"ALTER TABLE people ADD COLUMN last_name text"},
			Down: []string{"ALTER TABLE people DROP COLUMN last_name"},
		},
		},
	}
	n, err := Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 3)

	// Note that migration 10_add_last_name.sql is missing from the new migrations source
	// so it is considered an "unknown" migration for the planner.
	migrations.MigrationsPatch = append(migrations.MigrationsPatch[:2], &MigrationPatch{
		Name: "0009_00_add_middle_name.sql",
		Up:   []string{"ALTER TABLE people ADD COLUMN middle_name text"},
		Down: []string{"ALTER TABLE people DROP COLUMN middle_name"},
	})

	_, _, err = PlanMigrationPatch(s.Db, "sqlite3", migrations, Up, 0)
	c.Assert(err, NotNil, Commentf("Up migrations should not have been applied when there "+
		"is an unknown migration in the database"))
	c.Assert(err, FitsTypeOf, &PlanError{})

	_, _, err = PlanMigrationPatch(s.Db, "sqlite3", migrations, Down, 0)
	c.Assert(err, NotNil, Commentf("Down migrations should not have been applied when there "+
		"is an unknown migration in the database"))
	c.Assert(err, FitsTypeOf, &PlanError{})
}

func (s *SqliteMigrateSuite) TestPlanMigrationWithIgnoredUnknownDatabaseMigrationAppliedPatch(c *C) {
	EnablePatchMode(true)
	migrations := &MemoryMigrationSource{
		MigrationsPatch: []*MigrationPatch{{
			Name: "0001_00_create_table.sql",
			Up:   []string{"CREATE TABLE people (id int)"},
			Down: []string{"DROP TABLE people"},
		}, {
			Name: "0002_00_alter_table.sql",
			Up:   []string{"ALTER TABLE people ADD COLUMN first_name text"},
			Down: []string{"SELECT 0"}, // Not really supported
		}, {
			Name: "0010_00_add_last_name.sql",
			Up:   []string{"ALTER TABLE people ADD COLUMN last_name text"},
			Down: []string{"ALTER TABLE people DROP COLUMN last_name"},
		}},
	}
	SetIgnoreUnknown(true)
	n, err := Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 3)

	// Note that migration 10_add_last_name.sql is missing from the new migrations source
	// so it is considered an "unknown" migration for the planner.
	migrations.MigrationsPatch = append(migrations.MigrationsPatch[:2], &MigrationPatch{
		Name: "0009_00_add_middle_name.sql",
		Up:   []string{"ALTER TABLE people ADD COLUMN middle_name text"},
		Down: []string{"ALTER TABLE people DROP COLUMN middle_name"},
	})

	_, _, err = PlanMigrationPatch(s.Db, "sqlite3", migrations, Up, 0)
	c.Assert(err, IsNil)

	_, _, err = PlanMigrationPatch(s.Db, "sqlite3", migrations, Down, 0)
	c.Assert(err, IsNil)
	SetIgnoreUnknown(false) // Make sure we are not breaking other tests as this is globaly set
}

// TestExecWithUnknownMigrationInDatabase makes sure that problems found with planning the
// migrations are propagated and returned by Exec.
func (s *SqliteMigrateSuite) TestExecWithUnknownMigrationInDatabasePatch(c *C) {
	EnablePatchMode(true)
	migrations := &MemoryMigrationSource{
		MigrationsPatch: sqliteMigrationsPatch[:2],
	}

	// Executes two migrations
	n, err := Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 2)

	// Then create a new migration source with one of the migrations missing
	var newSqliteMigrations = []*MigrationPatch{{
		Name: "0124_01_other.sql",
		Up:   []string{"ALTER TABLE people ADD COLUMN middle_name text"},
		Down: []string{"ALTER TABLE people DROP COLUMN middle_name"},
	}, {
		Name: "0125_00_other.sql",
		Up:   []string{"ALTER TABLE people ADD COLUMN age int"},
		Down: []string{"ALTER TABLE people DROP COLUMN age"},
	},
	}
	migrations = &MemoryMigrationSource{
		MigrationsPatch: append(sqliteMigrationsPatch[:1], newSqliteMigrations...),
	}

	n, err = Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, NotNil, Commentf("Migrations should not have been applied when there "+
		"is an unknown migration in the database"))
	c.Assert(err, FitsTypeOf, &PlanError{})
	c.Assert(n, Equals, 0)

	// Make sure the new columns are not actually created
	_, err = s.DbMap.Exec("SELECT middle_name FROM people")
	c.Assert(err, NotNil)
	_, err = s.DbMap.Exec("SELECT age FROM people")
	c.Assert(err, NotNil)
}

func (s *SqliteMigrateSuite) TestRunMigrationObjDefaultTablePatch(c *C) {
	migrations := &MemoryMigrationSource{
		MigrationsPatch: sqliteMigrationsPatch[:1],
	}

	ms := MigrationSet{
		EnablePatchMode: true,
	}
	// Executes one migration
	n, err := ms.Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 1)

	// Can use table now
	_, err = s.DbMap.Exec("SELECT * FROM people")
	c.Assert(err, IsNil)

	// Uses default tableName
	_, err = s.DbMap.Exec("SELECT * FROM gorp_migrations")
	c.Assert(err, IsNil)

	// Shouldn't apply migration again
	n, err = ms.Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 0)
}

func (s *SqliteMigrateSuite) TestRunMigrationObjOtherTablePatch(c *C) {
	migrations := &MemoryMigrationSource{
		MigrationsPatch: sqliteMigrationsPatch[:1],
	}

	ms := MigrationSet{
		TableName:       "other_migrations",
		EnablePatchMode: true,
	}
	// Executes one migration
	n, err := ms.Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 1)

	// Can use table now
	_, err = s.DbMap.Exec("SELECT * FROM people")
	c.Assert(err, IsNil)

	// Uses default tableName
	_, err = s.DbMap.Exec("SELECT * FROM other_migrations")
	c.Assert(err, IsNil)

	// Shouldn't apply migration again
	n, err = ms.Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 0)
}

// TestErrorFormatMigrationName return error format
func (s *SqliteMigrateSuite) TestErrorFormatMigrationName(c *C) {
	EnablePatchMode(true)
	migrations := &MemoryMigrationSource{
		MigrationsPatch: []*MigrationPatch{
			{
				Name: "0124_other.sql",
				Up:   []string{"ALTER TABLE people ADD COLUMN middle_name text"},
				Down: []string{"ALTER TABLE people DROP COLUMN middle_name"},
			},
		},
	}

	// Executes two migrations
	n, err := Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, Not(IsNil))
	c.Assert(n, Equals, 0)

	migrations.MigrationsPatch[0].Name = "0124_00"

	// Executes two migrations
	n, err = Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, Not(IsNil))
	c.Assert(n, Equals, 0)
}
