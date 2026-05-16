package store

import (
	"reflect"
	"strings"
	"testing"
)

func TestDialectorSupportsConfiguredDrivers(t *testing.T) {
	tests := []struct {
		name     string
		driver   string
		dsn      string
		wantType string
	}{
		{name: "empty driver defaults to sqlite", driver: "", dsn: "file::memory:?cache=shared", wantType: "*sqlite.Dialector"},
		{name: "sqlite", driver: "sqlite", dsn: "file::memory:?cache=shared", wantType: "*sqlite.Dialector"},
		{name: "sqlite3 alias", driver: "sqlite3", dsn: "file::memory:?cache=shared", wantType: "*sqlite.Dialector"},
		{name: "mysql", driver: "mysql", dsn: "user:pass@tcp(localhost:3306)/app", wantType: "*mysql.Dialector"},
		{name: "mariadb alias", driver: "mariadb", dsn: "user:pass@tcp(localhost:3306)/app", wantType: "*mysql.Dialector"},
		{name: "postgres", driver: "postgres", dsn: "host=localhost user=postgres dbname=app", wantType: "*postgres.Dialector"},
		{name: "postgresql alias", driver: "postgresql", dsn: "host=localhost user=postgres dbname=app", wantType: "*postgres.Dialector"},
		{name: "pgsql alias", driver: "pgsql", dsn: "host=localhost user=postgres dbname=app", wantType: "*postgres.Dialector"},
		{name: "sqlserver", driver: "sqlserver", dsn: "sqlserver://user:pass@localhost:1433?database=app", wantType: "*sqlserver.Dialector"},
		{name: "mssql alias", driver: "mssql", dsn: "sqlserver://user:pass@localhost:1433?database=app", wantType: "*sqlserver.Dialector"},
		{name: "normalizes case and whitespace", driver: " PostgreSQL ", dsn: "host=localhost user=postgres dbname=app", wantType: "*postgres.Dialector"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dial, err := dialector(tt.driver, tt.dsn)
			if err != nil {
				t.Fatalf("dialector() error = %v", err)
			}
			if got := reflect.TypeOf(dial).String(); got != tt.wantType {
				t.Fatalf("dialector() type = %s, want %s", got, tt.wantType)
			}
		})
	}
}

func TestDialectorRejectsUnsupportedDriver(t *testing.T) {
	_, err := dialector("oracle", "")
	if err == nil {
		t.Fatal("dialector() expected unsupported driver error")
	}
	if !strings.Contains(err.Error(), "supported drivers") {
		t.Fatalf("dialector() error = %q, want supported drivers hint", err.Error())
	}
}
