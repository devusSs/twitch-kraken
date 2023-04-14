package bot

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/devusSs/twitch-kraken/internal/bot/types"
	"github.com/devusSs/twitch-kraken/internal/database"
	"github.com/devusSs/twitch-kraken/internal/logging"
	"github.com/devusSs/twitch-kraken/internal/utils"
	"github.com/gempir/go-twitch-irc/v4"
)

// Handles any message which starts with a "!".
func (b *TwitchBot) commandHandler(message twitch.PrivateMessage) string {
	messageSplit := strings.Split(message.Message, " ")
	commName := messageSplit[0]

	switch commName {
	case "!commands":
		var err error

		// Handle subcommands like add, edit, delete here.
		subCommand := ""

		// Check if there is an actual subcommand, prevent out of range error.
		if len(messageSplit) > 1 {
			subCommand = messageSplit[1]
		}

		switch subCommand {
		// expected format: !commands add <commandname> <output> (-ul=<userlevel>) (-cd=<cooldown>)
		case "add":
			if !b.IsUserModOrOwner(message) {
				return "You are not allowed to use that command."
			}

			// Remove useless strings from slice.
			for i, str := range messageSplit {
				if str == "add" {
					messageSplit = messageSplit[i+1:]
					break
				}
			}

			commandName := messageSplit[0]
			userlevel := "anyone"
			cooldown := 0

			uLevelInMsg := false
			cdInMsg := false

			// Check the slice for userlevel and cooldown specification.
			for _, str := range messageSplit {
				if strings.Contains(str, "-ul=") {
					userlevel = strings.Split(str, "-ul=")[1]
					uLevelInMsg = true
				}

				if strings.Contains(str, "-cd=") {
					cooldown, err = strconv.Atoi(strings.Split(str, "-cd=")[1])
					if err != nil {
						return err.Error()
					}
					cdInMsg = true
				}
			}

			// Remove command name, uselevel and cooldown from slice.
			messageSplit, err = utils.RemoveStringFromSlice(messageSplit, 0)
			if err != nil {
				return err.Error()
			}

			if uLevelInMsg {
				for i, str := range messageSplit {
					if strings.Contains(str, "-ul=") {
						messageSplit, err = utils.RemoveStringFromSlice(messageSplit, i)
						if err != nil {
							return err.Error()
						}
					}
				}
			}

			if cdInMsg {
				for i, str := range messageSplit {
					if strings.Contains(str, "-cd=") {
						messageSplit, err = utils.RemoveStringFromSlice(messageSplit, i)
						if err != nil {
							return err.Error()
						}
					}
				}
			}

			// Double check for valid userlevel and cooldown.
			userlevel = strings.ToLower(userlevel)

			if userlevel != "anyone" && userlevel != "moderator" && userlevel != "owner" {
				return fmt.Sprintf("Invalid userlevel specified: %s", userlevel)
			}

			if cooldown < 0 {
				return fmt.Sprintf("Invalid cooldown specified: %d", cooldown)
			}

			// Add command to database.
			comm := database.TwitchCommand{}
			comm.Name = commandName
			comm.Output = strings.Join(messageSplit, " ")

			// Convert userlevel to corresponding enum value.
			switch userlevel {
			case "anyone":
				comm.Userlevel = types.Anyone
			case "moderator":
				comm.Userlevel = types.Moderator
			case "owner":
				comm.Userlevel = types.Owner
			default:
				return fmt.Sprintf("Invalid userlevel specified: %s", userlevel)
			}

			comm.Cooldown = cooldown

			comm.Added = time.Now()

			if err := b.Service.AddTwitchCommand(comm); err != nil {
				if err == sql.ErrTxDone {
					return fmt.Sprintf("Command %s already exists.", comm.Name)
				}
				return err.Error()
			}

			var event database.AuthEvent

			innerData, err := utils.MarshalStruct(types.CommandEvent{
				Issuer:      message.User.Name,
				CommandName: comm.Name,
			})
			if err != nil {
				logging.WriteError(err)
			}

			event.Type = types.CommandAdded
			event.Data = innerData
			event.Timestamp = time.Now()

			_, err = b.Service.AddAuthEvent(event)
			if err != nil {
				logging.WriteError(err)
			}

			return fmt.Sprintf("Command %s has successfully been added.", comm.Name)

		// expected format: !commands edit <commandname> <output> (-ul=<userlevel>) (-cd=<cooldown>)
		case "edit":
			if !b.IsUserModOrOwner(message) {
				return "You are not allowed to use that command."
			}

			// Remove useless strings from slice.
			for i, str := range messageSplit {
				if str == "edit" {
					messageSplit = messageSplit[i+1:]
					break
				}
			}

			commandName := messageSplit[0]
			userlevel := ""
			cooldown := 0

			uLevelInMsg := false
			cdInMsg := false

			// Check the slice for userlevel and cooldown specification.
			for _, str := range messageSplit {
				if strings.Contains(str, "-ul=") {
					userlevel = strings.Split(str, "-ul=")[1]
					uLevelInMsg = true
				}

				if strings.Contains(str, "-cd=") {
					cooldown, err = strconv.Atoi(strings.Split(str, "-cd=")[1])
					if err != nil {
						return err.Error()
					}
					cdInMsg = true
				}
			}

			// Remove command name, userlevel and cooldown from slice.
			messageSplit, err = utils.RemoveStringFromSlice(messageSplit, 0)
			if err != nil {
				return err.Error()
			}

			if uLevelInMsg {
				for i, str := range messageSplit {
					if strings.Contains(str, "-ul=") {
						messageSplit, err = utils.RemoveStringFromSlice(messageSplit, i)
						if err != nil {
							return err.Error()
						}
					}
				}
			}

			if cdInMsg {
				for i, str := range messageSplit {
					if strings.Contains(str, "-cd=") {
						messageSplit, err = utils.RemoveStringFromSlice(messageSplit, i)
						if err != nil {
							return err.Error()
						}
					}
				}
			}

			// Grab old userlevel and cooldown if none were specified
			oldCmd, err := b.Service.GetOneTwitchCommand(commandName)
			if err != nil {
				if err == sql.ErrNoRows {
					return fmt.Sprintf("Command %s does not exist.", commandName)
				}
				return err.Error()
			}

			if !uLevelInMsg {
				userlevel = oldCmd.Userlevel.String()
			}

			if !cdInMsg {
				cooldown = oldCmd.Cooldown
			}

			// Double check for valid userlevel and cooldown.
			userlevel = strings.ToLower(userlevel)

			if userlevel != "anyone" && userlevel != "moderator" && userlevel != "owner" {
				return fmt.Sprintf("Invalid userlevel specified: %s", userlevel)
			}

			if cooldown < 0 {
				return fmt.Sprintf("Invalid cooldown specified: %d", cooldown)
			}

			newCmd := database.TwitchCommand{}

			// Convert userlevel to corresponding enum value.
			switch userlevel {
			case "anyone":
				newCmd.Userlevel = types.Anyone
			case "moderator":
				newCmd.Userlevel = types.Moderator
			case "owner":
				newCmd.Userlevel = types.Owner
			default:
				return fmt.Sprintf("Invalid userlevel specified: %s", userlevel)
			}

			newCmd.Name = commandName
			newCmd.Output = strings.Join(messageSplit, " ")
			newCmd.Cooldown = cooldown
			newCmd.Edited.Time = time.Now()

			cmdReturn, err := b.Service.UpdateTwitchCommand(newCmd)
			if err != nil {
				return err.Error()
			}

			var event database.AuthEvent

			innerData, err := utils.MarshalStruct(types.CommandEvent{
				Issuer:      message.User.Name,
				CommandName: newCmd.Name,
			})
			if err != nil {
				logging.WriteError(err)
			}

			event.Type = types.CommandEdited
			event.Data = innerData
			event.Timestamp = time.Now()

			_, err = b.Service.AddAuthEvent(event)
			if err != nil {
				logging.WriteError(err)
			}

			return fmt.Sprintf("Command %s has successfully been updated.", cmdReturn.Name)

		// expected format: !commands delete <commandname>
		case "delete":
			if !b.IsUserModOrOwner(message) {
				return "You are not allowed to use that command."
			}

			if err := b.Service.DeleteTwitchCommand(messageSplit[len(messageSplit)-1]); err != nil {
				if err == sql.ErrNoRows {
					return fmt.Sprintf("Command %s does not exist.", messageSplit[len(messageSplit)-1])
				}
				return err.Error()
			}

			var event database.AuthEvent

			innerData, err := utils.MarshalStruct(types.CommandEvent{
				Issuer:      message.User.Name,
				CommandName: messageSplit[len(messageSplit)-1],
			})
			if err != nil {
				logging.WriteError(err)
			}

			event.Type = types.CommandDeleted
			event.Data = innerData
			event.Timestamp = time.Now()

			_, err = b.Service.AddAuthEvent(event)
			if err != nil {
				logging.WriteError(err)
			}

			return fmt.Sprintf("Command %s has successfully been deleted.", messageSplit[len(messageSplit)-1])

		// expected format: !commands
		// Return a list of all available commands if no subcommand was specified.
		case "":
			comms, err := b.Service.GetAllTwitchCommands()
			if err != nil {
				return err.Error()
			}

			// Return custom message if no commands on database.
			if len(comms) == 0 {
				return "No commands on database yet."
			}

			cNames := []string{}

			for _, comm := range comms {
				cNames = append(cNames, comm.Name)
			}

			return strings.Join(cNames, ", ")

		// No matching subcommand was found.
		default:
			return fmt.Sprintf("Invalid subcommand: %s", subCommand)
		}

	// TODO: implement more built-in commands like title, setttitle etc.

	// Return any matching command output from database here.
	default:
		comm, err := b.Service.GetOneTwitchCommand(commName)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Sprintf("Command %s not found.", commName)
			}
			return err.Error()
		}

		if strings.ToLower(comm.Userlevel.String()) != "anyone" {
			if !b.IsUserModOrOwner(message) {
				return "You are not allowed to use that command."
			}
		}

		// Check if command is still on cooldown.
		if types.CheckCommandOnCooldown(comm.Name) {
			return fmt.Sprintf("Command %s is still on cooldown.", comm.Name)
		}

		// Add command to cooldown list in case it has a cooldown.
		if err := types.SetCommandOnCooldown(comm.Name, comm.Cooldown); err != nil {
			return err.Error()
		}

		var event database.AuthEvent

		innerData, err := utils.MarshalStruct(types.CommandEvent{
			Issuer:      message.User.Name,
			CommandName: comm.Name,
		})
		if err != nil {
			logging.WriteError(err)
		}

		event.Type = types.CommandCalled
		event.Data = innerData
		event.Timestamp = time.Now()

		_, err = b.Service.AddAuthEvent(event)
		if err != nil {
			logging.WriteError(err)
		}

		return comm.Output
	}
}

// ! BUILT-IN COMMANDS / TWITCH BUILT-IN COMMANDS

// Purges a user's last message.
//
// https://www.alphr.com/delete-single-message-twitch/
func (b *TwitchBot) PurgeUser(username, reason string) {
	b.Client.Say(b.Channel, fmt.Sprintf("/timeout %s 1s %s", username, reason))
}

// Timeouts a user for duration X seconds.
func (b *TwitchBot) TimeoutUser(username, reason string, duration int) {
	b.Client.Say(b.Channel, fmt.Sprintf("/timeout %s %ds %s", username, duration, reason))
}

// Permanently bans a user.
func (b *TwitchBot) BanUser(username, reason string) {
	b.Client.Say(b.Channel, fmt.Sprintf("/ban %s %s", username, reason))
}
