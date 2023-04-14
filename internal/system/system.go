// Package to collect (NOT be send anywhere) OS & network information.
package system

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

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

// Detects the network latency if the OS is either Linux or Darwin, skips for Windows (not really supported) yet.
//
// Will issue a warning if the latency exceeds 500 ms since commands may be slowed.
func TestConnection() (float64, error) {
	// Detect the OS version and exit if it does not support sh per default.
	osV := DetermineOS()
	if osV != "linux" && osV != "darwin" {
		return 0, nil
	}

	// Test the actual connection to Twitch.
	comm := "ping -c 3 twitch.tv"
	out, err := exec.Command("/bin/sh", "-c", comm).Output()
	if err != nil {
		return 0, err
	}

	avgResult := strings.Split(string(out), "/")[4]

	avg, err := strconv.ParseFloat(avgResult, 64)
	if err != nil {
		return 0, err
	}

	return avg, err
}
