package config_test

import (
	"testing"

	"github.com/krelinga/go-libs/deep"
	"github.com/krelinga/go-libs/exam"
	"github.com/krelinga/go-libs/match"
	"github.com/krelinga/video-manager/internal/lib/config"
)

func TestNew(t *testing.T) {
	e := exam.New(t)
	env := deep.NewEnv()

	// Required variables.
	tempDir := e.TempDir()
	exam.SetEnv(e, config.EnvPostgresHost, "localhost")
	exam.SetEnv(e, config.EnvPostgresUser, "testuser")
	exam.SetEnv(e, config.EnvPostgresDBName, "testdb")
	exam.SetEnv(e, config.EnvPostgresPassword, "testpassword")
	exam.SetEnv(e, config.EnvInboxDVDDir, tempDir)

	e.Run("successful config creation", func(e exam.E) {
		cfg := config.New()
		expectedPg := &config.Postgres{
			Host:     "localhost",
			Port:   5432,
			DBName: "testdb",
			User:    "testuser",
			Password: "testpassword",
		}
		exam.Equal(e, env, expectedPg, cfg.Postgres)
	})

	e.Run("override postgres port", func(e exam.E) {
		exam.SetEnv(e, config.EnvPostgresPort, "6543")
		cfg := config.New()
		exam.Equal(e, env, 6543, cfg.Postgres.Port)
	})

	e.Run("malformed postgres port", func(e exam.E) {
		exam.SetEnv(e, config.EnvPostgresPort, "notanint")
		exam.PanicWith(e, env, match.As[error](match.ErrorIs(config.ErrMalformedEnvVar)), func() {
			config.New()
		})
	})

	e.Run("required vars missing", func(e exam.E) {
		tests := []string{
			config.EnvPostgresHost,
			config.EnvPostgresDBName,
			config.EnvPostgresUser,
			config.EnvPostgresPassword,
			config.EnvInboxDVDDir,
		}
		for _, v := range tests {
			e.Run(v, func(e exam.E) {
				exam.ClearEnv(e, v)
				exam.PanicWith(e, env, match.As[error](match.ErrorIs(config.ErrMissingRequiredEnvVar)), func() {
					config.New()
				})
			})
		}
	})

}
