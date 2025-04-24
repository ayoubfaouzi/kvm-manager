package config

import (
	"os"

	"github.com/spf13/viper"
)

// BrokerCfg represents the broker producer config.
type BrokerCfg struct {
	// the data source name (DSN) for connecting to the broker server.
	Address string `mapstructure:"address"`
	// Topic name to write to.
	Topic string `mapstructure:"topic"`
}

// Config represents our application config.
type Config struct {
	// The IP:Port. Defaults to 8080.
	Address string `mapstructure:"address"`
	// Log level. Defaults to info.
	LogLevel string `mapstructure:"log_level"`
	// Broker server configuration.
	Broker BrokerCfg `mapstructure:"nsq"`
}

// Load returns an application configuration which is populated
// from the given configuration file.
func Load(path string) (*Config, error) {

	// Create a new config.
	c := Config{}

	// Adding our TOML config file.
	viper.AddConfigPath(path)

	// Load the config type depending on env variable.
	var name string
	env := os.Getenv("LX_DEPLOYMENT_KIND")
	switch env {
	case "local":
		name = "local"
	case "dev":
		name = "dev"
	case "prod":
		name = "prod"
	default:
		name = "local"
	}

	// Set the config name to choose from the config path
	// Extension not needed.
	viper.SetConfigName(name)

	// Load the configuration from disk.
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	// Unmarshal the config into a struct.
	err = viper.Unmarshal(&c)
	if err != nil {
		return nil, err
	}

	return &c, err
}
