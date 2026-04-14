package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	defaultAPIEndpoint   = "http://localhost:8080"
	defaultLogLevel      = "info"
	defaultDBPath        = ".unconf.db"
	defaultDBMaxOpen     = 25
	defaultDBMaxIdle     = 25
	defaultDBBusyMS      = 5000
	defaultTokenTTL      = 24 * time.Hour
	defaultRefreshTTL    = 7 * 24 * time.Hour
	defaultConfigName    = ".unconf"
	defaultConfigType    = "yaml"
	defaultEmailProvider = "smtp"
	defaultSMTPPort      = 1025
	defaultSendGridURL   = "https://api.sendgrid.com/v3/mail/send"
	defaultMailgunURL    = "https://api.mailgun.net/v3"
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
	v.SetDefault("db_path", defaultDBPath)
	v.SetDefault("db_max_open_conns", defaultDBMaxOpen)
	v.SetDefault("db_max_idle_conns", defaultDBMaxIdle)
	v.SetDefault("db_busy_timeout_ms", defaultDBBusyMS)
	v.SetDefault("token_ttl", defaultTokenTTL.String())
	v.SetDefault("refresh_ttl", defaultRefreshTTL.String())
	v.SetDefault("email_provider", defaultEmailProvider)
	v.SetDefault("smtp_port", defaultSMTPPort)
	v.SetDefault("sendgrid_base_url", defaultSendGridURL)
	v.SetDefault("mailgun_base_url", defaultMailgunURL)

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

func (c *Config) GetDBPath() string {
	return c.viper.GetString("db_path")
}

func (c *Config) GetDBMaxOpenConns() int {
	return c.viper.GetInt("db_max_open_conns")
}

func (c *Config) GetDBMaxIdleConns() int {
	return c.viper.GetInt("db_max_idle_conns")
}

func (c *Config) GetDBBusyTimeoutMS() int {
	return c.viper.GetInt("db_busy_timeout_ms")
}

func (c *Config) GetConfigFile() string {
	if c.configFile != "" {
		return c.configFile
	}

	return c.viper.GetString("config_file")
}

func (c *Config) GetGitHubClientID() string {
	return c.viper.GetString("github_client_id")
}

func (c *Config) GetGitHubClientSecret() string {
	return c.viper.GetString("github_client_secret")
}

func (c *Config) GetPasetoSymmetricKey() string {
	return c.viper.GetString("paseto_symmetric_key")
}

func (c *Config) GetTokenTTL() time.Duration {
	ttl, err := time.ParseDuration(c.viper.GetString("token_ttl"))
	if err != nil || ttl <= 0 {
		return defaultTokenTTL
	}

	return ttl
}

func (c *Config) GetRefreshTTL() time.Duration {
	ttl, err := time.ParseDuration(c.viper.GetString("refresh_ttl"))
	if err != nil || ttl <= 0 {
		return defaultRefreshTTL
	}

	return ttl
}

func (c *Config) GetEmailProvider() string {
	return c.viper.GetString("email_provider")
}

func (c *Config) GetEmailFromAddress() string {
	return c.viper.GetString("email_from_address")
}

func (c *Config) GetEmailFromName() string {
	return c.viper.GetString("email_from_name")
}

func (c *Config) GetSMTPHost() string {
	return c.viper.GetString("smtp_host")
}

func (c *Config) GetSMTPPort() int {
	return c.viper.GetInt("smtp_port")
}

func (c *Config) GetSMTPUsername() string {
	return c.viper.GetString("smtp_username")
}

func (c *Config) GetSMTPPassword() string {
	return c.viper.GetString("smtp_password")
}

func (c *Config) GetSMTPUseTLS() bool {
	return c.viper.GetBool("smtp_use_tls")
}

func (c *Config) GetSendGridAPIKey() string {
	return c.viper.GetString("sendgrid_api_key")
}

func (c *Config) GetSendGridBaseURL() string {
	return c.viper.GetString("sendgrid_base_url")
}

func (c *Config) GetMailgunAPIKey() string {
	return c.viper.GetString("mailgun_api_key")
}

func (c *Config) GetMailgunDomain() string {
	return c.viper.GetString("mailgun_domain")
}

func (c *Config) GetMailgunBaseURL() string {
	return c.viper.GetString("mailgun_base_url")
}
