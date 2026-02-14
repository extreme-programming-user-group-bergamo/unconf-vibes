package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	defaultAPIEndpoint = "http://localhost:8080"
	defaultLogLevel    = "info"
	defaultConfigName  = ".unconf"
	defaultConfigType  = "yaml"
)

type LoadOptions struct {
	ConfigFile   string
	Overrides    map[string]any
	FlagSet      *pflag.FlagSet
	FlagBindings map[string]string
}

type Config struct {
	viper      *viper.Viper
	configFile string
}

func LoadConfig(ctx context.Context, opts LoadOptions) (*Config, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context canceled before loading config: %w", err)
	}

	v := viper.New()
	v.SetEnvPrefix("UNCONF")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	v.SetDefault("api_endpoint", defaultAPIEndpoint)
	v.SetDefault("log_level", defaultLogLevel)

	for configKey, flagName := range opts.FlagBindings {
		if opts.FlagSet == nil {
			return nil, fmt.Errorf("failed to bind flag %q for key %q: flagset is nil", flagName, configKey)
		}

		flag := opts.FlagSet.Lookup(flagName)
		if flag == nil {
			return nil, fmt.Errorf("failed to bind flag %q for key %q: flag not found", flagName, configKey)
		}

		if err := v.BindPFlag(configKey, flag); err != nil {
			return nil, fmt.Errorf("failed to bind flag %q for key %q: %w", flagName, configKey, err)
		}
	}

	if opts.ConfigFile != "" {
		v.SetConfigFile(opts.ConfigFile)
	} else {
		v.SetConfigName(defaultConfigName)
		v.SetConfigType(defaultConfigType)
		v.AddConfigPath(".")

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to resolve home directory for config discovery: %w", err)
		}
		v.AddConfigPath(homeDir)
	}

	if err := v.ReadInConfig(); err != nil {
		var cfgNotFound viper.ConfigFileNotFoundError
		if !errors.As(err, &cfgNotFound) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	for key, value := range opts.Overrides {
		v.Set(key, value)
	}

	configFileUsed := v.ConfigFileUsed()
	if opts.ConfigFile != "" && configFileUsed == "" {
		configFileUsed = opts.ConfigFile
	}

	return &Config{viper: v, configFile: configFileUsed}, nil
}

func (c *Config) GetAPIEndpoint() string {
	return c.viper.GetString("api_endpoint")
}

func (c *Config) GetLogLevel() string {
	return c.viper.GetString("log_level")
}

func (c *Config) GetConfigFile() string {
	if c.configFile != "" {
		return c.configFile
	}

	return c.viper.GetString("config_file")
}
