package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	content := []byte(`
server:
  port: 8080
  host: "127.0.0.1"
database:
  path: "test.db"
jwt:
  access_secret: "test-access-secret-32chars-minimum!"
  refresh_secret: "test-refresh-secret-32chars-minimum!"
  access_expire: "30m"
  refresh_expire: "72h"
smtp:
  host: "smtp.test.com"
  port: 465
  username: "test@test.com"
  password: "test-password"
  from: "test@test.com"
  admin_email: "admin@test.com"
  use_tls: true
lsky:
  base_url: "https://lsky.test.com"
  email: "lsky@test.com"
  password: "lsky-password"
  token: ""
admin:
  email: "admin@test.com"
site:
  name: "测试站点"
  start_date: "2024-01-01"
upload:
  max_size_mb: 20
`)

	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write(content)
	require.NoError(t, err)
	tmpFile.Close()

	cfg, err := LoadConfig(tmpFile.Name())
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Server
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)

	// Database
	assert.Equal(t, "test.db", cfg.Database.Path)

	// JWT
	assert.Equal(t, "test-access-secret-32chars-minimum!", cfg.JWT.AccessSecret)
	assert.Equal(t, "test-refresh-secret-32chars-minimum!", cfg.JWT.RefreshSecret)
	assert.Equal(t, 30*time.Minute, cfg.JWT.AccessExpire)
	assert.Equal(t, 72*time.Hour, cfg.JWT.RefreshExpire)

	// SMTP
	assert.Equal(t, "smtp.test.com", cfg.SMTP.Host)
	assert.Equal(t, 465, cfg.SMTP.Port)
	assert.Equal(t, "test@test.com", cfg.SMTP.Username)
	assert.Equal(t, "test-password", cfg.SMTP.Password)
	assert.Equal(t, "test@test.com", cfg.SMTP.From)
	assert.Equal(t, "admin@test.com", cfg.SMTP.AdminEmail)
	assert.True(t, cfg.SMTP.UseTLS)

	// Lsky
	assert.Equal(t, "https://lsky.test.com", cfg.Lsky.BaseURL)
	assert.Equal(t, "lsky@test.com", cfg.Lsky.Email)
	assert.Equal(t, "lsky-password", cfg.Lsky.Password)
	assert.Equal(t, "", cfg.Lsky.Token)

	// Admin
	assert.Equal(t, "admin@test.com", cfg.Admin.Email)

	// Site
	assert.Equal(t, "测试站点", cfg.Site.Name)
	assert.Equal(t, "2024-01-01", cfg.Site.StartDate)

	// Upload
	assert.Equal(t, 20, cfg.Upload.MaxSizeMB)
}

func TestDefaults(t *testing.T) {
	// Create a minimal config with one override to verify defaults fill in
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("server:\n  port: 9999\n")
	require.NoError(t, err)
	tmpFile.Close()

	cfg, err := LoadConfig(tmpFile.Name())
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Overridden value
	assert.Equal(t, 9999, cfg.Server.Port)

	// Default values
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, "gm_site.db", cfg.Database.Path)
	assert.Equal(t, 10, cfg.Upload.MaxSizeMB)
	assert.Equal(t, "顾夏", cfg.Site.Name)
	assert.Equal(t, 15*time.Minute, cfg.JWT.AccessExpire)
	assert.Equal(t, 168*time.Hour, cfg.JWT.RefreshExpire)
	assert.Equal(t, 587, cfg.SMTP.Port)
	assert.True(t, cfg.SMTP.UseTLS)
}

func TestEnvOverride(t *testing.T) {
	// Set env vars before loading config
	require.NoError(t, os.Setenv("GM_SMTP_PASSWORD", "env-smtp-pass"))
	require.NoError(t, os.Setenv("GM_JWT_ACCESS_SECRET", "env-access-secret-32chars-minimum!!"))
	require.NoError(t, os.Setenv("GM_JWT_REFRESH_SECRET", "env-refresh-secret-32chars-minimum!"))
	require.NoError(t, os.Setenv("GM_LSKY_PASSWORD", "env-lsky-pass"))
	t.Cleanup(func() {
		os.Unsetenv("GM_SMTP_PASSWORD")
		os.Unsetenv("GM_JWT_ACCESS_SECRET")
		os.Unsetenv("GM_JWT_REFRESH_SECRET")
		os.Unsetenv("GM_LSKY_PASSWORD")
	})

	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Write config with different values to verify env override
	_, err = tmpFile.WriteString(`
smtp:
  password: "file-smtp-pass"
jwt:
  access_secret: "file-access-secret-32chars-minimum!!!"
  refresh_secret: "file-refresh-secret-32chars-minimum!!"
lsky:
  password: "file-lsky-pass"
`)
	require.NoError(t, err)
	tmpFile.Close()

	cfg, err := LoadConfig(tmpFile.Name())
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Env vars should override YAML file values
	assert.Equal(t, "env-smtp-pass", cfg.SMTP.Password)
	assert.Equal(t, "env-access-secret-32chars-minimum!!", cfg.JWT.AccessSecret)
	assert.Equal(t, "env-refresh-secret-32chars-minimum!", cfg.JWT.RefreshSecret)
	assert.Equal(t, "env-lsky-pass", cfg.Lsky.Password)
}

func TestLoadConfigNoFile(t *testing.T) {
	// LoadConfig with a non-existent path should return defaults without error
	cfg, err := LoadConfig("non-existent-file.yaml")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Should have defaults
	assert.Equal(t, 1323, cfg.Server.Port)
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, "gm_site.db", cfg.Database.Path)
	assert.Equal(t, 10, cfg.Upload.MaxSizeMB)
}
