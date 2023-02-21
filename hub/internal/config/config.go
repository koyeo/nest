package config

type Config struct {
	APIListen       *string    `toml:"api_listen"`
	PublisherListen *string    `toml:"publisher_listen"`
	DSN             *string    `toml:"dsn"`
	Publisher       *Publisher `toml:"publisher"`
}

type Publisher struct {
	Storage *string `toml:"storage"`
}
