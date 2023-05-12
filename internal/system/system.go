// Package to collect (NOT be send anywhere) OS & network information.
package system

import (
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"

	"github.com/devusSs/twitch-kraken/internal/logging"
)

var (
	// Clear screen function for Windows, Linux and MacOS.
	//
	// Needs to be initiated first!
	clear map[string]func()
)

// Detects the OS of the runtime => determined by compile option GOOS?
func DetermineOS() string {
	switch runtime.GOOS {
	case "windows":
		return "windows"
	case "linux":
		return "linux"
	case "darwin":
		return "darwin"
	default:
		return "unknown"
	}
}

// Detects if there is a working DNS resolver on the host or the network.
//
// This might not represend a working internet connection, usually a good hint however.
func TestConnection() error {
	_, err := net.LookupHost("twitch.tv")
	return err
}

// Init function which clears the screen.
func InitClearScreen() {
	clear = make(map[string]func())

	clear["linux"] = func() {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			return
		}
	}
	clear["darwin"] = func() {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			return
		}
	}
	clear["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			return
		}
	}
}

// Actual screen clearing function
func CallClear() {
	value, ok := clear[runtime.GOOS]
	if ok {
		value()
	} else {
		log.Fatalf("[%s] Unsupported platform\n", logging.ErrorSign)
	}
}
