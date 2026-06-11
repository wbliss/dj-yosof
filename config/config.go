// Package config loads the bot configuration from a YAML file, mirroring the
// original Python settings.py + config.yaml setup.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds all runtime settings for the bot.
type Config struct {
	// DiscordToken is the Discord bot token.
	DiscordToken string `yaml:"discord_token"`
	// GuildIDs are the guilds where slash commands are registered.
	GuildIDs []string `yaml:"guild_ids"`
	// SpotifyCredentialsFile is where the cached Spotify stored-credentials blob
	// is read from / written to. On first run an interactive OAuth2 login is
	// performed and the resulting credentials are persisted here.
	SpotifyCredentialsFile string `yaml:"spotify_credentials_file"`
	// SpotifyOAuthCallbackPort is the local port used for the OAuth2 redirect
	// during the first-time interactive Spotify login. 0 lets the OS choose.
	SpotifyOAuthCallbackPort int `yaml:"spotify_oauth_callback_port"`
	// SpotifyClientID and SpotifyClientSecret are a Spotify developer app's
	// credentials, used to authenticate Web API search/metadata requests via the
	// Client Credentials flow. Spotify rejects Web API requests made with the
	// go-librespot session (desktop client-ID) token with persistent 429s
	// (go-librespot#282); set these to avoid that. Leave empty to fall back to
	// the session token.
	SpotifyClientID     string `yaml:"spotify_client_id"`
	SpotifyClientSecret string `yaml:"spotify_client_secret"`
	// Debug enables verbose disgo gateway/voice and DAVE logging, useful for
	// diagnosing voice connection problems.
	Debug bool `yaml:"debug"`
}

// Load reads and parses the YAML config at path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config %q: %w", path, err)
	}

	if cfg.DiscordToken == "" {
		return nil, fmt.Errorf("config: discord_token is required")
	}
	if cfg.SpotifyCredentialsFile == "" {
		cfg.SpotifyCredentialsFile = "spotify_credentials.json"
	}

	return &cfg, nil
}
