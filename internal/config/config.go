package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AccountID string `json:"account_id"`
	BaseURL   string `json:"base_url"`
	Token     string `json:"token"`
	Email     string `json:"email,omitempty"`
	Password  string `json:"password,omitempty"`
	TokenExp  string `json:"token_expiry,omitempty"`
}

func configDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".keygen-cli")
}

func configPath() string {
	return filepath.Join(configDir(), "config.json")
}

func Load() *Config {
	// Load .env from cwd first, then fallback to global
	if err := godotenv.Load(); err != nil {
		globalEnv := filepath.Join(configDir(), ".env")
		_ = godotenv.Load(globalEnv)
	}

	cfg := &Config{}

	// Load saved config file
	if data, err := os.ReadFile(configPath()); err == nil {
		_ = json.Unmarshal(data, cfg)
	}

	// Environment variables override saved config
	if v := os.Getenv("KEYGEN_ACCOUNT_ID"); v != "" {
		cfg.AccountID = v
	}
	if v := os.Getenv("KEYGEN_BASE_URL"); v != "" {
		cfg.BaseURL = v
	}
	if v := os.Getenv("KEYGEN_TOKEN"); v != "" {
		cfg.Token = v
	} else if v := os.Getenv("KEYGEN_API_TOKEN"); v != "" {
		cfg.Token = v
	}
	if v := os.Getenv("KEYGEN_EMAIL"); v != "" {
		cfg.Email = v
	} else if v := os.Getenv("KEYGEN_ACCOUNT_EMAIL"); v != "" {
		cfg.Email = v
	}
	if v := os.Getenv("KEYGEN_PASSWORD"); v != "" {
		cfg.Password = v
	} else if v := os.Getenv("KEYGEN_ACCOUNT_PASSWORD"); v != "" {
		cfg.Password = v
	}

	return cfg
}

func (c *Config) Save() error {
	dir := configDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling config: %w", err)
	}

	return os.WriteFile(configPath(), data, 0600)
}

func (c *Config) Clear() error {
	return os.Remove(configPath())
}

func (c *Config) IsTokenExpired() bool {
	if c.TokenExp == "" {
		return false
	}
	t, err := time.Parse(time.RFC3339, c.TokenExp)
	if err != nil {
		return false
	}
	return time.Now().After(t)
}

func (c *Config) Validate() error {
	if c.AccountID == "" {
		return fmt.Errorf("account ID not configured (set KEYGEN_ACCOUNT_ID or run keygen login)")
	}
	if c.BaseURL == "" {
		return fmt.Errorf("base URL not configured (set KEYGEN_BASE_URL or run keygen login)")
	}
	if c.Token == "" {
		return fmt.Errorf("token not configured (set KEYGEN_TOKEN or run keygen login)")
	}
	return nil
}
