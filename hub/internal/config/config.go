package config

type Config struct {
	Listen  *string `toml:"listen"`
	Storage *string `toml:"storage"`
	DSN     *string `toml:"dsn"`
}
