package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/devusSs/twitch-kraken/internal/auth/authspotify"
	"github.com/devusSs/twitch-kraken/internal/config"
	"github.com/devusSs/twitch-kraken/internal/logging"
	"github.com/devusSs/twitch-kraken/internal/system"
)

// Actually perform the Spotify authentication.
func DoSpotifyAuth(cfg *config.Config) {
	spotifyHandlerFunc := authspotify.InitSpotifyAuth(cfg)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	// Configures and returns the Spotify auth server.
	srv := authspotify.SetupSpotifyAuth(spotifyHandlerFunc, cfg, wg)

	// Start the Spotify auth server in a seperate goroutine, so it is non-blocking.
	go authspotify.StartSpotifyAuth(srv)

	// Check for access token every second.
	// If there is an access token exit the http server.
	atChecker := time.AfterFunc(1*time.Second, func() {
		// Loop until we get an access token, refresh token and token expiry.
		for !gotSpotifyAccessToken() {
			continue
		}
		system.CallClear()
		log.Printf("[%s] Got Spotify tokens and expiry, shutting down server in 2 seconds\n", logging.InfoSign)
		if err := shutdownSpotifyAuth(srv, wg); err != nil {
			log.Fatalf("[%s] Error shutting down Spotify auth server: %s\n", logging.ErrorSign, err.Error())
		}
	})

	// Shut down the Spotify auth server and app after 5 minutes.
	// Only when we do not receive tokens and expiry until then.
	time.AfterFunc(5*time.Minute, func() {
		if !gotSpotifyAccessToken() {
			if err := shutdownSpotifyAuth(srv, wg); err != nil {
				log.Fatalf("[%s] Error shutting down Spotify auth server: %s\n", logging.ErrorSign, err.Error())
			}
			log.Printf("[%s] Shut down Spotify auth server due to inactivity\n", logging.ErrorSign)
			os.Exit(1)
		}
	})

	log.Printf("[%s] Please head to 'http://localhost%s/spotify' to authenticate\n", logging.InfoSign, srv.Addr)
	log.Printf("[%s] ATTENTION: The server will automatically be shut down in 5 minutes\n", logging.WarnSign)
	log.Printf("[%s] Please make sure to authenticate until then\n", logging.WarnSign)

	// Wait until we get an access token.
	wg.Wait()

	// Prevent leaking goroutines.
	atChecker.Stop()
}

// Helper function which checks if we have an access token, refresh token and token expiry.
// Should be used as time.Ticker function (like time.Afterfunc()) to keep running.
func gotSpotifyAccessToken() bool {
	return authspotify.AccessToken != "" && authspotify.RefreshToken != "" && !authspotify.TokenExpiry.IsZero()
}

// Helper function which actually shuts down the server and decrements the wait group.
// Use it after gotTwitchAccessToken() returns true.
func shutdownSpotifyAuth(srv *http.Server, wg *sync.WaitGroup) error {
	time.Sleep(2 * time.Second)

	if err := srv.Shutdown(context.TODO()); err != nil {
		return err
	}

	wg.Done()

	return nil
}

// Function to refresh the token we got.
func RefreshSpotifyTokensFunc() error {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", authspotify.RefreshToken)

	req, err := http.NewRequest(http.MethodPost, "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	authDataBytes := []byte(authspotify.ClientID + ":" + authspotify.ClientSecret)

	encodedDetails := base64.StdEncoding.EncodeToString(authDataBytes)

	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", encodedDetails))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	req.Header.Add("Accept", "application/json")

	client := http.Client{Timeout: 5 * time.Second}

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("got unwanted Spotify response code: %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var refresh authspotify.SpotifyRefreshStruct

	if err := json.Unmarshal(body, &refresh); err != nil {
		return err
	}

	if refresh.AccessToken != "" && refresh.ExpiresIn != 0 {
		authspotify.AccessToken = refresh.AccessToken
		authspotify.TokenExpiry = time.Now().Local().Add(time.Second * time.Duration(refresh.ExpiresIn))
		return nil
	}

	return errors.New("did not get new Spotify token")
}
