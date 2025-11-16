package config

type Config struct {
	DiscInboxDir string
	HttpPort     int
	RunDiscService bool
}

func New() *Config {
	return &Config{
		// TODO: make these configurable.
		DiscInboxDir: "/nas/media/video-manager/disc/inbox",
		HttpPort:     25009,
		RunDiscService: true,
	}
}
