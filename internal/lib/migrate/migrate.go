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

func Up(cfg *config.Postgres) error {
	d, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("%w: failed to create iofs source: %w", Err, err)
	}
	m, err := migrate.NewWithSourceInstance("embed", d, cfg.URL())
	if err != nil {
		return fmt.Errorf("%w: failed to create migrate instance: %w", Err, err)
	}
	m.Log = logger{Logger: log.Default()}

	log.Println("Starting database migrations...")
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("%w: migration failed: %w", Err, err)
	}

	log.Println("Database migrations completed successfully.")
	return nil
}
