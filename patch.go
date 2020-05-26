package migrate

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/gorp.v1"

	"github.com/lancer-kit/sql-migrate/sqlparse"
)

var numberPrefixPatchRegex = regexp.MustCompile(`^(\d+)_(\d+)_.+$`)

type MigrationPatch struct {
	Name     string
	Ver      string
	VerInt   int64
	Patch    string
	PatchInt int64
	Up       []string
	Down     []string

	DisableTransactionUp   bool
	DisableTransactionDown bool
}

func (m MigrationPatch) Less(other *MigrationPatch) bool {
	if m.VerInt == other.VerInt && m.PatchInt == other.PatchInt {
		return m.Name < other.Name
	}
	return m.VerInt < other.VerInt || (m.VerInt == other.VerInt && m.PatchInt < other.PatchInt)
}

func (m *MigrationPatch) ParseName() error {
	var err error
	m.VerInt, err = strconv.ParseInt(m.Ver, 10, 64)
	if err != nil {
		return fmt.Errorf("could not parse version %q into int64: %s", m.Name, err)
	}

	m.PatchInt, err = strconv.ParseInt(m.Patch, 10, 64)
	if err != nil {
		return fmt.Errorf("could not parse patch %q into int64: %s", m.Name, err)
	}
	return nil
}

type PlannedMigrationPatch struct {
	*MigrationPatch

	DisableTransaction bool
	Queries            []string
}

type byIdPatch []*MigrationPatch

func (b byIdPatch) Len() int           { return len(b) }
func (b byIdPatch) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byIdPatch) Less(i, j int) bool { return b[i].Less(b[j]) }

type MigrationPatchRecord struct {
	Ver       string    `db:"ver"`
	Patch     string    `db:"patch"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (m MemoryMigrationSource) FindMigrationsPatch() ([]*MigrationPatch, error) {
	// Make sure migrations are sorted. In order to make the MemoryMigrationSource safe for
	// concurrent use we should not mutate it in place. So `FindMigrations` would sort a copy
	// of the m.Migrations.
	migrations := make([]*MigrationPatch, len(m.MigrationsPatch))
	copy(migrations, m.MigrationsPatch)
	for _, migration := range migrations {
		prefixMatches := numberPrefixPatchRegex.FindStringSubmatch(migration.Name)
		if len(prefixMatches) < 3 {
			return nil, fmt.Errorf("failed. Name migrations %s not format 0000_00_name.sql", migration.Name)
		}

		migration.Ver = prefixMatches[1]
		migration.Patch = prefixMatches[2]

		if err := migration.ParseName(); err != nil {
			return nil, err
		}
	}
	sort.Sort(byIdPatch(migrations))
	return migrations, nil
}

func (f HttpFileSystemMigrationSource) FindMigrationsPatch() ([]*MigrationPatch, error) {
	return findMigrationsPatch(f.FileSystem)
}

func (f FileMigrationSource) FindMigrationsPatch() ([]*MigrationPatch, error) {
	filesystem := http.Dir(f.Dir)
	return findMigrationsPatch(filesystem)
}

func findMigrationsPatch(dir http.FileSystem) ([]*MigrationPatch, error) {
	migrations := make([]*MigrationPatch, 0)

	file, err := dir.Open("/")
	if err != nil {
		return nil, err
	}

	files, err := file.Readdir(0)
	if err != nil {
		return nil, err
	}

	for _, info := range files {
		if strings.HasSuffix(info.Name(), ".sql") {
			migration, err := migrationFromFilePatch(dir, info)
			if err != nil {
				return nil, err
			}

			migrations = append(migrations, migration)
		}
	}

	// Make sure migrations are sorted
	sort.Sort(byIdPatch(migrations))

	return migrations, nil
}

func migrationFromFilePatch(dir http.FileSystem, info os.FileInfo) (*MigrationPatch, error) {
	file, err := dir.Open(info.Name())
	if err != nil {
		return nil, fmt.Errorf("Error while opening %s: %s", info.Name(), err)
	}
	defer func() { _ = file.Close() }()

	migration, err := ParseMigrationPatch(info.Name(), file)
	if err != nil {
		return nil, fmt.Errorf("Error while parsing %s: %s", info.Name(), err)
	}
	return migration, nil
}

func (a AssetMigrationSource) FindMigrationsPatch() ([]*MigrationPatch, error) {
	migrations := make([]*MigrationPatch, 0)

	files, err := a.AssetDir(a.Dir)
	if err != nil {
		return nil, err
	}

	for _, name := range files {
		if strings.HasSuffix(name, ".sql") {
			file, err := a.Asset(path.Join(a.Dir, name))
			if err != nil {
				return nil, err
			}

			migration, err := ParseMigrationPatch(name, bytes.NewReader(file))
			if err != nil {
				return nil, err
			}

			migrations = append(migrations, migration)
		}
	}

	// Make sure migrations are sorted
	sort.Sort(byIdPatch(migrations))

	return migrations, nil
}

func (p PackrMigrationSource) FindMigrationsPatch() ([]*MigrationPatch, error) {
	migrations := make([]*MigrationPatch, 0)
	items := p.Box.List()

	prefix := ""
	dir := path.Clean(p.Dir)
	if dir != "." {
		prefix = fmt.Sprintf("%s/", dir)
	}

	for _, item := range items {
		if !strings.HasPrefix(item, prefix) {
			continue
		}
		name := strings.TrimPrefix(item, prefix)
		if strings.Contains(name, "/") {
			continue
		}

		if strings.HasSuffix(name, ".sql") {
			file, err := p.Box.Find(item)
			if err != nil {
				return nil, err
			}

			migration, err := ParseMigrationPatch(name, bytes.NewReader(file))
			if err != nil {
				return nil, err
			}

			migrations = append(migrations, migration)
		}
	}

	// Make sure migrations are sorted
	sort.Sort(byIdPatch(migrations))

	return migrations, nil
}

// Migration parsing
func ParseMigrationPatch(nameFile string, r io.ReadSeeker) (*MigrationPatch, error) {
	m := &MigrationPatch{
		Name: nameFile,
	}

	prefixMatches := numberPrefixPatchRegex.FindStringSubmatch(m.Name)
	if len(prefixMatches) < 3 {
		return nil, fmt.Errorf("failed. Name migrations %s not format 0000_00_name.sql", m.Name)
	}

	m.Ver = prefixMatches[1]
	m.Patch = prefixMatches[2]

	err := m.ParseName()
	if err != nil {
		return nil, fmt.Errorf("error parsing name migrations (%s): %s", nameFile, err)
	}

	parsed, err := sqlparse.ParseMigration(r)
	if err != nil {
		return nil, fmt.Errorf("Error parsing migration (%s): %s", nameFile, err)
	}

	m.Up = parsed.UpStatements
	m.Down = parsed.DownStatements

	m.DisableTransactionUp = parsed.DisableTransactionUp
	m.DisableTransactionDown = parsed.DisableTransactionDown

	return m, nil
}

// Returns the number of applied migrations.
func (ms MigrationSet) ExecMaxPatch(db *sql.DB, dialect string, m MigrationSource, dir MigrationDirection, max int) (int, error) {
	migrations, dbMap, err := ms.PlanMigrationPatch(db, dialect, m, dir, max)
	if err != nil {
		return 0, err
	}

	minPatches := make(map[int64]int64)
	for _, migration := range migrations {
		curMinPatch, ok := minPatches[migration.VerInt]
		if !ok || migration.PatchInt < curMinPatch {
			minPatches[migration.VerInt] = migration.PatchInt
		}
	}

	// Apply migrations
	applied := 0
	for _, migration := range migrations {
		var executor SqlExecutor

		if migration.DisableTransaction {
			executor = dbMap
		} else {
			executor, err = dbMap.Begin()
			if err != nil {
				return applied, newTxError(migration.Name, err)
			}
		}

		for _, stmt := range migration.Queries {
			// remove the semicolon from stmt, fix ORA-00922 issue in database oracle
			stmt = strings.TrimSuffix(stmt, "\n")
			stmt = strings.TrimSuffix(stmt, " ")
			stmt = strings.TrimSuffix(stmt, ";")
			if _, err := executor.Exec(stmt); err != nil {
				if trans, ok := executor.(*gorp.Transaction); ok {
					_ = trans.Rollback()
				}

				return applied, newTxError(migration.Name, err)
			}
		}

		switch dir {
		case Up:
			obj, err := executor.Get(MigrationPatchRecord{}, migration.Ver)
			if err != nil {
				if trans, ok := executor.(*gorp.Transaction); ok {
					_ = trans.Rollback()
				}
				return applied, newTxError(migration.Name, err)
			}

			if obj == nil {
				err = executor.Insert(&MigrationPatchRecord{
					Ver:       migration.Ver,
					Patch:     migration.Patch,
					Name:      migration.Name,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				})
			} else {
				original := obj.(*MigrationPatchRecord)
				original.Patch = migration.Patch
				original.UpdatedAt = time.Now()
				_, err = executor.Update(original)
			}
			if err != nil {
				if trans, ok := executor.(*gorp.Transaction); ok {
					_ = trans.Rollback()
				}

				return applied, newTxError(migration.Name, err)
			}
		case Down:
			minPatch, ok := minPatches[migration.VerInt]

			if !ok || migration.PatchInt == minPatch {
				_, err = executor.Delete(&MigrationPatchRecord{
					Ver:   migration.Ver,
					Patch: migration.Patch,
					Name:  migration.Name,
				})
			} else {
				obj, err := executor.Get(MigrationPatchRecord{}, migration.Ver)
				if err != nil {
					if trans, ok := executor.(*gorp.Transaction); ok {
						_ = trans.Rollback()
					}
					return 0, err
				}
				original := obj.(*MigrationPatchRecord)
				original.Patch = migration.Patch
				original.UpdatedAt = time.Now()
				_, err = executor.Update(original)
			}
			if err != nil {
				if trans, ok := executor.(*gorp.Transaction); ok {
					_ = trans.Rollback()
				}

				return applied, newTxError(migration.Name, err)
			}
		default:
			panic("Not possible")
		}

		if trans, ok := executor.(*gorp.Transaction); ok {
			if err := trans.Commit(); err != nil {
				return applied, newTxError(migration.Name, err)
			}
		}

		applied++
	}

	return applied, nil
}

// Plan a migration.
func PlanMigrationPatch(db *sql.DB, dialect string, m MigrationSource, dir MigrationDirection,
	max int) ([]*PlannedMigrationPatch, *gorp.DbMap, error) {
	return migSet.PlanMigrationPatch(db, dialect, m, dir, max)
}

func (ms MigrationSet) PlanMigrationPatch(db *sql.DB, dialect string, m MigrationSource, dir MigrationDirection,
	max int) ([]*PlannedMigrationPatch, *gorp.DbMap, error) {
	dbMap, err := ms.getMigrationDbMap(db, dialect)
	if err != nil {
		return nil, nil, err
	}

	migrations, err := m.FindMigrationsPatch()
	if err != nil {
		return nil, nil, err
	}

	var migrationRecords []MigrationPatchRecord
	_, err = dbMap.Select(&migrationRecords, fmt.Sprintf("SELECT * FROM %s", dbMap.Dialect.QuotedTableForQuery(ms.SchemaName, ms.getTableName())))
	if err != nil {
		return nil, nil, err
	}

	// Sort migrations that have been run by Id.
	var existingMigrations []*MigrationPatch
	for _, migrationRecord := range migrationRecords {
		em := &MigrationPatch{
			Name:  migrationRecord.Name,
			Ver:   migrationRecord.Ver,
			Patch: migrationRecord.Patch,
		}

		if err := em.ParseName(); err != nil {
			return nil, nil, err
		}
		existingMigrations = append(existingMigrations, em)
	}
	sort.Sort(byIdPatch(existingMigrations))

	// Make sure all migrations in the database are among the found migrations which
	// are to be applied.
	if !ms.IgnoreUnknown {
		migrationsSearch := make(map[string]int64)
		for _, migration := range migrations {
			migrationsSearch[migration.Ver] = migration.PatchInt
		}

		for _, existingMigration := range existingMigrations {
			if _, ok := migrationsSearch[existingMigration.Ver]; !ok {
				return nil, nil, newPlanError(existingMigration.Name, "unknown migration in database")
			}
		}
	}

	// Get last migration that was run
	record := &MigrationPatch{}
	if len(existingMigrations) > 0 {
		record = existingMigrations[len(existingMigrations)-1]
	}

	result := make([]*PlannedMigrationPatch, 0)

	// Add missing migrations up to the last run migration.
	// This can happen for example when merges happened.
	if len(existingMigrations) > 0 {
		result = append(result, ToCatchupPatch(migrations, existingMigrations, record)...)
	}

	// Figure out which migrations to apply
	toApply := ToApplyPatch(migrations, record, dir)
	toApplyCount := len(toApply)
	if max > 0 && max < toApplyCount {
		toApplyCount = max
	}
	for _, v := range toApply[0:toApplyCount] {

		if dir == Up {
			result = append(result, &PlannedMigrationPatch{
				MigrationPatch:     v,
				Queries:            v.Up,
				DisableTransaction: v.DisableTransactionUp,
			})
		} else if dir == Down {
			result = append(result, &PlannedMigrationPatch{
				MigrationPatch:     v,
				Queries:            v.Down,
				DisableTransaction: v.DisableTransactionDown,
			})
		}
	}

	return result, dbMap, nil
}

// Skip a set of migrations
//
// Will skip at most `max` migrations. Pass 0 for no limit.
//
// Returns the number of skipped migrations.
func SkipMaxPatch(db *sql.DB, dialect string, m MigrationSource, dir MigrationDirection, max int) (int, error) {
	migrations, dbMap, err := PlanMigrationPatch(db, dialect, m, dir, max)
	if err != nil {
		return 0, err
	}

	// Skip migrations
	applied := 0
	for _, migration := range migrations {
		var executor SqlExecutor

		if migration.DisableTransaction {
			executor = dbMap
		} else {
			executor, err = dbMap.Begin()
			if err != nil {
				return applied, newTxError(migration.Name, err)
			}
		}

		obj, err := executor.Get(MigrationPatchRecord{}, migration.Ver)
		if err != nil {
			if trans, ok := executor.(*gorp.Transaction); ok {
				_ = trans.Rollback()
			}
			return 0, newTxError(migration.Name, err)
		}

		if obj == nil {
			err = executor.Insert(&MigrationPatchRecord{
				Ver:       migration.Ver,
				Patch:     migration.Patch,
				Name:      migration.Name,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			})
		} else {
			original := obj.(*MigrationPatchRecord)
			original.Patch = migration.Patch
			original.UpdatedAt = time.Now()
			_, err = executor.Update(original)
		}

		if err != nil {
			if trans, ok := executor.(*gorp.Transaction); ok {
				_ = trans.Rollback()
			}

			return applied, newTxError(migration.Name, err)
		}

		if trans, ok := executor.(*gorp.Transaction); ok {
			if err := trans.Commit(); err != nil {
				return applied, newTxError(migration.Name, err)
			}
		}

		applied++
	}

	return applied, nil
}

// Filter a slice of migrations into ones that should be applied.
func ToApplyPatch(migrations []*MigrationPatch, current *MigrationPatch, direction MigrationDirection) []*MigrationPatch {
	var index = -1
	if current.Name != "" {
		for index < len(migrations)-1 {
			index++
			if migrations[index].VerInt == current.VerInt && migrations[index].PatchInt == current.PatchInt {
				break
			}
		}
	}

	if direction == Up {
		return migrations[index+1:]
	} else if direction == Down {
		if index == -1 {
			return []*MigrationPatch{}
		}

		// Add in reverse order
		toApply := make([]*MigrationPatch, index+1)
		for i := 0; i < index+1; i++ {
			toApply[index-i] = migrations[i]
		}
		return toApply
	}

	panic("Not possible")
}

func ToCatchupPatch(migrations, existingMigrations []*MigrationPatch, lastRun *MigrationPatch) []*PlannedMigrationPatch {
	missing := make([]*PlannedMigrationPatch, 0)
	for _, migration := range migrations {
		found := false
		for _, existing := range existingMigrations {
			if existing.VerInt == migration.VerInt && existing.PatchInt >= migration.PatchInt {
				found = true
				break
			}
		}
		if !found && migration.Less(lastRun) {
			missing = append(missing, &PlannedMigrationPatch{
				MigrationPatch:     migration,
				Queries:            migration.Up,
				DisableTransaction: migration.DisableTransactionUp,
			})
		}
	}
	return missing
}

func GetMigrationPatchRecords(db *sql.DB, dialect string) ([]*MigrationPatchRecord, error) {
	return migSet.GetMigrationPatchRecords(db, dialect)
}

func (ms MigrationSet) GetMigrationPatchRecords(db *sql.DB, dialect string) ([]*MigrationPatchRecord, error) {
	dbMap, err := ms.getMigrationDbMap(db, dialect)
	if err != nil {
		return nil, err
	}

	var records []*MigrationPatchRecord
	query := fmt.Sprintf("SELECT * FROM %s ORDER BY ver ASC", dbMap.Dialect.QuotedTableForQuery(ms.SchemaName, ms.getTableName()))
	_, err = dbMap.Select(&records, query)
	if err != nil {
		return nil, err
	}

	return records, nil
}
