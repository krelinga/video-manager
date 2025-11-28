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
	e.Run("env var not set", func(e exam.E) {
		exam.ClearEnv(e, config.EnvPostgresPassword)
		exam.PanicWith(e, env, match.As[error](match.ErrorIs(config.ErrMissingRequiredEnvVar)), func() {
			config.New()
		})
	})
	e.Run("env var set", func(e exam.E) {
		const pwVal = "supersecretpassword"
		exam.SetEnv(e, config.EnvPostgresPassword, pwVal)
		cfg := config.New()
		exam.Equal(e, env, pwVal, cfg.PostresConfig.Password)
	})
}
