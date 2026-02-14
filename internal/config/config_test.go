package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

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

	cfg, err := LoadConfig(context.Background(), LoadOptions{})
	require.NoError(t, err)

	assert.Equal(t, "https://from-env.example", cfg.GetAPIEndpoint())
	assert.Equal(t, "debug", cfg.GetLogLevel())
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
