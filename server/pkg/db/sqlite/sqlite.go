package sqlite

import (
	"database/sql"
	"fmt"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var DB *sql.DB

func OpenDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("error connecting to the database: %v", err)
	}

	if err := applyMigrations(); err != nil {
		return fmt.Errorf("error applying migrations: %v", err)
	}

	return nil
}

func applyMigrations() error {
	driver, err := sqlite3.WithInstance(DB, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("could not create driver: %v", err)
	}

	migrationsPath, err := filepath.Abs("pkg/db/migrations/sqlite")
	if err != nil {
		return fmt.Errorf("could not get migrations path: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"sqlite3",
		driver,
	)
	if err != nil {
		return fmt.Errorf("could not create migration instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not apply migrations: %v", err)
	}

	return nil
}

func ClearDatabase() error {
	tables := []string{"users", "followers", "posts", "comments", "groups", "group_members", "chat_messages", "notifications"}
	
	for _, table := range tables {
		_, err := DB.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			return fmt.Errorf("error clearing table %s: %v", table, err)
		}
	}
	return nil
}

func RollbackMigrations() error {
	driver, err := sqlite3.WithInstance(DB, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("could not create driver: %v", err)
	}

	migrationsPath, err := filepath.Abs("pkg/db/migrations/sqlite")
	if err != nil {
		return fmt.Errorf("could not get migrations path: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"sqlite3",
		driver,
	)
	if err != nil {
		return fmt.Errorf("could not create migration instance: %v", err)
	}

	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not rollback migrations: %v", err)
	}

	return nil
}
