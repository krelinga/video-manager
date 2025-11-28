package config

import (
	"errors"
	"fmt"
	"os"
)

var ErrMissingRequiredEnvVar = errors.New("missing required environment variable")

const EnvPostgresPassword = "VIDEO_MANAGER_POSTGRES_PASSWORD"

type Config struct {
	DiscInboxDir string
	HttpPort     int
	RunDiscService bool
	PostresConfig  *Postres
}

type Postres struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

func New() *Config {
	return &Config{
		// TODO: make these configurable.
		DiscInboxDir: "/nas/media/video-manager/disc/inbox",
		HttpPort:     25009,
		RunDiscService: true,
		PostresConfig: &Postres{
			Host:     "nas-docker.i.krel.ing",
			Port:	 5432,
			User: "video_manager_prod",
			Password: getPostgresPassword(),
			DBName:   "video_manager_prod",
		},
	}
}

func getPostgresPassword() string {
	pw, ok := os.LookupEnv(EnvPostgresPassword)
	if !ok {
		panic(fmt.Errorf("%w: %s", ErrMissingRequiredEnvVar, EnvPostgresPassword))
	}
	return pw
}