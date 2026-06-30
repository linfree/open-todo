package schema

import (
	"database/sql"
	"fmt"
)

type Migration struct {
	Version int
	SQL     []string
}

var Migrations = []Migration{
	{Version: 1, SQL: CreateTablesSQL},
}

func RunMigrations(db *sql.DB) error {
	_, err := db.Exec("PRAGMA journal_mode=WAL")
	if err != nil {
		return fmt.Errorf("enable WAL: %w", err)
	}

	var currentVersion int
	err = db.QueryRow("PRAGMA user_version").Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("read version: %w", err)
	}

	for _, m := range Migrations {
		if m.Version <= currentVersion {
			continue
		}
		for _, s := range m.SQL {
			if _, err := db.Exec(s); err != nil {
				return fmt.Errorf("migration v%d: %w", m.Version, err)
			}
		}
		_, err := db.Exec(fmt.Sprintf("PRAGMA user_version = %d", m.Version))
		if err != nil {
			return fmt.Errorf("set version v%d: %w", m.Version, err)
		}
	}
	return nil
}
