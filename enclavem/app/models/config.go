package models

import (
	"errors"

	"github.com/BurntSushi/toml"
	"github.com/dedis/odyssey/dsmanager/app/helpers"
	"golang.org/x/xerrors"
)

// Config is a struct matching the configuration parameters that are stored in
// the 'config.toml' file.
type Config struct {
	*TOMLConfig
	Executor    helpers.Executor
	RunHTTP     helpers.RunHTTP
	CloudClient helpers.CloudClient
}

// TOMLConfig is a struct matching the configuration parameters that are stored
// in the 'config.toml' file.
type TOMLConfig struct {
	BCPath     string
	DarcID     string
	KeyID      string
	ConfigPath string
	NetworkID  string
}

// NewConfig creates a new Config
func NewConfig() (*Config, error) {
	tomlConf := &TOMLConfig{}
	_, err := toml.DecodeFile("config.toml", tomlConf)
	if err != nil {
		return nil, errors.New("failed to read config: " + err.Error())
	}

	cloudClient, err := helpers.NewMinioCloudClient()
	if err != nil {
		return nil, xerrors.Errorf("failed to create minion cloud client: %v", err)
	}

	conf := &Config{TOMLConfig: tomlConf,
		Executor:    helpers.NewOSExecutor(),
		CloudClient: cloudClient,
		RunHTTP:     helpers.NewDefaultRunHTTP(),
	}
	return conf, nil
}
