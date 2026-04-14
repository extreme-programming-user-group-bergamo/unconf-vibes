package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfigFromExplicitFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "custom.yaml")

	err := os.WriteFile(configPath, []byte("api_endpoint: https://from-file.example\nlog_level: warn\n"), 0o600)
	require.NoError(t, err)

	cfg, err := LoadConfig(context.Background(), LoadOptions{ConfigFile: configPath})
	require.NoError(t, err)

	assert.Equal(t, "https://from-file.example", cfg.GetAPIEndpoint())
	assert.Equal(t, "warn", cfg.GetLogLevel())
	assert.Equal(t, configPath, cfg.GetConfigFile())
}

func TestLoadConfigFromEnvironment(t *testing.T) {
	t.Setenv("UNCONF_API_ENDPOINT", "https://from-env.example")
	t.Setenv("UNCONF_LOG_LEVEL", "debug")
	t.Setenv("UNCONF_DB_PATH", "./from-env.db")
	t.Setenv("UNCONF_DB_MAX_OPEN_CONNS", "12")
	t.Setenv("UNCONF_DB_MAX_IDLE_CONNS", "6")
	t.Setenv("UNCONF_DB_BUSY_TIMEOUT_MS", "7500")

	cfg, err := LoadConfig(context.Background(), LoadOptions{})
	require.NoError(t, err)

	assert.Equal(t, "https://from-env.example", cfg.GetAPIEndpoint())
	assert.Equal(t, "debug", cfg.GetLogLevel())
	assert.Equal(t, "./from-env.db", cfg.GetDBPath())
	assert.Equal(t, 12, cfg.GetDBMaxOpenConns())
	assert.Equal(t, 6, cfg.GetDBMaxIdleConns())
	assert.Equal(t, 7500, cfg.GetDBBusyTimeoutMS())
}

func TestLoadConfigAuthSettingsFromEnvironment(t *testing.T) {
	t.Setenv("UNCONF_GITHUB_CLIENT_ID", "gh-client-id")
	t.Setenv("UNCONF_GITHUB_CLIENT_SECRET", "gh-client-secret")
	t.Setenv("UNCONF_PASETO_SYMMETRIC_KEY", "0123456789abcdef0123456789abcdef")
	t.Setenv("UNCONF_TOKEN_TTL", "12h")
	t.Setenv("UNCONF_REFRESH_TTL", "72h")

	cfg, err := LoadConfig(context.Background(), LoadOptions{})
	require.NoError(t, err)

	assert.Equal(t, "gh-client-id", cfg.GetGitHubClientID())
	assert.Equal(t, "gh-client-secret", cfg.GetGitHubClientSecret())
	assert.Equal(t, "0123456789abcdef0123456789abcdef", cfg.GetPasetoSymmetricKey())
	assert.Equal(t, 12*time.Hour, cfg.GetTokenTTL())
	assert.Equal(t, 72*time.Hour, cfg.GetRefreshTTL())
}

func TestLoadConfigOverridesTakePriority(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "custom.yaml")
	err := os.WriteFile(configPath, []byte("api_endpoint: https://from-file.example\n"), 0o600)
	require.NoError(t, err)

	t.Setenv("UNCONF_API_ENDPOINT", "https://from-env.example")

	cfg, err := LoadConfig(context.Background(), LoadOptions{
		ConfigFile: configPath,
		Overrides: map[string]any{
			"api_endpoint": "https://from-override.example",
		},
	})
	require.NoError(t, err)

	assert.Equal(t, "https://from-override.example", cfg.GetAPIEndpoint())
}

func TestLoadConfigDefaultsApplied(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	cfg, err := LoadConfig(context.Background(), LoadOptions{})
	require.NoError(t, err)

	assert.Equal(t, defaultAPIEndpoint, cfg.GetAPIEndpoint())
	assert.Equal(t, defaultLogLevel, cfg.GetLogLevel())
	assert.Equal(t, defaultDBPath, cfg.GetDBPath())
	assert.Equal(t, defaultDBMaxOpen, cfg.GetDBMaxOpenConns())
	assert.Equal(t, defaultDBMaxIdle, cfg.GetDBMaxIdleConns())
	assert.Equal(t, defaultDBBusyMS, cfg.GetDBBusyTimeoutMS())
	assert.Equal(t, defaultTokenTTL, cfg.GetTokenTTL())
	assert.Equal(t, defaultRefreshTTL, cfg.GetRefreshTTL())
	assert.Equal(t, defaultEmailProvider, cfg.GetEmailProvider())
	assert.Equal(t, defaultSMTPPort, cfg.GetSMTPPort())
	assert.Equal(t, defaultSendGridURL, cfg.GetSendGridBaseURL())
	assert.Equal(t, defaultMailgunURL, cfg.GetMailgunBaseURL())
}

func TestLoadConfigAutoDiscoversHomeFile(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	homeCfgPath := filepath.Join(tempHome, ".unconf.yaml")
	err := os.WriteFile(homeCfgPath, []byte("api_endpoint: https://from-home.example\n"), 0o600)
	require.NoError(t, err)

	cfg, err := LoadConfig(context.Background(), LoadOptions{})
	require.NoError(t, err)

	assert.Equal(t, "https://from-home.example", cfg.GetAPIEndpoint())
	assert.Equal(t, homeCfgPath, cfg.GetConfigFile())
}

func TestLoadConfigEmailSettingsFromEnvironment(t *testing.T) {
	t.Setenv("UNCONF_EMAIL_PROVIDER", "mailgun")
	t.Setenv("UNCONF_EMAIL_FROM_ADDRESS", "noreply@unconf.dev")
	t.Setenv("UNCONF_EMAIL_FROM_NAME", "UNCONF")
	t.Setenv("UNCONF_SMTP_HOST", "localhost")
	t.Setenv("UNCONF_SMTP_PORT", "1025")
	t.Setenv("UNCONF_SMTP_USERNAME", "smtp-user")
	t.Setenv("UNCONF_SMTP_PASSWORD", "smtp-pass")
	t.Setenv("UNCONF_SMTP_USE_TLS", "true")
	t.Setenv("UNCONF_SENDGRID_API_KEY", "sg-key")
	t.Setenv("UNCONF_SENDGRID_BASE_URL", "https://sendgrid.example.local/send")
	t.Setenv("UNCONF_MAILGUN_API_KEY", "mg-key")
	t.Setenv("UNCONF_MAILGUN_DOMAIN", "mg.example.com")
	t.Setenv("UNCONF_MAILGUN_BASE_URL", "https://mailgun.example.local/v3")

	cfg, err := LoadConfig(context.Background(), LoadOptions{})
	require.NoError(t, err)

	assert.Equal(t, "mailgun", cfg.GetEmailProvider())
	assert.Equal(t, "noreply@unconf.dev", cfg.GetEmailFromAddress())
	assert.Equal(t, "UNCONF", cfg.GetEmailFromName())
	assert.Equal(t, "localhost", cfg.GetSMTPHost())
	assert.Equal(t, 1025, cfg.GetSMTPPort())
	assert.Equal(t, "smtp-user", cfg.GetSMTPUsername())
	assert.Equal(t, "smtp-pass", cfg.GetSMTPPassword())
	assert.True(t, cfg.GetSMTPUseTLS())
	assert.Equal(t, "sg-key", cfg.GetSendGridAPIKey())
	assert.Equal(t, "https://sendgrid.example.local/send", cfg.GetSendGridBaseURL())
	assert.Equal(t, "mg-key", cfg.GetMailgunAPIKey())
	assert.Equal(t, "mg.example.com", cfg.GetMailgunDomain())
	assert.Equal(t, "https://mailgun.example.local/v3", cfg.GetMailgunBaseURL())
}
