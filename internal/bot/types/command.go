package types

import (
	"fmt"
	"time"

	"github.com/devusSs/twitch-kraken/internal/logging"
	"github.com/devusSs/twitch-kraken/internal/utils"
)

type UserLevel int

const (
	Anyone UserLevel = iota
	Moderator
	Owner
)

var (
	inactiveCommands = []string{}
)

func SetCommandOnCooldown(commandName string, commandCD int) error {
	if utils.CheckStringSliceForDuplicates(inactiveCommands, commandName) {
		return fmt.Errorf("command %s already on cooldown list", commandName)
	}

	inactiveCommands = append(inactiveCommands, commandName)

	time.AfterFunc(time.Duration(commandCD)*time.Second, func() {
		err := removeCommandFromCooldown(commandName)
		if err != nil {
			logging.WriteError("Removing command failed: " + err.Error())
			return
		}
	})

	return nil
}

func removeCommandFromCooldown(commandName string) error {
	var err error
	var commIndex int = 0
	var commFound bool = false

	for i, str := range inactiveCommands {
		if str == commandName {
			commIndex = i
			commFound = true
		}
	}

	if !commFound {
		return fmt.Errorf("command %s not on cooldown list", commandName)
	}

	inactiveCommands, err = utils.RemoveStringFromSlice(inactiveCommands, commIndex)
	if err != nil {
		return err
	}

	return nil
}

func CheckCommandOnCooldown(commandName string) bool {
	for _, comm := range inactiveCommands {
		if comm == commandName {
			return true
		}
	}
	return false
}
