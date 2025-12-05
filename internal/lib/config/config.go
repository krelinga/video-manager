package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
)

var (
	ErrMissingRequiredEnvVar = errors.New("missing required environment variable")
	ErrMalformedEnvVar       = errors.New("malformed environment variable")
)

const (
	EnvPostgresHost     = "VIDEO_MANAGER_POSTGRES_HOST"
	EnvPostgresPort     = "VIDEO_MANAGER_POSTGRES_PORT"
	EnvPostgresDBName   = "VIDEO_MANAGER_POSTGRES_DBNAME"
	EnvPostgresUser     = "VIDEO_MANAGER_POSTGRES_USER"
	EnvPostgresPassword = "VIDEO_MANAGER_POSTGRES_PASSWORD"
)

type Config struct {
	DiscInboxDir      string
	HttpPort          int
	Postgres          *Postgres
}

type Postgres struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

func (p *Postgres) URL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		url.QueryEscape(p.User),
		url.QueryEscape(p.Password),
		url.QueryEscape(p.Host),
		p.Port,
		url.QueryEscape(p.DBName),
	)
}

func New() *Config {
	return &Config{
		// TODO: make these configurable.
		DiscInboxDir:      "/nas/media/video-manager/disc/inbox",
		HttpPort:          25009,
		Postgres: &Postgres{
			Host:     getRequiredVar(EnvPostgresHost),
			Port:     parseInt(getVarWithDefault(EnvPostgresPort, "5432")),
			User:     getRequiredVar(EnvPostgresUser),
			Password: getRequiredVar(EnvPostgresPassword),
			DBName:   getRequiredVar(EnvPostgresDBName),
		},
	}
}

func getRequiredVar(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		panic(fmt.Errorf("%w: %s", ErrMissingRequiredEnvVar, key))
	}
	return val
}

func getVarWithDefault(key, defaultVal string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return defaultVal
	}
	return val
}

func parseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(fmt.Errorf("%w: could not parse %q as int", ErrMalformedEnvVar, s))
	}

	return i
}
