package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/joho/godotenv"
)

// Config holds credentials for a single profile.
type Config struct {
	AccountID   string `json:"account_id"`
	BaseURL     string `json:"base_url"`
	Token       string `json:"token"`
	Email       string `json:"email,omitempty"`
	Password    string `json:"password,omitempty"`
	TokenExp    string `json:"token_expiry,omitempty"`
	ProfileName string `json:"-"` // runtime-only, not persisted inside the profile
}

// ProfilesConfig is the top-level structure stored in profiles.json.
type ProfilesConfig struct {
	DefaultProfile string             `json:"default_profile"`
	Profiles       map[string]*Config `json:"profiles"`
}

func configDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".keygen-cli")
}

func profilesPath() string {
	return filepath.Join(configDir(), "profiles.json")
}

// legacyConfigPath returns the old flat config.json path (used for migration).
func legacyConfigPath() string {
	return filepath.Join(configDir(), "config.json")
}

// loadProfiles reads the profiles.json file. If it doesn't exist but a legacy
// config.json does, it migrates the old config into a "default" profile.
func loadProfiles() *ProfilesConfig {
	pc := &ProfilesConfig{
		DefaultProfile: "default",
		Profiles:       make(map[string]*Config),
	}

	if data, err := os.ReadFile(profilesPath()); err == nil {
		_ = json.Unmarshal(data, pc)
		if pc.Profiles == nil {
			pc.Profiles = make(map[string]*Config)
		}
		return pc
	}

	// Migrate legacy config.json → profiles.json
	if data, err := os.ReadFile(legacyConfigPath()); err == nil {
		legacy := &Config{}
		if json.Unmarshal(data, legacy) == nil && (legacy.AccountID != "" || legacy.Token != "") {
			pc.Profiles["default"] = legacy
			_ = saveProfiles(pc)
			// Keep old file around (non-destructive migration)
		}
	}

	return pc
}

func saveProfiles(pc *ProfilesConfig) error {
	dir := configDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	data, err := json.MarshalIndent(pc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling profiles: %w", err)
	}
	return os.WriteFile(profilesPath(), data, 0600)
}

// LoadProfile loads a named profile. If name is empty, the default profile is
// used. Environment variables and .env files are layered on top.
func LoadProfile(name string) *Config {
	// Load .env from cwd first, then fallback to global
	if err := godotenv.Load(); err != nil {
		globalEnv := filepath.Join(configDir(), ".env")
		_ = godotenv.Load(globalEnv)
	}

	pc := loadProfiles()

	if name == "" {
		name = pc.DefaultProfile
	}
	if name == "" {
		name = "default"
	}

	cfg := &Config{ProfileName: name}
	if p, ok := pc.Profiles[name]; ok && p != nil {
		*cfg = *p
		cfg.ProfileName = name
	}

	// Environment variables override saved config
	applyEnvOverrides(cfg)

	return cfg
}

// Load loads the default profile (backward-compatible wrapper).
func Load() *Config {
	return LoadProfile("")
}

func applyEnvOverrides(cfg *Config) {
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
}

// Save persists this config to its named profile in profiles.json.
func (c *Config) Save() error {
	pc := loadProfiles()
	name := c.ProfileName
	if name == "" {
		name = pc.DefaultProfile
	}
	if name == "" {
		name = "default"
	}

	// Store a copy without the runtime ProfileName field
	saved := *c
	saved.ProfileName = ""
	pc.Profiles[name] = &saved

	return saveProfiles(pc)
}

// Clear removes this profile from profiles.json.
func (c *Config) Clear() error {
	pc := loadProfiles()
	name := c.ProfileName
	if name == "" {
		name = pc.DefaultProfile
	}
	if name == "" {
		name = "default"
	}
	delete(pc.Profiles, name)
	return saveProfiles(pc)
}

// SaveProfile saves a config under the given profile name.
func SaveProfile(name string, cfg *Config) error {
	pc := loadProfiles()
	saved := *cfg
	saved.ProfileName = ""
	pc.Profiles[name] = &saved
	return saveProfiles(pc)
}

// DeleteProfile removes a named profile.
func DeleteProfile(name string) error {
	pc := loadProfiles()
	if _, ok := pc.Profiles[name]; !ok {
		return fmt.Errorf("profile %q not found", name)
	}
	delete(pc.Profiles, name)
	if pc.DefaultProfile == name {
		pc.DefaultProfile = "default"
	}
	return saveProfiles(pc)
}

// ListProfiles returns all profile names, sorted.
func ListProfiles() ([]string, string) {
	pc := loadProfiles()
	names := make([]string, 0, len(pc.Profiles))
	for n := range pc.Profiles {
		names = append(names, n)
	}
	sort.Strings(names)
	return names, pc.DefaultProfile
}

// GetProfile returns a specific profile's config (without env overrides).
func GetProfile(name string) (*Config, error) {
	pc := loadProfiles()
	p, ok := pc.Profiles[name]
	if !ok {
		return nil, fmt.Errorf("profile %q not found", name)
	}
	cfg := *p
	cfg.ProfileName = name
	return &cfg, nil
}

// SetDefaultProfile sets which profile is used when --profile is not specified.
func SetDefaultProfile(name string) error {
	pc := loadProfiles()
	if _, ok := pc.Profiles[name]; !ok {
		return fmt.Errorf("profile %q not found", name)
	}
	pc.DefaultProfile = name
	return saveProfiles(pc)
}

// GetDefaultProfileName returns the current default profile name.
func GetDefaultProfileName() string {
	pc := loadProfiles()
	if pc.DefaultProfile == "" {
		return "default"
	}
	return pc.DefaultProfile
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
		return fmt.Errorf("account ID not configured (set KEYGEN_ACCOUNT_ID or run keygen login --profile %s)", c.ProfileName)
	}
	if c.BaseURL == "" {
		return fmt.Errorf("base URL not configured (set KEYGEN_BASE_URL or run keygen login --profile %s)", c.ProfileName)
	}
	if c.Token == "" {
		return fmt.Errorf("token not configured (set KEYGEN_TOKEN or run keygen login --profile %s)", c.ProfileName)
	}
	return nil
}
