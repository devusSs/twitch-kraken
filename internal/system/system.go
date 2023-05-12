// Package to collect (NOT be send anywhere) OS & network information.
package system

import (
	"fmt"
	"net"
	"runtime"

	"github.com/devusSs/twitch-kraken/internal/logging"
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

// While go compiles to most platforms, this program is intended for usage on Linux (headless) or MacOS.
//
// Some functions like ping tests will not be executed when on Windows and the program will issue a warning.
//
// This may be changed in future versions.
func PrintOSWarning(osV string) {
	logging.WriteWarn(fmt.Sprintf("Detected OS: %s ; recommended: Linux", osV))
}

// Detects if there is a working DNS resolver on the host or the network.
//
// This might not represend a working internet connection, usually a good hint however.
func TestConnection() error {
	_, err := net.LookupHost("twitch.tv")
	return err
}
