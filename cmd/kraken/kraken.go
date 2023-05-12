package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/devusSs/twitch-kraken/internal/auth"
	"github.com/devusSs/twitch-kraken/internal/auth/authtwitch"
	"github.com/devusSs/twitch-kraken/internal/bot"
	"github.com/devusSs/twitch-kraken/internal/bot/gatekeeper"
	"github.com/devusSs/twitch-kraken/internal/config"
	"github.com/devusSs/twitch-kraken/internal/database/postgres"
	"github.com/devusSs/twitch-kraken/internal/diagnosis"
	"github.com/devusSs/twitch-kraken/internal/logging"
	"github.com/devusSs/twitch-kraken/internal/system"
	"github.com/devusSs/twitch-kraken/internal/updater"
	"github.com/devusSs/twitch-kraken/internal/utils"
)

func main() {
	startTime := time.Now()

	/*
		Usually the default flags will work fine.
		Check the Makefile or documentation for any configuration questions.
	*/
	logPath := flag.String("l", "./logs", "[REQ] sets the logging path")
	cfgPath := flag.String("c", "./files/config.json", "[REQ] sets config path")

	// Diagnosis mode is designed for the app to parse it's own log files.
	//
	// It will print any results from error.log here to help the user figure out potential errors at runtime.
	diagMode := flag.Bool("d", false, "[OPT] runs the app in diagnosis mode")

	// Prints available app build information.
	versionMode := flag.Bool("v", false, "[OPT prints the build information of the app")

	// Generates a secure cookie which will be needed for config file.
	scCookieGen := flag.Bool("sc", false, "[OPT] only generate a secure cookie and exit")

	flag.Parse()

	// Print the version / build information if user wants to, exits after.
	if *versionMode {
		updater.PrintBuildInformationRaw()
		return
	}

	// Generate a random string if user wants to and exit.
	if *scCookieGen {
		log.Printf("[%s] Add to config, DO NOT SHARE: %s\n", logging.WarnSign, utils.RandomString(24))
		return
	}

	// Let's init the map for clear screen functions for supported OS.
	//
	// Windows, Linux, MacOS - anything else will log.Fatal().
	system.InitClearScreen()

	log.Printf("[%s] Checking for updates...\n", logging.InfoSign)

	// Update check - check for release url.
	updateURL, newVersion, updateChangelog, err := updater.FindLatestReleaseURL()
	if err != nil {
		log.Fatalf("[%s] Error checking for updates: %s", logging.ErrorSign, err.Error())
	}

	// Update check - check for release url.
	newVersionAvailable, err := updater.NewerVersionAvailable(newVersion)
	if err != nil {
		log.Fatalf("[%s] Error checking for updates: %s", logging.ErrorSign, err.Error())
	}

	// Update check - perform the actual update.
	if newVersionAvailable {
		log.Printf("[%s] New version available, performing update now...\n", logging.WarnSign)

		if err := updater.DoUpdate(updateURL); err != nil {
			log.Fatalf("[%s] Error performing updates: %s", logging.ErrorSign, err.Error())
		}

		log.Printf("[%s] Update changelog (%s): %s\n", logging.InfoSign, newVersion, updateChangelog)

		log.Printf("[%s] Update successful, please restart the app\n", logging.SuccessSign)

		return
	} else {
		log.Printf("[%s] App is up to date\n", logging.SuccessSign)
	}

	// Update check - setup periodic update check.
	time.AfterFunc(1*time.Hour, func() {
		if err := updater.PeriodicUpdateCheck(); err != nil {
			log.Fatalf("[%s] Error on periodic update check: %s", logging.ErrorSign, err.Error())
		}

		// TODO: maybe print warning to Twitch chat as well? / or send whisper msg to owner
	})

	log.Printf("[%s] Set up periodic update check (1 hour)\n", logging.SuccessSign)

	system.CallClear()

	if err := logging.CreateLogsDirectory(*logPath); err != nil {
		log.Fatalf("[%s] Error creating logs directory: %s", logging.ErrorSign, err.Error())
	}

	if err := logging.CreateFileLoggers(); err != nil {
		log.Fatalf("[%s] Error creating log files: %s", logging.ErrorSign, err.Error())
	}

	logging.CreateConsoleLoggers()

	// ! It's safe to use the logging.WriteX methods from here.

	// Run diagnosis if user wishes to.
	if *diagMode {
		errCount, err := diagnosis.RunDiagnosis(*logPath, *cfgPath)
		if err != nil {
			log.Fatalf("Error running diagnosis: %s", err.Error())
		}
		fmt.Printf("\n[S] Total errors found: %d\n", errCount)
		return
	}

	// Runtime check which might issue a warning since this program should run in headless mode on Linux.
	osV := system.DetermineOS()
	if osV == "unknown" {
		logging.WriteError("Unsupported OS, exiting.")
		os.Exit(1)
	}

	// Test DNS resolution so we know if we are connected to a network.
	if err := system.TestConnection(); err != nil {
		logging.WriteError(err)
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(*cfgPath)
	if err != nil {
		logging.WriteError(err)
		os.Exit(1)
	}

	logging.WriteSuccess("Successfully loaded config")

	if err := cfg.CheckConfig(); err != nil {
		logging.WriteError(err)
		os.Exit(1)
	}

	logging.WriteSuccess("Successfully checked config")

	// Authenticate against services here (Twitch, Spotify, ...).
	auth.DoTwitchAuth(cfg)

	log.Printf("[%s] Got user access token from Twitch: %s\n", logging.InfoSign, authtwitch.AccessToken)

	svc, err := postgres.New(cfg)
	if err != nil {
		logging.WriteError(err)
		os.Exit(1)
	}

	if err := svc.Ping(); err != nil {
		logging.WriteError(err)
		os.Exit(1)
	}

	logging.WriteSuccess("Successfully connected to Postgres database")

	if err := svc.Migrate(); err != nil {
		logging.WriteError(err)
		os.Exit(1)
	}

	logging.WriteSuccess("Successfully migrated database tables")

	// Init Gatekeeper.
	gateKeeper := gatekeeper.InitGateKeeper(cfg.Twitch.BotOwner, svc)

	if err := gateKeeper.LoadSettingsFromStore(); err != nil {
		logging.WriteError(err)
		os.Exit(1)
	}

	if err := gateKeeper.StoreInitialSettings(); err != nil {
		logging.WriteError(err)
		os.Exit(1)
	}

	logging.WriteSuccess("Initiated Gatekeeper")

	twitchBot := bot.New(cfg, svc)

	// Setup needed functions to handle Twitch events.
	twitchBot.SetupHandleFuncs(gateKeeper)

	// Wait group to handle async events.
	wg := &sync.WaitGroup{}

	go func() {
		err := twitchBot.Connect()
		if err != nil {
			// Ignore the disconnect error since that might occur on program exit.
			if err.Error() != "client called Disconnect()" {
				logging.WriteError(err)
				os.Exit(1)
			}
		}
	}()

	logging.WriteInfo(fmt.Sprintf("Initiating app took %.2f second(s)", time.Since(startTime).Seconds()))

	// Artificial delay to stop connection abuse.
	time.Sleep(2 * time.Second)

	logging.WriteInfo("Press CTRL+C to shutdown the app")

	// Wait for CTRL+C for app exit.
	twitchBot.AwaitCancel()

	logging.WriteInfo("Received CTRL+C, shutting down...")

	// !APP EXIT

	// Disconnect from Twitch.
	wg.Add(1)
	if err := twitchBot.Disconnect(wg); err != nil {
		logging.WriteError(err)
	}

	logging.WriteSuccess("Successfully disconnected from Twitch")

	wg.Add(1)
	if err := svc.Close(); err != nil {
		log.Fatalf("[%s] Error closing database connection: %s", logging.ErrorSign, err.Error())
	}
	wg.Done()

	logging.WriteSuccess("Successfully closed database connection")

	// DO NOT USE CONSOLE OR FILE LOGGERS AT THIS POINT ANYMORE
	wg.Add(1)
	if err := logging.CloseLogFiles(); err != nil {
		log.Fatalf("[%s] Error closing logs: %s", logging.ErrorSign, err.Error())
	}
	wg.Done()

	log.Printf("[%s] Successfully closed log files and loggers\n", logging.SuccessSign)

	wg.Wait() // Wait for all operations to finish before exiting app.

	log.Printf("[%s] App ran for %.2f second(s)", logging.InfoSign, time.Since(startTime).Seconds())
}
