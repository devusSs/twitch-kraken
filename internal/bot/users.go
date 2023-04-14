package bot

import (
	"time"

	"github.com/devusSs/twitch-kraken/internal/database"
	"github.com/gempir/go-twitch-irc/v4"
)

// Function registers username on database. Cannot add any details like twitchid etc.
func (b *TwitchBot) AddUserOnConnect(message twitch.UserJoinMessage) error {
	return b.Service.RegisterTwitchUser(message.User, time.Now())
}

// Function updates user's last seen on database. Cannot add any details like twitchid etc.
func (b *TwitchBot) EditUserOnDisconnect(message twitch.UserPartMessage) error {
	return b.Service.UpdateTwitchUserDC(message.User, time.Now())
}

// Function updates user's base details. This will add details like twitchid etc.
func (b *TwitchBot) AddUserDetailsOnMessage(message twitch.PrivateMessage) error {
	user := database.TwitchUser{}

	user.TwitchID = message.User.ID
	user.Username = message.User.Name
	user.DisplayName = message.User.DisplayName
	if message.Tags["mod"] == "1" || user.Username == b.Owner {
		user.IsMod.Bool = true
	} else {
		user.IsMod.Bool = false
	}
	user.LastSeen.Time = time.Now()

	return b.Service.UpdateTwitchUserBaseDetails(user)
}

// Function updates user's details on ban events (no timeouts yet). Some details like ismod may be missing.
func (b *TwitchBot) AddUserDetailsOnBan(message twitch.ClearChatMessage) error {
	user := database.TwitchUser{}

	user.TwitchID = message.TargetUserID
	user.Username = message.TargetUsername
	user.LastSeen.Time = time.Now()
	user.HasBeenBanned.Bool = true
	user.LastBan.Time = time.Now()

	return b.Service.UpdateTwitchUserOnBan(user)
}

// Function to check if a user is mod or owner of the bot.
func (b *TwitchBot) IsUserModOrOwner(message twitch.PrivateMessage) bool {
	user := message.User
	if user.Badges["moderator"] == 1 || user.Name == b.Owner {
		return true
	}
	return false
}
