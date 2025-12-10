package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
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
	EnvRootDir			= "VIDEO_MANAGER_ROOT_DIR"
)

type Config struct {
	Paths        Paths
	HttpPort     int
	Postgres     *Postgres
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
		Paths: Paths{
			RootDir: getRequiredVar(EnvRootDir),
		},
		HttpPort:     25009,
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

type PathKind bool

const (
	PathKindAbsolute PathKind = true
	PathKindRelative PathKind = false
)

type Paths struct {
	RootDir string
}

// Makes sure that all necessary directories exist.
func (p Paths) Bootstrap() error {
	necessary := []string{
		p.InboxDvd(PathKindAbsolute),
		p.MediaDvd(PathKindAbsolute),
	}
	for _, dir := range necessary {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

func (p Paths) makePath(pk PathKind, parts ...string) string {
	if pk == PathKindAbsolute {
		parts = append([]string{p.RootDir}, parts...)
	}
	return filepath.Join(parts...)
}

// Returns the path to the inbox dvd directory.
// Entries under this directory represent DVDs that have not been imported yet.
func (p Paths) InboxDvd(pk PathKind) string {
	return p.makePath(pk, "inbox", "dvd")
}

// Returns the path to the media directory that contains all imported DVDs.
func (p Paths) MediaDvd(pk PathKind) string {
	return p.makePath(pk, "media", "dvd")
}

// Returns the path to the directory that contains a specific imported DVD.
func (p Paths) MediaDvdId(pk PathKind, mediaId uint32) string {
	return p.makePath(pk, "media", "dvd", fmt.Sprintf("%d", mediaId))
}
