package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Postgres struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Database string `json:"database"`
	} `json:"postgres"`
	Twitch struct {
		BotLogin    string   `json:"bot_login"`    // bot login name / username
		BotPassword string   `json:"bot_password"` // oauth token: https://twitchapps.com/tmi/
		JoinChannel string   `json:"join_channel"` // which IRC / Twitch channel the bot should join on start
		BotOwner    string   `json:"bot_owner"`    // actual bot owner / channel owner, usually matches JoinChannel
		Editors     []string `json:"editors"`      // not mods, users with access to internal bot commands
	} `json:"twitch"`
	// Since the bot needs access to user resources as well as api resources
	// we will need developer specific vars.
	// https://dev.twitch.tv/
	TwitchAuth struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		RedirectURL  string `json:"redirect_url"`
	} `json:"twitch_auth"`
	Command struct {
		Prefix          string `json:"prefix"`           // prefix to call commands from chat, like "!" or "."
		DefaultCooldown int    `json:"default_cooldown"` // usually 0 or < 5 seconds
	} `json:"command"`
}

// Instances new config from json file, but does not check for any missing keys or errors.
func LoadConfig(cfgPath string) (*Config, error) {
	f, err := os.Open(cfgPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config

	err = json.NewDecoder(f).Decode(&cfg)

	return &cfg, err
}

// Checks config for important or missing keys / values and returns error if missing.
func (c *Config) CheckConfig() error {
	if c.Postgres.Host == "" {
		return fmt.Errorf("missing key: postgres host")
	}

	if c.Postgres.Port == 0 {
		return fmt.Errorf("missing key: postgres port")
	}

	if c.Postgres.User == "" {
		return fmt.Errorf("missing key: postgres user")
	}

	if c.Postgres.Password == "" {
		return fmt.Errorf("missing key: postgres password")
	}

	if c.Postgres.Database == "" {
		return fmt.Errorf("missing key: postgres database")
	}

	if c.Postgres.Database == "" {
		return fmt.Errorf("missing key: postgres database")
	}

	if c.Twitch.BotLogin == "" {
		return fmt.Errorf("missing key: twitch bot login")
	}

	if c.Twitch.BotPassword == "" {
		return fmt.Errorf("missing key: twitch bot password")
	}

	if c.Twitch.JoinChannel == "" {
		return fmt.Errorf("missing key: twitch join channel")
	}

	if c.Twitch.BotOwner == "" {
		return fmt.Errorf("missing key: twitch bot owner")
	}

	if c.TwitchAuth.ClientID == "" {
		return fmt.Errorf("missing key: twitch auth client id")
	}

	if c.TwitchAuth.ClientSecret == "" {
		return fmt.Errorf("missing key: twitch auth client secret")
	}

	if c.TwitchAuth.RedirectURL == "" {
		return fmt.Errorf("missing key: twitch auth redirect url")
	}

	if c.Command.Prefix == "" {
		return fmt.Errorf("missing key: command prefix")
	}

	return nil
}
