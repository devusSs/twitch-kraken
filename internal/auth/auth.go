package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/devusSs/twitch-kraken/internal/auth/authtwitch"
	"github.com/devusSs/twitch-kraken/internal/config"
	"github.com/devusSs/twitch-kraken/internal/logging"
	"github.com/devusSs/twitch-kraken/internal/system"
)

// Actually perform the Twitch authentication.
func DoTwitchAuth(cfg *config.Config) {
	twitchHandlerFunc := authtwitch.InitTwitchAuth(cfg)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	// Configures and returns the Twitch auth server.
	srv := authtwitch.SetupTwitchAuth(twitchHandlerFunc, cfg, wg)

	// Start the Twitch auth server in a seperate goroutine, so it is non-blocking.
	go authtwitch.StartTwitchAuth(srv)

	// Check for access token every second.
	// If there is an access token exit the http server.
	atChecker := time.AfterFunc(1*time.Second, func() {
		// Loop until we get an access token, refresh token and token expiry.
		for !gotAccessToken() {
			continue
		}
		system.CallClear()
		log.Printf("[%s] Got Twitch tokens and expiry, shutting down server in 5 seconds\n", logging.InfoSign)
		if err := shutdownTwitchAuth(srv, wg); err != nil {
			log.Fatalf("[%s] Error shutting down Twitch auth server: %s\n", logging.ErrorSign, err.Error())
		}
	})

	// Shut down the Twitch auth server and app after 5 minutes.
	// Only when we do not receive tokens and expiry until then.
	time.AfterFunc(5*time.Minute, func() {
		if !gotAccessToken() {
			if err := shutdownTwitchAuth(srv, wg); err != nil {
				log.Fatalf("[%s] Error shutting down Twitch auth server: %s\n", logging.ErrorSign, err.Error())
			}
			log.Printf("[%s] Shut down Twitch auth server due to inactivity\n", logging.ErrorSign)
			os.Exit(1)
		}
	})

	log.Printf("[%s] Please head to '%s' to authenticate\n", logging.InfoSign, srv.Addr)
	log.Printf("[%s] ATTENTION: The server will automatically be shut down in 5 minutes\n", logging.WarnSign)
	log.Printf("[%s] Please make sure to authenticate until then\n", logging.WarnSign)

	// Wait until we get an access token.
	wg.Wait()

	// Prevent leaking goroutines.
	atChecker.Stop()
}

// Helper function which checks if we have an access token, refresh token and token expiry.
// Should be used as time.Ticker function (like time.Afterfunc()) to keep running.
func gotAccessToken() bool {
	return authtwitch.AccessToken != "" && authtwitch.RefreshToken != "" && !authtwitch.TokenExpiry.IsZero()
}

// Helper function which actually shuts down the server and decrements the wait group.
// Use it after gotAccessToken() returns true.
func shutdownTwitchAuth(srv *http.Server, wg *sync.WaitGroup) error {
	time.Sleep(2 * time.Second)

	if err := srv.Shutdown(context.TODO()); err != nil {
		return err
	}

	wg.Done()

	return nil
}

// Function to refresh the token we got.
func RefreshTokensFunc() error {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", authtwitch.RefreshToken)
	data.Set("client_id", authtwitch.ClientID)
	data.Set("client_secret", authtwitch.ClientSecret)

	req, err := http.NewRequest(http.MethodPost, "https://id.twitch.tv/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")

	client := http.Client{Timeout: 5 * time.Second}

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("got unwated Twitch response code: %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var refresh authtwitch.TwitchRefreshStruct

	if err := json.Unmarshal(body, &refresh); err != nil {
		return err
	}

	if refresh.AccessToken != "" && refresh.RefreshToken != "" {
		authtwitch.AccessToken = refresh.AccessToken
		authtwitch.RefreshToken = refresh.RefreshToken
		return nil
	}

	return errors.New("did not get new Twitch tokens")
}
