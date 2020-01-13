package models

// Config is a struct matching the configuration parameters that are stored in
// the 'config.toml' file.
type Config struct {
	ConfigPath string
	CatalogID  string
	BCPath     string
	LtsID      string
	LtsKey     string
}
