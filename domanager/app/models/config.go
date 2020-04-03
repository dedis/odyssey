package models

import "github.com/dedis/odyssey/dsmanager/app/helpers"

// Config holds the global configuration of the server
type Config struct {
	*TOMLConfig
	TaskFactory helpers.TaskFactoryI
}

// TOMLConfig is a struct matching the configuration parameters that are stored in
// the 'config.toml' file.
type TOMLConfig struct {
	ConfigPath string
	CatalogID  string
	LtsID      string
	LtsKey     string
	// Standalone mode means that the domanager will take responsibility itself
	// for making the Darc that controls access to the write instance for new
	// data sets. When Standalone is false (the default), domanager talks to the
	// Enclave manager to create the Darc.
	Standalone bool
}
