package migrate

import (
	"database/sql"
	"net/http"

	"github.com/gobuffalo/packr/v2"
	_ "github.com/mattn/go-sqlite3"
	. "gopkg.in/check.v1"
	"gopkg.in/gorp.v1"
)

var sqliteMigrations = []*Migration{
	&Migration{
		Name:     "0123_00_test.sql",
		Ver:      "0123",
		Patch:    "00",
		Up:       []string{"CREATE TABLE people (id int)"},
		Down:     []string{"DROP TABLE people"},
		verInt:   123,
		patchInt: 0,
	},
	&Migration{
		Name:     "0124_00_test.sql",
		Ver:      "0124",
		Patch:    "00",
		Up:       []string{"ALTER TABLE people ADD COLUMN first_name text"},
		Down:     []string{"SELECT 0"}, // Not really supported
		verInt:   124,
		patchInt: 0,
	},
}

type SqliteMigrateSuite struct {
	Db    *sql.DB
	DbMap *gorp.DbMap
}

var _ = Suite(&SqliteMigrateSuite{})

func (s *SqliteMigrateSuite) SetUpTest(c *C) {
	var err error
	db, err := sql.Open("sqlite3", ":memory:")
	c.Assert(err, IsNil)

	s.Db = db
	s.DbMap = &gorp.DbMap{Db: db, Dialect: &gorp.SqliteDialect{}}
}

func (s *SqliteMigrateSuite) TestRunMigration(c *C) {
	migrations := &MemoryMigrationSource{
		Migrations: sqliteMigrations[:1],
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

func (s *SqliteMigrateSuite) TestRunMigrationEscapeTable(c *C) {
	migrations := &MemoryMigrationSource{
		Migrations: sqliteMigrations[:1],
	}

	SetTable(`my migrations`)

	// Executes one migration
	n, err := Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 1)
}

func (s *SqliteMigrateSuite) TestMigrateMultiple(c *C) {
	migrations := &MemoryMigrationSource{
		Migrations: sqliteMigrations[:2],
	}

	// Executes two migrations
	n, err := Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 2)

	// Can use column now
	_, err = s.DbMap.Exec("SELECT first_name FROM people")
	c.Assert(err, IsNil)
}

func (s *SqliteMigrateSuite) TestMigrateIncremental(c *C) {
	migrations := &MemoryMigrationSource{
		Migrations: sqliteMigrations[:1],
	}

	// Executes one migration
	n, err := Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 1)

	// Execute a new migration
	migrations = &MemoryMigrationSource{
		Migrations: sqliteMigrations[:2],
	}
	n, err = Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 1)

	// Can use column now
	_, err = s.DbMap.Exec("SELECT first_name FROM people")
	c.Assert(err, IsNil)
}

func (s *SqliteMigrateSuite) TestFileMigrate(c *C) {
	migrations := &FileMigrationSource{
		Dir: "test-migrations",
	}

	// Executes two migrations
	n, err := Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 2)

	// Has data
	id, err := s.DbMap.SelectInt("SELECT id FROM people")
	c.Assert(err, IsNil)
	c.Assert(id, Equals, int64(1))
}

func (s *SqliteMigrateSuite) TestHttpFileSystemMigrate(c *C) {
	migrations := &HttpFileSystemMigrationSource{
		FileSystem: http.Dir("test-migrations"),
	}

	// Executes two migrations
	n, err := Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 2)

	// Has data
	id, err := s.DbMap.SelectInt("SELECT id FROM people")
	c.Assert(err, IsNil)
	c.Assert(id, Equals, int64(1))
}

//go:generate go-bindata --ignore .+\.go$ -pkg migrate -o bindata_test.go ./test-migrations
func (s *SqliteMigrateSuite) TestAssetMigrate(c *C) {
	migrations := &AssetMigrationSource{
		Asset:    Asset,
		AssetDir: AssetDir,
		Dir:      "test-migrations",
	}

	// Executes two migrations
	n, err := Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 2)

	// Has data
	id, err := s.DbMap.SelectInt("SELECT id FROM people")
	c.Assert(err, IsNil)
	c.Assert(id, Equals, int64(1))
}

func (s *SqliteMigrateSuite) TestPackrMigrate(c *C) {
	migrations := &PackrMigrationSource{
		Box: packr.New("migrations", "test-migrations"),
	}

	// Executes two migrations
	n, err := Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 2)

	// Has data
	id, err := s.DbMap.SelectInt("SELECT id FROM people")
	c.Assert(err, IsNil)
	c.Assert(id, Equals, int64(1))
}

func (s *SqliteMigrateSuite) TestPackrMigrateDir(c *C) {
	migrations := &PackrMigrationSource{
		Box: packr.NewBox("."),
		Dir: "./test-migrations/",
	}

	// Executes two migrations
	n, err := Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 2)

	// Has data
	id, err := s.DbMap.SelectInt("SELECT id FROM people")
	c.Assert(err, IsNil)
	c.Assert(id, Equals, int64(1))
}

func (s *SqliteMigrateSuite) TestMigrateMax(c *C) {
	migrations := &FileMigrationSource{
		Dir: "test-migrations",
	}

	// Executes one migration
	n, err := ExecMax(s.Db, "sqlite3", migrations, Up, 1)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 1)

	id, err := s.DbMap.SelectInt("SELECT COUNT(*) FROM people")
	c.Assert(err, IsNil)
	c.Assert(id, Equals, int64(0))
}

func (s *SqliteMigrateSuite) TestMigrateDown(c *C) {
	migrations := &FileMigrationSource{
		Dir: "test-migrations",
	}

	n, err := Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 2)

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
	n, err = ExecMax(s.Db, "sqlite3", migrations, Down, 1)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 1)

	// Cannot query it anymore
	_, err = s.DbMap.SelectInt("SELECT COUNT(*) FROM people")
	c.Assert(err, Not(IsNil))

	// Nothing left to do.
	n, err = ExecMax(s.Db, "sqlite3", migrations, Down, 1)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 0)
}

func (s *SqliteMigrateSuite) TestMigrateDownFull(c *C) {
	migrations := &FileMigrationSource{
		Dir: "test-migrations",
	}

	n, err := Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 2)

	// Has data
	id, err := s.DbMap.SelectInt("SELECT id FROM people")
	c.Assert(err, IsNil)
	c.Assert(id, Equals, int64(1))

	// Undo the last one
	n, err = Exec(s.Db, "sqlite3", migrations, Down)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 2)

	// Cannot query it anymore
	_, err = s.DbMap.SelectInt("SELECT COUNT(*) FROM people")
	c.Assert(err, Not(IsNil))

	// Nothing left to do.
	n, err = Exec(s.Db, "sqlite3", migrations, Down)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 0)
}

func (s *SqliteMigrateSuite) TestMigrateTransaction(c *C) {
	migrations := &MemoryMigrationSource{
		Migrations: []*Migration{
			sqliteMigrations[0],
			sqliteMigrations[1],
			&Migration{
				Name:     "0125_00_test.sql",
				Ver:      "0125",
				Patch:    "00",
				Up:       []string{"INSERT INTO people (id, first_name) VALUES (1, 'Test')", "SELECT fail"},
				Down:     []string{}, // Not important here
				verInt:   125,
				patchInt: 0,
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

func (s *SqliteMigrateSuite) TestPlanMigration(c *C) {
	migrations := &MemoryMigrationSource{
		Migrations: []*Migration{
			&Migration{
				Name:     "0001_00_create_table.sql",
				Ver:      "0001",
				Patch:    "00",
				Up:       []string{"CREATE TABLE people (id int)"},
				Down:     []string{"DROP TABLE people"},
				verInt:   1,
				patchInt: 0,
			},
			&Migration{
				Name:     "0002_00_alter_table.sql",
				Ver:      "0002",
				Patch:    "00",
				Up:       []string{"ALTER TABLE people ADD COLUMN first_name text"},
				Down:     []string{"SELECT 0"}, // Not really supported
				verInt:   2,
				patchInt: 0,
			},
			&Migration{
				Name:     "0010_00_add_last_name.sql",
				Ver:      "0010",
				Patch:    "00",
				Up:       []string{"ALTER TABLE people ADD COLUMN last_name text"},
				Down:     []string{"ALTER TABLE people DROP COLUMN last_name"},
				verInt:   10,
				patchInt: 0,
			},
		},
	}
	n, err := Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 3)

	migrations.Migrations = append(migrations.Migrations, &Migration{
		Name:     "0011_00_add_middle_name.sql",
		Ver:      "0011",
		Patch:    "00",
		Up:       []string{"ALTER TABLE people ADD COLUMN middle_name text"},
		Down:     []string{"ALTER TABLE people DROP COLUMN middle_name"},
		verInt:   11,
		patchInt: 0,
	})

	plannedMigrations, _, err := PlanMigration(s.Db, "sqlite3", migrations, Up, 0)
	c.Assert(err, IsNil)
	c.Assert(plannedMigrations, HasLen, 1)
	c.Assert(plannedMigrations[0].Migration, Equals, migrations.Migrations[3])

	plannedMigrations, _, err = PlanMigration(s.Db, "sqlite3", migrations, Down, 0)
	c.Assert(err, IsNil)
	c.Assert(plannedMigrations, HasLen, 3)
	c.Assert(plannedMigrations[0].Migration, Equals, migrations.Migrations[2])
	c.Assert(plannedMigrations[1].Migration, Equals, migrations.Migrations[1])
	c.Assert(plannedMigrations[2].Migration, Equals, migrations.Migrations[0])
}

func (s *SqliteMigrateSuite) TestSkipMigration(c *C) {
	migrations := &MemoryMigrationSource{
		Migrations: []*Migration{
			&Migration{
				Name:     "0001_00_create_table.sql",
				Ver:      "0001",
				Patch:    "00",
				Up:       []string{"CREATE TABLE people (id int)"},
				Down:     []string{"DROP TABLE people"},
				verInt:   1,
				patchInt: 0,
			},
			&Migration{
				Name:     "0002_00_alter_table.sql",
				Ver:      "0002",
				Patch:    "00",
				Up:       []string{"ALTER TABLE people ADD COLUMN first_name text"},
				Down:     []string{"SELECT 0"}, // Not really supported
				verInt:   2,
				patchInt: 0,
			},
			&Migration{
				Name:     "0010_00_add_last_name.sql",
				Ver:      "0010",
				Patch:    "00",
				Up:       []string{"ALTER TABLE people ADD COLUMN last_name text"},
				Down:     []string{"ALTER TABLE people DROP COLUMN last_name"},
				verInt:   10,
				patchInt: 0,
			},
		},
	}
	n, err := SkipMax(s.Db, "sqlite3", migrations, Up, 0)
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

func (s *SqliteMigrateSuite) TestPlanMigrationWithHoles(c *C) {
	up := "SELECT 0"
	down := "SELECT 1"
	migrations := &MemoryMigrationSource{
		Migrations: []*Migration{
			&Migration{
				Name:     "0001_00",
				Ver:      "0001",
				Patch:    "00",
				Up:       []string{up},
				Down:     []string{down},
				verInt:   1,
				patchInt: 0,
			},
			&Migration{
				Name:     "0003_00",
				Ver:      "0003",
				Patch:    "00",
				Up:       []string{up},
				Down:     []string{down},
				verInt:   3,
				patchInt: 0,
			},
		},
	}
	n, err := Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 2)

	migrations.Migrations = append(migrations.Migrations, &Migration{
		Name:     "0002_00",
		Ver:      "0002",
		Patch:    "00",
		Up:       []string{up},
		Down:     []string{down},
		verInt:   2,
		patchInt: 0,
	})

	migrations.Migrations = append(migrations.Migrations, &Migration{
		Name:     "0004_00",
		Ver:      "0004",
		Patch:    "00",
		Up:       []string{up},
		Down:     []string{down},
		verInt:   4,
		patchInt: 0,
	})

	migrations.Migrations = append(migrations.Migrations, &Migration{
		Name:     "0005_00",
		Ver:      "0005",
		Patch:    "00",
		Up:       []string{up},
		Down:     []string{down},
		verInt:   5,
		patchInt: 0,
	})

	// apply all the missing migrations
	plannedMigrations, _, err := PlanMigration(s.Db, "sqlite3", migrations, Up, 0)
	c.Assert(err, IsNil)
	c.Assert(plannedMigrations, HasLen, 3)
	c.Assert(plannedMigrations[0].Migration.Name, Equals, "0002_00")
	c.Assert(plannedMigrations[0].Queries[0], Equals, up)
	c.Assert(plannedMigrations[1].Migration.Name, Equals, "0004_00")
	c.Assert(plannedMigrations[1].Queries[0], Equals, up)
	c.Assert(plannedMigrations[2].Migration.Name, Equals, "0005_00")
	c.Assert(plannedMigrations[2].Queries[0], Equals, up)

	// first catch up to current target state 123, then migrate down 1 step to 12
	plannedMigrations, _, err = PlanMigration(s.Db, "sqlite3", migrations, Down, 1)
	c.Assert(err, IsNil)
	c.Assert(plannedMigrations, HasLen, 2)
	c.Assert(plannedMigrations[0].Migration.Name, Equals, "0002_00")
	c.Assert(plannedMigrations[0].Queries[0], Equals, up)
	c.Assert(plannedMigrations[1].Migration.Name, Equals, "0003_00")
	c.Assert(plannedMigrations[1].Queries[0], Equals, down)

	// first catch up to current target state 123, then migrate down 2 steps to 1
	plannedMigrations, _, err = PlanMigration(s.Db, "sqlite3", migrations, Down, 2)
	c.Assert(err, IsNil)
	c.Assert(plannedMigrations, HasLen, 3)
	c.Assert(plannedMigrations[0].Migration.Name, Equals, "0002_00")
	c.Assert(plannedMigrations[0].Queries[0], Equals, up)
	c.Assert(plannedMigrations[1].Migration.Name, Equals, "0003_00")
	c.Assert(plannedMigrations[1].Queries[0], Equals, down)
	c.Assert(plannedMigrations[2].Migration.Name, Equals, "0002_00")
	c.Assert(plannedMigrations[2].Queries[0], Equals, down)
}

func (s *SqliteMigrateSuite) TestLess(c *C) {
	c.Assert((Migration{verInt: 1, patchInt: 0}).Less(&Migration{verInt: 2, patchInt: 0}), Equals, true) // 1 less than 2
	c.Assert((Migration{verInt: 2}).Less(&Migration{verInt: 2}), Equals, false)                          // 2 not less than 1
	c.Assert((Migration{verInt: 1}).Less(&Migration{Name: "a"}), Equals, false)                          // a(0) less than 1
	c.Assert((Migration{Name: "a"}).Less(&Migration{Name: "1"}), Equals, false)                          // a not less than 1
	c.Assert((Migration{Name: "a"}).Less(&Migration{Name: "a"}), Equals, false)                          // a not less than a
	c.Assert((Migration{Name: "1-a"}).Less(&Migration{Name: "1-b"}), Equals, true)                       // 1-a less than 1-b
	c.Assert((Migration{Name: "1-b"}).Less(&Migration{Name: "1-a"}), Equals, false)                      // 1-b not less than 1-a
	c.Assert((Migration{Name: "1"}).Less(&Migration{Name: "10"}), Equals, true)                          // 1 less than 10
	c.Assert((Migration{Name: "10"}).Less(&Migration{Name: "1"}), Equals, false)                         // 10 not less than 1
	// 20160126_1100 less than 20160126_1200
	c.Assert((Migration{verInt: 20160126, patchInt: 1100}).
		Less(&Migration{verInt: 20160126, patchInt: 1200}), Equals, true)
	// 20160126_1200 not less than 20160126_1100
	c.Assert((Migration{verInt: 20160126, patchInt: 1200}).
		Less(&Migration{verInt: 20160126, patchInt: 1100}), Equals, false)
}

func (s *SqliteMigrateSuite) TestPlanMigrationWithUnknownDatabaseMigrationApplied(c *C) {
	migrations := &MemoryMigrationSource{
		Migrations: []*Migration{
			&Migration{
				Name:  "0001_00_create_table.sql",
				Ver:   "0001",
				Patch: "00",
				Up:    []string{"CREATE TABLE people (id int)"},
				Down:  []string{"DROP TABLE people"},
			},
			&Migration{
				Name:  "0002_00_alter_table.sql",
				Ver:   "0002",
				Patch: "00",
				Up:    []string{"ALTER TABLE people ADD COLUMN first_name text"},
				Down:  []string{"SELECT 0"}, // Not really supported
			},
			&Migration{
				Name:  "0010_00_add_last_name.sql",
				Ver:   "0010",
				Patch: "00",
				Up:    []string{"ALTER TABLE people ADD COLUMN last_name text"},
				Down:  []string{"ALTER TABLE people DROP COLUMN last_name"},
			},
		},
	}
	n, err := Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 3)

	// Note that migration 10_add_last_name.sql is missing from the new migrations source
	// so it is considered an "unknown" migration for the planner.
	migrations.Migrations = append(migrations.Migrations[:2], &Migration{
		Name:  "0009_00_add_middle_name.sql",
		Ver:   "0009",
		Patch: "00",
		Up:    []string{"ALTER TABLE people ADD COLUMN middle_name text"},
		Down:  []string{"ALTER TABLE people DROP COLUMN middle_name"},
	})

	_, _, err = PlanMigration(s.Db, "sqlite3", migrations, Up, 0)
	c.Assert(err, NotNil, Commentf("Up migrations should not have been applied when there "+
		"is an unknown migration in the database"))
	c.Assert(err, FitsTypeOf, &PlanError{})

	_, _, err = PlanMigration(s.Db, "sqlite3", migrations, Down, 0)
	c.Assert(err, NotNil, Commentf("Down migrations should not have been applied when there "+
		"is an unknown migration in the database"))
	c.Assert(err, FitsTypeOf, &PlanError{})
}

func (s *SqliteMigrateSuite) TestPlanMigrationWithIgnoredUnknownDatabaseMigrationApplied(c *C) {
	migrations := &MemoryMigrationSource{
		Migrations: []*Migration{
			&Migration{
				Name:  "0001_00_create_table.sql",
				Ver:   "0001",
				Patch: "00",
				Up:    []string{"CREATE TABLE people (id int)"},
				Down:  []string{"DROP TABLE people"},
			},
			&Migration{
				Name:  "0002_00_alter_table.sql",
				Ver:   "0002",
				Patch: "00",
				Up:    []string{"ALTER TABLE people ADD COLUMN first_name text"},
				Down:  []string{"SELECT 0"}, // Not really supported
			},
			&Migration{
				Name:  "0010_00_add_last_name.sql",
				Ver:   "0010",
				Patch: "00",
				Up:    []string{"ALTER TABLE people ADD COLUMN last_name text"},
				Down:  []string{"ALTER TABLE people DROP COLUMN last_name"},
			},
		},
	}
	SetIgnoreUnknown(true)
	n, err := Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 3)

	// Note that migration 10_add_last_name.sql is missing from the new migrations source
	// so it is considered an "unknown" migration for the planner.
	migrations.Migrations = append(migrations.Migrations[:2], &Migration{
		Name:  "0009_00_add_middle_name.sql",
		Ver:   "0009",
		Patch: "00",
		Up:    []string{"ALTER TABLE people ADD COLUMN middle_name text"},
		Down:  []string{"ALTER TABLE people DROP COLUMN middle_name"},
	})

	_, _, err = PlanMigration(s.Db, "sqlite3", migrations, Up, 0)
	c.Assert(err, IsNil)

	_, _, err = PlanMigration(s.Db, "sqlite3", migrations, Down, 0)
	c.Assert(err, IsNil)
	SetIgnoreUnknown(false) // Make sure we are not breaking other tests as this is globaly set
}

// TestExecWithUnknownMigrationInDatabase makes sure that problems found with planning the
// migrations are propagated and returned by Exec.
func (s *SqliteMigrateSuite) TestExecWithUnknownMigrationInDatabase(c *C) {
	migrations := &MemoryMigrationSource{
		Migrations: sqliteMigrations[:2],
	}

	// Executes two migrations
	n, err := Exec(s.Db, "sqlite3", migrations, Up)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 2)

	// Then create a new migration source with one of the migrations missing
	var newSqliteMigrations = []*Migration{
		&Migration{
			Name: "0124_00_other.sql",
			Up:   []string{"ALTER TABLE people ADD COLUMN middle_name text"},
			Down: []string{"ALTER TABLE people DROP COLUMN middle_name"},
		},
		&Migration{
			Name: "0125_00_other.sql",
			Up:   []string{"ALTER TABLE people ADD COLUMN age int"},
			Down: []string{"ALTER TABLE people DROP COLUMN age"},
		},
	}
	migrations = &MemoryMigrationSource{
		Migrations: append(sqliteMigrations[:1], newSqliteMigrations...),
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

func (s *SqliteMigrateSuite) TestRunMigrationObjDefaultTable(c *C) {
	migrations := &MemoryMigrationSource{
		Migrations: sqliteMigrations[:1],
	}

	ms := MigrationSet{}
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

func (s *SqliteMigrateSuite) TestRunMigrationObjOtherTable(c *C) {
	migrations := &MemoryMigrationSource{
		Migrations: sqliteMigrations[:1],
	}

	ms := MigrationSet{TableName: "other_migrations"}
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
