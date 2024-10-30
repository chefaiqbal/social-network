package sqlite

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func OpenDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	// test the database connection
	if err := DB.Ping(); err != nil {
		return err
	}

	log.Println("Connected to database")

	if err := runMigrations(); err != nil {
		log.Fatalf("Migration Error: %v", err)
	}

	return nil
}

// Flush all database tables and remove all data
func ClearDatabase() error {
	// Start a transaction
	tx, err := DB.Begin()
	if err != nil {
		return err
	}

	// Query all table names in the database
	rows, err := tx.Query("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer rows.Close()

	// Loop through tables and delete data
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			tx.Rollback()
			return err
		}

		// Execute deletion for each table
		_, err = tx.Exec(fmt.Sprintf("DELETE FROM %s", tableName))
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return err
	}

	log.Println("Database has been flushed")
	return nil
}

func runMigrations() error {
	driver, err := sqlite3.WithInstance(DB, &sqlite3.Config{})
	if err != nil {
		return err
	}

	absPath, err := filepath.Abs("/pkg/db/migrations/sqlite")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance("file:/"+absPath, "sqlite3", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	log.Println("Migration applied succesffuly")
	return nil
}

func RollbackMigrations() error {
	driver, err := sqlite3.WithInstance(DB, &sqlite3.Config{})
	if err != nil {
		return err
	}

	absPath, err := filepath.Abs("/pkg/db/migrations/sqlite")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance("file:/"+absPath, "sqlite3", driver)
	if err != nil {
		return err
	}

	// roll back migrations all the way down
	if err := m.Down(); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("No migrations to rollback")
		} else {
			// if error is encounterd force the migration into a clean state
			if err := m.Force(0); err != nil {
				return fmt.Errorf("Error forcing migration reset: %v", err)
			}
			log.Println("migrations reset to version 0")
		}
	}

	log.Printf("All migrations rollbacked successfully")
	return nil
}
