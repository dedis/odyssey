package models

// Config is a struct matching the configuration parameters that are stored in
// the 'config.toml' file.
type Config struct {
	BCPath     string
	DarcID     string
	KeyID      string
	ConfigPath string
	NetworkID  string
}
