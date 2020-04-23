package test_bindata

import (
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/pkg/errors"

	migrate "github.com/rubenv/sql-migrate"
)

//go:generate go-bindata --ignore .+\.go$ -pkg test_bindata -o bindata.go ./...
//go:generate gofmt -w bindata.go

func Migrate(connStr string, dir migrate.MigrationDirection) (int, error) {
	src := migrate.AssetMigrationSource{
		Asset:    Asset,
		AssetDir: AssetDir,
		Dir:      "postgres",
	}
	dbConnection, err := sql.Open("postgres", connStr)
	if err != nil {
		return 0, errors.New("unable to connect to the database " + err.Error())
	}

	return migrate.Exec(dbConnection, "postgres", src, dir)
}
