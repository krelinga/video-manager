package pgxcatalog

import (
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

type Migrator struct {
	ConnStr string
}

//go:embed migrations/*.sql
var migrationsFS embed.FS

func (m *Migrator) MigrateUp() error {
	return m.impl(func(migrator *migrate.Migrate) error {
		if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("failed to migrate up: %w", err)
		}
		return nil
	})
}

func (m *Migrator) MigrateDown() error {
	return m.impl(func(migrator *migrate.Migrate) error {
		if err := migrator.Down(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("failed to migrate down: %w", err)
		}
		return nil
	})
}

func (m *Migrator) impl(body func(migrator *migrate.Migrate) error) error {
	d, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create iofs driver: %w", err)
	}

	migrator, err := migrate.NewWithSourceInstance("iofs", d, m.ConnStr)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	var closed bool
	defer func() {
		if closed {
			return
		}
		migrator.Close()
	}()

	if err := body(migrator); err != nil {
		return err
	}

	sourceErr, databaseErr := migrator.Close()
	closed = true
	if sourceErr != nil {
		return fmt.Errorf("failed to close migration source: %w", sourceErr)
	}
	if databaseErr != nil {
		return fmt.Errorf("failed to close migration database: %w", databaseErr)
	}

	return nil
}
