package models

import (
	"errors"

	"github.com/BurntSushi/toml"
	"github.com/dedis/odyssey/dsmanager/app/helpers"
	"golang.org/x/xerrors"
)

// Config holds the global configuration of the server
type Config struct {
	*TOMLConfig
	TaskManager helpers.TaskManagerI
	Executor    helpers.Executor
	CloudClient helpers.CloudClient
	RunHTTP     helpers.RunHTTP
}

// TOMLConfig is a struct matching the configuration parameters that are stored
// in the 'config.toml' file.
type TOMLConfig struct {
	CatalogID  string
	BCPath     string
	DarcID     string
	KeyID      string
	ConfigPath string
	PubKeyPath string
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
		TaskManager: helpers.NewDefaultTaskManager(),
		Executor:    helpers.NewOSExecutor(),
		CloudClient: cloudClient,
		RunHTTP:     helpers.NewDefaultRunHTTP(),
	}
	return conf, nil
}
