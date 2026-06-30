package schema

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestRunMigrations(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	if err := RunMigrations(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	var version int
	db.QueryRow("PRAGMA user_version").Scan(&version)
	if version != CurrentVersion {
		t.Errorf("version = %d, want %d", version, CurrentVersion)
	}

	for _, table := range Tables {
		var count int
		db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		if count != 1 {
			t.Errorf("table %s not found", table)
		}
	}
}
