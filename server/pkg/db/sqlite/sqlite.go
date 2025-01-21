package sqlite

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var DB *sql.DB

// Add new type for database options
type DBOptions struct {
    ApplyMigrations bool
    TargetVersion   *uint // nil means latest version
}

// Modify OpenDB to accept options
func OpenDB(dbPath string, opts *DBOptions) error {
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("error connecting to the database: %v", err)
	}

	// Only apply migrations if explicitly requested
	if opts != nil && opts.ApplyMigrations {
		if err := applyMigrations(opts.TargetVersion); err != nil {
			return fmt.Errorf("error applying migrations: %v", err)
		}
	}

	return nil
}

func applyMigrations(targetVersion *uint) error {
	log.Println("Starting database migrations...")
	
	driver, err := sqlite3.WithInstance(DB, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("could not create driver: %v", err)
	}

	// Use relative path from project root
	migrationsPath, err := filepath.Abs("pkg/db/migrations/sqlite")
	if err != nil {
		return fmt.Errorf("could not get migrations path: %v", err)
	}
	
	// Ensure migrations directory exists
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory not found at: %s", migrationsPath)
	}

	sourceURL := fmt.Sprintf("file://%s", migrationsPath)
	log.Printf("Using migrations from: %s", sourceURL)

	m, err := migrate.NewWithDatabaseInstance(
		sourceURL,
		"sqlite3",
		driver,
	)
	if err != nil {
		return fmt.Errorf("could not create migration instance: %v", err)
	}

	if targetVersion != nil {
        log.Printf("Migrating to specific version: %d", *targetVersion)
        err = m.Migrate(*targetVersion) // Migrate to specific version
    } else {
        log.Println("Migrating to latest version")
        err = m.Up() // Migrate to latest version
    }

    if err != nil && err != migrate.ErrNoChange {
        return fmt.Errorf("could not apply migrations: %v", err)
    }

    version, _, _ := m.Version()
    log.Printf("Database migrated to version: %d", version)
    return nil
}

func ClearDatabase() error {
    // Updated list of tables based on your migration files
    tables := []string{
        "users",
        "followers",
        "follow_requests",
        "posts",
        "comments",
        "groups",
        "group_members",
        "group_posts",
        "group_post_comments",
        "chat_messages",
        "group_chat_messages",
        "notifications",
        "likes",
        "friends",  // for friends_index
    }

    // First disable foreign key constraints
    if _, err := DB.Exec("PRAGMA foreign_keys=OFF"); err != nil {
        return fmt.Errorf("error disabling foreign keys: %v", err)
    }

    for _, table := range tables {
        _, err := DB.Exec(fmt.Sprintf("DELETE FROM %s", table))
        if err != nil {
            log.Printf("Warning while clearing table %s: %v", table, err)
            continue
        }
        // Reset the SQLite autoincrement counters
        _, err = DB.Exec(fmt.Sprintf("DELETE FROM sqlite_sequence WHERE name='%s'", table))
        if err != nil {
            log.Printf("Warning while resetting sequence for table %s: %v", table, err)
        }
        log.Printf("Table %s cleared successfully\n", table)
    }

    // Re-enable foreign key constraints
    if _, err := DB.Exec("PRAGMA foreign_keys=ON"); err != nil {
        return fmt.Errorf("error re-enabling foreign keys: %v", err)
    }

    return nil
}

func CleanupMigrations() error {
    log.Println("Cleaning up migration state...")
    _, err := DB.Exec("UPDATE schema_migrations SET dirty = 0")
    if err != nil {
        return fmt.Errorf("could not clean migration state: %v", err)
    }
    return nil
}

// Update RollbackMigrations for better control
func RollbackMigrations(steps int) error {
    log.Printf("Starting rollback process for %d steps...", steps)
    
    driver, err := sqlite3.WithInstance(DB, &sqlite3.Config{})
    if err != nil {
        return fmt.Errorf("could not create driver: %v", err)
    }

    migrationsPath, err := filepath.Abs("pkg/db/migrations/sqlite")
    if err != nil {
        return fmt.Errorf("could not get migrations path: %v", err)
    }

    sourceURL := fmt.Sprintf("file://%s", migrationsPath)
    m, err := migrate.NewWithDatabaseInstance(sourceURL, "sqlite3", driver)
    if err != nil {
        return fmt.Errorf("could not create migration instance: %v", err)
    }

    // Get current version
    version, dirty, err := m.Version()
    if err != nil {
        if strings.Contains(err.Error(), "no migration") {
            return fmt.Errorf("no migrations found")
        }
        return fmt.Errorf("error checking version: %v", err)
    }

    log.Printf("Current migration version: %d, Dirty: %v", version, dirty)

    // Handle dirty state
    if dirty {
        if err := m.Force(int(version)); err != nil {
            return fmt.Errorf("could not force clean state: %v", err)
        }
        log.Printf("Forced clean state at version %d", version)
    }

    // Roll back specified number of steps
    err = m.Steps(-steps)
    if err != nil {
        if err == migrate.ErrNoChange {
            return fmt.Errorf("no migrations to roll back")
        }
        return fmt.Errorf("rollback failed: %v", err)
    }

    newVersion, _, _ := m.Version()
    log.Printf("Successfully rolled back from version %d to %d", version, newVersion)

    // Force the version to prevent auto-migrations
    if err := m.Force(int(newVersion)); err != nil {
        return fmt.Errorf("could not force version after rollback: %v", err)
    }

    return nil
}
