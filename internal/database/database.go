package database

import (
	"database/sql"
	"time"

	"github.com/devusSs/twitch-kraken/internal/bot/types"
	"github.com/lib/pq"
)

// Service layer for the Postgres connection.
type Service interface {
	Ping() error
	Close() error
	Migrate() error

	LoadGateKeeperSettings() (GateKeeperSettings, error)
	UpdateGateKeeperSettings(GateKeeperSettings) error

	RegisterTwitchUser(string, time.Time) error
	UpdateTwitchUserDC(string, time.Time) error
	UpdateTwitchUserBaseDetails(TwitchUser) error
	UpdateTwitchUserOnBan(TwitchUser) error

	AddTwitchCommand(TwitchCommand) error
	UpdateTwitchCommand(TwitchCommand) (TwitchCommand, error)
	DeleteTwitchCommand(string) error
	GetAllTwitchCommands() ([]TwitchCommand, error)
	GetOneTwitchCommand(string) (TwitchCommand, error)

	AddAuthEvent(AuthEvent) (AuthEvent, error)
	AddMessageEvent(MessageEvent) (MessageEvent, error)
}

// Model for Gatekeeper settings.
type GateKeeperSettings struct {
	ID          int            `db:"id"`
	FilterChat  bool           `db:"filter_chat"`
	FilterLinks bool           `db:"filter_links"`
	IgnoreMods  bool           `db:"ignore_mods"`
	IgnoreSubs  bool           `db:"ignore_subs"`
	SymbolsMax  int            `db:"symbols_max"`
	EmotesMax   int            `db:"emotes_max"`
	BadWords    pq.StringArray `db:"bad_words"`
	SetTime     time.Time      `db:"set"`
}

// Model for Twitch users on database.
type TwitchUser struct {
	ID            int          `db:"id"`
	TwitchID      string       `db:"twitchid"`
	Username      string       `db:"username"`
	DisplayName   string       `db:"displayname"`
	IsMod         sql.NullBool `db:"ismod"`
	FirstSeen     sql.NullTime `db:"firstseen"`
	LastSeen      sql.NullTime `db:"lastseen"`
	HasBeenBanned sql.NullBool `db:"hasbeenbanned"`
	LastBan       sql.NullTime `db:"lastban"`
}

// Model for Twitch commands which can be used by Twitch chat users.
type TwitchCommand struct {
	ID        int             `db:"id"`
	Name      string          `db:"name"`
	Output    string          `db:"output"`
	Userlevel types.UserLevel `db:"userlevel"`
	Cooldown  int             `db:"cooldown"`
	Added     time.Time       `db:"added"`
	Edited    sql.NullTime    `db:"edited"`
}

// Model for Auth Events like Command Added, Command Edited, Command Deleted, Timeouts, Bans, ...
//
// Check the internal/bot/types/event.go file for more information.
type AuthEvent struct {
	ID        int             `db:"id"`
	Type      types.EventType `db:"event_type"`
	Data      string          `db:"event_data"`
	Timestamp time.Time       `db:"event_time"`
}

// Model for messages sent via the Twitch chat. Logs EVERY chat message => might cause overhead.
//
// # Does not log whisper messages.
type MessageEvent struct {
	ID      int       `db:"id"`
	Issuer  string    `db:"issuer"`
	Content string    `db:"content"`
	Sent    time.Time `db:"sent"`
}
