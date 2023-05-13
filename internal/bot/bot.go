package bot

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/devusSs/twitch-kraken/internal/bot/gatekeeper"
	"github.com/devusSs/twitch-kraken/internal/bot/types"
	"github.com/devusSs/twitch-kraken/internal/config"
	"github.com/devusSs/twitch-kraken/internal/database"
	"github.com/devusSs/twitch-kraken/internal/logging"
	"github.com/devusSs/twitch-kraken/internal/utils"
	"github.com/gempir/go-twitch-irc/v4"
)

type TwitchBot struct {
	Channel string
	Owner   string
	Editors []string
	Client  *twitch.Client
	Service database.Service
}

// Inits a new Twitch client and bot instance.
func New(cfg *config.Config, svc database.Service) *TwitchBot {
	bot := TwitchBot{}

	client := twitch.NewClient(cfg.Twitch.BotLogin, fmt.Sprintf("oauth:%s", cfg.Twitch.BotPassword))

	client.Capabilities = []string{twitch.TagsCapability, twitch.CommandsCapability, twitch.MembershipCapability}

	bot.Channel = cfg.Twitch.JoinChannel
	bot.Owner = cfg.Twitch.BotOwner
	bot.Editors = cfg.Twitch.Editors
	bot.Client = client
	bot.Service = svc

	return &bot
}

// Join the specified channel and connect to Twitch's IRC server.
func (b *TwitchBot) Connect() error {
	b.Client.Join(b.Channel)
	return b.Client.Connect()
}

// Default function to send a chat message on specified channel.
func (b *TwitchBot) SendMessage(message string) {
	b.Client.Say(b.Channel, message)
}

// Send a welcome message to chat so the users know the bot is online.
func (b *TwitchBot) SendHelloMessage() {
	b.SendMessage(fmt.Sprintf("@%s => Bot is up and running!", b.Owner))
	logging.WriteSuccess("Sent hello message to chat")
}

// Send a whisper message to the bot owner that an update is available.
func (b *TwitchBot) SendUpdateNotification() {
	b.SendMessage(fmt.Sprintf("@%s => New Bot version available!", b.Owner))
	logging.WriteSuccess("Sent update message to chat")
}

// This function will block further execution until CTRL+C is hit.
//
// # NOTE: This function will NOT disconnect the bot. Use the default function Disconnect() for that.
func (b *TwitchBot) AwaitCancel() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done
	fmt.Println("")
}

// Leave the channel and disconnect from Twitch server.
func (b *TwitchBot) Disconnect(wg *sync.WaitGroup) error {
	b.Client.Depart(b.Channel)
	err := b.Client.Disconnect()
	wg.Done()
	return err
}

// General function to setup handlers for all Twitch / TMI events.
func (b *TwitchBot) SetupHandleFuncs(g *gatekeeper.GateKeeper) {
	// When we connect to the Twitch chat.
	b.Client.OnConnect(func() {
		logging.WriteSuccess("Successfully connected to Twitch")
	})

	// Default channel message. Bot ignores whisper messages.
	b.Client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		if err := b.AddUserDetailsOnMessage(message); err != nil {
			logging.WriteError(err)
		}

		// Checks a user's Twitch chat message for the specified filters.
		//
		// Will issue a purge, timeout (10 mins) or ban depending on result.
		gkRes, gkReas, gkLog := g.FilterMessage(message)
		if gkRes != gatekeeper.NoneResult {
			b.Client.Say(b.Channel, fmt.Sprintf("@%s => %s", message.User.DisplayName, gkReas))

			switch gkRes {
			case gatekeeper.IssuePurge:
				b.PurgeUser(message.User.DisplayName, string(gkLog))
			case gatekeeper.IssueTimeout:
				b.TimeoutUser(message.User.DisplayName, string(gkLog), 600)
			case gatekeeper.IssueBan:
				b.BanUser(message.User.DisplayName, string(gkLog))
			default:
				logging.WriteError(fmt.Sprintf("Got invalid GateKeeper result %d with reason %s", gkRes, gkReas))
			}

			return
		}

		_, err := b.Service.AddMessageEvent(database.MessageEvent{
			Issuer:  message.User.Name,
			Content: message.Message,
			Sent:    message.Time,
		})
		if err != nil {
			logging.WriteError(err)
		}

		// Check for leading "!" in case message might be a bot command.
		if strings.Index(message.Message, "!") == 0 {
			commReturn := b.commandHandler(message)
			b.Client.Say(b.Channel, fmt.Sprintf("@%s => %s", message.User.DisplayName, commReturn))
			return
		}
	})

	// Ban or timeout events
	b.Client.OnClearChatMessage(func(message twitch.ClearChatMessage) {
		if err := b.AddUserDetailsOnBan(message); err != nil {
			logging.WriteError(err)
		}

		eventData, err := utils.MarshalStruct(types.UserEvent{
			Target:   message.TargetUsername,
			Duration: message.BanDuration,
		})
		if err != nil {
			logging.WriteError(err)
			return
		}

		var event database.AuthEvent
		event.Type = types.UserBan
		// If ban duration > 0 it's a timeout event, else it's a ban event.
		if message.BanDuration > 0 {
			event.Type = types.UserTimeout
		}
		event.Data = eventData
		event.Timestamp = time.Now()

		_, err = b.Service.AddAuthEvent(event)
		if err != nil {
			logging.WriteError(err)
		}
	})

	// TODO: implement later
	/*
		// Message was purged event
		b.Client.OnClearMessage(func(message twitch.ClearMessage) {})

			// Events like submode, follower-only, etc.
			b.Client.OnRoomStateMessage(func(message twitch.RoomStateMessage) {})

			// ??
			b.Client.OnGlobalUserStateMessage(func(message twitch.GlobalUserStateMessage) {})

			// Events like hosting or raiding another channel
			b.Client.OnNoticeMessage(func(message twitch.NoticeMessage) {})

			// Events like gaining a sub, resub or raids
			b.Client.OnUserNoticeMessage(func(message twitch.UserNoticeMessage) {})

			// ??
			b.Client.OnUserStateMessage(func(message twitch.UserStateMessage) {})
	*/

	// User joins channel event
	b.Client.OnUserJoinMessage(func(message twitch.UserJoinMessage) {
		if err := b.AddUserOnConnect(message); err != nil {
			logging.WriteError(err)
		}
	})

	// User leaves channel event
	b.Client.OnUserPartMessage(func(message twitch.UserPartMessage) {
		if err := b.EditUserOnDisconnect(message); err != nil {
			logging.WriteError(err)
		}
	})

	// Reconnect event
	b.Client.OnSelfJoinMessage(func(message twitch.UserJoinMessage) {
		logging.WriteSuccess(fmt.Sprintf("Successfully joined channel \"%s\"", b.Channel))
		b.SendHelloMessage()
	})

	// Disconnect event
	b.Client.OnSelfPartMessage(func(message twitch.UserPartMessage) {
		logging.WriteError("Left channel, attempting to rejoin...")
		b.Client.Join(b.Channel)
	})

	logging.WriteSuccess("Setup handle functions for Twitch events")
}
