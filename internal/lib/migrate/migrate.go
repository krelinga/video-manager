package migrate

import (
	"embed"
	"errors"
	"fmt"
	"log"

	"github.com/krelinga/video-manager/internal/lib/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations
var migrationsFS embed.FS

var Err = errors.New("migration error")

type logger struct {
	*log.Logger
}

func (l logger) Verbose() bool {
	return false
}

func setup(cfg *config.Postgres) (*migrate.Migrate, error) {
	d, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create iofs source: %w", Err, err)
	}
	m, err := migrate.NewWithSourceInstance("embed", d, cfg.URL())
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create migrate instance: %w", Err, err)
	}
	m.Log = logger{Logger: log.Default()}
	return m, nil
}

func Up(cfg *config.Postgres) error {
	m, err := setup(cfg)
	if err != nil {
		return err
	}
	log.Println("Starting database UP migrations...")
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("%w: migration failed: %w", Err, err)
	}

	log.Println("Database UP migrations completed successfully.")
	return nil
}

func Down(cfg *config.Postgres) error {
	m, err := setup(cfg)
	if err != nil {
		return err
	}
	log.Println("Starting database DOWN migrations...")
	err = m.Down()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("%w: migration failed: %w", Err, err)
	}

	log.Println("Database DOWN migrations completed successfully.")
	return nil
}
