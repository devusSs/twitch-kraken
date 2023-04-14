package gatekeeper

import (
	"regexp"
	"strings"
	"time"

	"github.com/devusSs/twitch-kraken/internal/logging"
	"github.com/gempir/go-twitch-irc/v4"
)

type gateKeeperResult int

type gateKeeperReason string

// Will be sent back for Twitch logging reasons.
type gateKeeperLogReason string

const (
	NoneResult gateKeeperResult = iota
	IssuePurge
	IssueTimeout
	IssueBan

	NoneReason     gateKeeperReason = ""
	LinkReason     gateKeeperReason = "You may only sent a link with permission (!permit)!"
	SymbolsReason  gateKeeperReason = "Stop spamming symbols!"
	EmotesReason   gateKeeperReason = "Stop spamming emotes!"
	BadWordsReason gateKeeperReason = "Please behave and refrain from using bad words!"
	SpammingReason gateKeeperReason = "Please stop spamming messages!"

	NoneLogReason     gateKeeperLogReason = ""
	LinkLogReason     gateKeeperLogReason = "BOT: sent link without permit"
	SymbolsLogReason  gateKeeperLogReason = "BOT: sent too many symbols"
	EmotesLogReason   gateKeeperLogReason = "BOT: sent too many emotes"
	BadWordsLogReason gateKeeperLogReason = "BOT: sent bad word"
	SpammingLogReason gateKeeperLogReason = "BOT: spamming"
)

var (
	// Users may sent 1 message per 2 seconds => 5 messages / 10 seconds.
	//
	// Every 2 seconds 1 message will be removed from this map.
	//
	// If the user exceeds a limit of 5 messages in 10 seconds they will receive a purge for spamming.
	// TODO: add a settings option for that
	// TODO: consider adding this to database, since a lot of users may result in a lot of ram usage
	userMessagesOverTime map[string]int
)

// Checks the message sent on Twitch chat for any of the specified / applied filters.
//
// Returns either an empty string or a corresponding error string which can be sent back to Twitch chat.
func (g *GateKeeper) FilterMessage(message twitch.PrivateMessage) (gateKeeperResult, gateKeeperReason, gateKeeperLogReason) {
	handleMessageCountForUser(message.User.DisplayName)

	filterEnabled := g.settings["filter_chat"]

	ignoreMods := g.settings["ignore_mods"]
	ignoreSubs := g.settings["ignore_subs"]

	filterLinks := g.settings["filter_links"]
	maxSymbols := g.filterParams["symbols_max"]
	maxEmojis := g.filterParams["emotes_max"]
	badWords := g.badWords

	if !filterEnabled {
		return NoneResult, NoneReason, NoneLogReason
	}

	if isUserModOrBroadcaster(message) && ignoreMods {
		return NoneResult, NoneReason, NoneLogReason
	}

	if userSpams(message.User.DisplayName) {
		return IssuePurge, SpammingReason, SpammingLogReason
	}

	if isUserSubscriber(message) && ignoreSubs {
		return NoneResult, NoneReason, NoneLogReason
	}

	if filterLinks {
		containsLink, err := linkInMessage(message)
		if err != nil {
			logging.WriteError(err)
		}

		if containsLink {
			return IssueTimeout, LinkReason, LinkLogReason
		}
	}

	if exceedsMaxSymbols(message, maxSymbols) {
		return IssuePurge, SymbolsReason, SymbolsLogReason
	}

	if exceedsMaxEmojis(message, maxEmojis) {
		return IssuePurge, EmotesReason, EmotesLogReason
	}

	if containsBadWord(message, badWords) {
		return IssuePurge, BadWordsReason, BadWordsLogReason
	}

	return NoneResult, NoneReason, NoneLogReason
}

// Function to check if a user is mod or owner of the bot.
func isUserModOrBroadcaster(message twitch.PrivateMessage) bool {
	user := message.User
	if user.Badges["moderator"] == 1 || user.Badges["broadcaster"] == 1 {
		return true
	}
	return false
}

// Function to check if a user is a subscriber to the channel.
func isUserSubscriber(message twitch.PrivateMessage) bool {
	user := message.User
	return user.Badges["subscriber"] == 1
}

// Filter / Regex logic to check a message for links (https://, http://, www., ...).
func linkInMessage(message twitch.PrivateMessage) (bool, error) {
	pattern := `(?i)\b((?:https?://|www\d{0,3}[.]|[a-z0-9.\-]+[.][a-z]{2,4}/)(?:[^\s()<>]+|\(([^\s()<>]+|(\([^\s()<>]+\)))*\))+(?:\(([^\s()<>]+|(\([^\s()<>]+\)))*\)|[^\s"'<>\[\]` + "`" + `\\{}|’‘“”‘’]))`

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}

	if regex.MatchString(message.Message) {
		return true, nil
	}

	return false, nil
}

// Filter logic to check a message for symbols (pre-defined list).
func exceedsMaxSymbols(message twitch.PrivateMessage, maxSymbols int) bool {
	symbols := []string{"!", "?", ",", ".", ":", ";", "", "(", ")", "\"\""}

	symbolsInMessage := 0

	for _, symbol := range symbols {
		if strings.ContainsAny(message.Message, symbol) {
			symbolsInMessage++
		}
	}

	return symbolsInMessage > maxSymbols
}

// Filer logic to check a message for emotes (pre-defined list).
func exceedsMaxEmojis(message twitch.PrivateMessage, maxEmojis int) bool {
	emojis := []string{
		"😂", "❤️", "😍", "🤣", "😊", "🙏", "💕", "😭", "😘", "👍", "😁", "🔥", "😉", "🎉", "🎶", "😩", "😎", "👌", "💔", "🤔",
		"😏", "😢", "👏", "😱", "😄", "😔", "💯", "😒", "🎂", "👀", "😜", "🌹", "🤷", "😴", "😳", "😋", "🤗", "😞", "💖", "🙄",
		"👊", "😑", "😌", "😕", "😀", "😅", "🤦", "😡", "😪", "💀", "😫", "🎁", "😤", "😬", "🤢", "😇", "😷", "😹", "💋", "👑",
		"🙌", "👋", "😠", "😘", "😨", "👉", "😗", "🌸", "😖", "😥", "🍀", "🍻", "🤑", "😰", "😚", "😆", "🎈", "💙", "🎊", "🤧",
		"🌟", "😻", "😛", "🎵", "🍁", "🐾", "🍁", "🦄", "👶", "🐻", "🙈", "🙉", "🙊", "🐵", "🦁", "🦊", "🐱", "🐶", "🐴", "🐷",
	}

	emojisInMessage := 0

	for _, emoji := range emojis {
		if strings.ContainsAny(message.Message, emoji) {
			emojisInMessage++
		}
	}

	return emojisInMessage > maxEmojis
}

// Filter logic to check a message for bad words (may be set via the bot's options).
func containsBadWord(message twitch.PrivateMessage, badWords []string) bool {
	for _, badWord := range badWords {
		// Check if there is a bad word in the message.
		if strings.Contains(message.Message, badWord) {
			return true
		}
		// Check if there is a bad word as part of any word of the message.
		if strings.ContainsAny(message.Message, badWord) {
			return true
		}
	}
	return false
}

// Adds 1 to the messages sent count for a user.
//
// Also sets up a timer to remove 1 of the message count of a user after 2 seconds.
func handleMessageCountForUser(username string) {
	userMessagesOverTime[username] = userMessagesOverTime[username] + 1

	time.AfterFunc(2*time.Second, func() {
		userMessagesOverTime[username] = userMessagesOverTime[username] - 1
	})
}

// Checks the userMessagesOverTime map to check if the user exceeds the treshhold of 1 message per 2 seconds.
//
// Issues a purge for the user.
func userSpams(username string) bool {
	messageCount := userMessagesOverTime[username]

	return messageCount > 5
}
