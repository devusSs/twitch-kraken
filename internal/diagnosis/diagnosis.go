package diagnosis

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/devusSs/twitch-kraken/internal/logging"
	"github.com/devusSs/twitch-kraken/internal/system"
	_ "github.com/lib/pq"
)

func RunDiagnosis(logPath, cfgPath string) (int, error) {
	errCount := 0

	printInfo("Running app in diagnostics mode...")

	printInfo("Loading config from file...")
	cfg, err := loadConfig(cfgPath)
	if err != nil {
		errCount++
		printError(fmt.Sprintf("Error loading config: %s", err.Error()))
	}

	printInfo("Checking config...")
	if err := cfg.checkConfig(); err != nil {
		errCount++
		printError(fmt.Sprintf("Error checking config: %s", err.Error()))
	}

	printInfo("Connecting to Postgres database...")
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Postgres.User, cfg.Postgres.Password,
		cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.Database)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		errCount++
		printError(fmt.Sprintf("Error creating Postgres connection: %s", err.Error()))
	}

	printInfo("Pinging database...")
	if err := db.Ping(); err != nil {
		errCount++
		printError(fmt.Sprintf("Error pinging database: %s", err.Error()))
	}
	defer db.Close()

	// Check the OS for unsupported versions / platforms.
	printInfo("Determining OS platform and version...")

	osV := system.DetermineOS()
	if osV != "linux" {
		errCount++
		printError(fmt.Sprintf("Determined OS \"%s\" may be unsupported (does not match recommended OS Linux)", osV))
	}

	// Check the network's connection to Twitch.
	printInfo("Testing connection to Twitch...")

	avg, err := system.TestConnection()
	if err != nil {
		return errCount, err
	}
	if avg > 500 {
		errCount++
		printError(fmt.Sprintf("Average ping exceeds 500 ms (%.0f ms). Program calls may be delayed", avg))
	}

	// Check the error.log file for information.
	printInfo("Checking error.log file for information...")

	foundErrsLogFile, err := logging.CheckErrorLogs(logPath)
	if err != nil {
		errCount++
		return errCount, err
	}
	if foundErrsLogFile != "" {
		errCount++
		printError(foundErrsLogFile)
	}

	return errCount, nil
}

func printInfo(message string) {
	fmt.Printf("[%s] %s\n", logging.InfoSign, message)
}

func printError(message string) {
	fmt.Printf("[%s] %s\n", logging.ErrorSign, message)
}

type config struct {
	Postgres struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Database string `json:"database"`
	} `json:"postgres"`
	Twitch struct {
		BotLogin    string   `json:"bot_login"`
		BotPassword string   `json:"bot_password"`
		JoinChannel string   `json:"join_channel"`
		BotOwner    string   `json:"bot_owner"`
		Editors     []string `json:"editors"`
	} `json:"twitch"`
	Command struct {
		Prefix          string `json:"prefix"`
		DefaultCooldown int    `json:"default_cooldown"`
	} `json:"command"`
}

func loadConfig(cfgPath string) (*config, error) {
	f, err := os.Open(cfgPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg config

	err = json.NewDecoder(f).Decode(&cfg)

	return &cfg, err
}

func (c *config) checkConfig() error {
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

	if c.Command.Prefix == "" {
		return fmt.Errorf("missing key: command prefix")
	}

	return nil
}
