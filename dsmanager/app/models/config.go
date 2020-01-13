package models

// Config is a struct matching the configuration parameters that are stored in
// the 'config.toml' file.
type Config struct {
	CatalogID  string
	BCPath     string
	DarcID     string
	KeyID      string
	ConfigPath string
	PubKeyPath string
}
