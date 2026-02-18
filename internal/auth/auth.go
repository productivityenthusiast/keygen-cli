package auth

import (
	"fmt"

	"github.com/productivityenthusiast/keygen-cli/internal/api"
	"github.com/productivityenthusiast/keygen-cli/internal/config"
)

func ResolveClient(cfg *config.Config) (*api.Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	client := api.NewClient(cfg.BaseURL, cfg.AccountID, cfg.Token)

	// If token is expired and we have credentials, try to refresh
	if cfg.IsTokenExpired() && cfg.Email != "" && cfg.Password != "" {
		token, err := client.CreateToken(cfg.Email, cfg.Password)
		if err != nil {
			return nil, fmt.Errorf("token expired and refresh failed: %w", err)
		}
		cfg.Token = token.Token
		cfg.TokenExp = token.Expiry
		_ = cfg.Save()
		client.Token = token.Token
	}

	return client, nil
}
