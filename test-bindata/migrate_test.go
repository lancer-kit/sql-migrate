package test_bindata

import (
	"fmt"
	"testing"

	migrate "github.com/rubenv/sql-migrate"
)

func TestMigrateDown(t *testing.T) {
	connStr := "postgres://admin:admin@localhost:5488/exchange?sslmode=disable"
	countDown, err := Migrate(connStr, migrate.Down)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("down", countDown)
}

func TestMigrateUp(t *testing.T) {
	connStr := "postgres://admin:admin@localhost:5488/exchange?sslmode=disable"
	countUp, err := Migrate(connStr, migrate.Up)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("up", countUp)
}