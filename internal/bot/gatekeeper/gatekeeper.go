package gatekeeper

import (
	"database/sql"
	"time"

	"github.com/devusSs/twitch-kraken/internal/database"
)

// GateKeeper "Engine".
type GateKeeper struct {
	owner        string
	service      database.Service
	settings     map[string]bool
	filterParams map[string]int
	badWords     []string
}

// Init a new GateKeeper instance.
//
// # Will load default settings initially, load custom settings via LoadSettingsFromStore().
func InitGateKeeper(owner string, svc database.Service) *GateKeeper {
	userMessagesOverTime = make(map[string]int)

	g := GateKeeper{}

	g.owner = owner

	g.service = svc

	g.settings = make(map[string]bool)
	g.settings["filter_chat"] = true
	g.settings["filter_links"] = true
	g.settings["ignore_mods"] = true
	g.settings["ignore_subs"] = false

	g.filterParams = make(map[string]int)
	// How many symbols can be sent per message before purge.
	g.filterParams["symbols_max"] = 5
	// How many emotes can be sent per message before purge.
	g.filterParams["emotes_max"] = 3

	g.badWords = []string{}

	return &g
}

// Loads settings in case they were already stored in database.
//
// # Loads default values if nothing was set on the database yet.
func (g *GateKeeper) LoadSettingsFromStore() error {
	settings, err := g.service.LoadGateKeeperSettings()
	if err != nil {
		// Use default values if no custom settings set on database.
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	g.settings["filter_chat"] = settings.FilterChat
	g.settings["filter_links"] = settings.FilterLinks
	g.settings["ignore_mods"] = settings.IgnoreMods
	g.settings["ignore_subs"] = settings.IgnoreSubs

	g.filterParams["symbols_max"] = settings.SymbolsMax
	g.filterParams["emotes_max"] = settings.EmotesMax

	g.badWords = settings.BadWords

	return nil
}

// Sets the settings of Gatekeeper on startup on the database.
func (g *GateKeeper) StoreInitialSettings() error {
	s := database.GateKeeperSettings{}

	s.FilterChat = g.settings["filter_chat"]
	s.FilterLinks = g.settings["filter_links"]
	s.IgnoreMods = g.settings["ignore_mods"]
	s.IgnoreSubs = g.settings["ignore_subs"]
	s.SymbolsMax = g.filterParams["symbols_max"]
	s.EmotesMax = g.filterParams["emotes_max"]
	s.BadWords = g.badWords
	s.SetTime = time.Now()

	return g.service.UpdateGateKeeperSettings(s)
}

// TODO: add function to change settings
