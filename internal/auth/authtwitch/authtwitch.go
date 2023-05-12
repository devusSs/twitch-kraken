// Coded using official Twitch documentation.
//
// https://github.com/twitchdev/authentication-go-sample/blob/main/oauth-authorization-code/main.go
// All rights go to their owners.
package authtwitch

import (
	"context"
	"crypto/rand"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/devusSs/twitch-kraken/internal/config"
	"github.com/devusSs/twitch-kraken/internal/logging"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/twitch"
)

const (
	stateCallbackKey = "oauth-state-callback"
	oauthSessionName = "oauth-session"
	oauthTokenKey    = "oauth-token"
)

var (
	ClientID     string
	ClientSecret string
	scopes       = []string{"channel:manage:broadcast", "moderator:manage:banned_users", "moderation:read", "moderator:manage:chat_settings", "user:edit"}
	redirectURL  string
	oauth2Config *oauth2.Config
	cookieSecret []byte
	cookieStore  *sessions.CookieStore

	AccessToken  string
	RefreshToken string
	TokenExpiry  time.Time
)

// Init function to prepare bot for Twitch token generation.
func InitTwitchAuth(cfg *config.Config) func(path string, handler handler) {
	// Init Client ID & Client Secret.
	ClientID = cfg.TwitchAuth.ClientID
	ClientSecret = cfg.TwitchAuth.ClientSecret

	// Init Redirect URL.
	redirectURL = cfg.TwitchAuth.RedirectURL

	// Init Cookie Secret.
	cookieSecret = []byte(cfg.TwitchAuth.SecureCookie)

	// Init the Cookie Store.
	cookieStore = sessions.NewCookieStore(cookieSecret)

	// Gob encoding for gorilla/sessions
	gob.Register(&oauth2.Token{})

	// Set the oauth2 config after initiating values.
	oauth2Config = &oauth2.Config{
		ClientID:     ClientID,
		ClientSecret: ClientSecret,
		Scopes:       scopes,
		Endpoint:     twitch.Endpoint,
		RedirectURL:  redirectURL,
	}

	var middleware = func(h handler) handler {
		return func(w http.ResponseWriter, r *http.Request) (err error) {
			// parse POST body, limit request size
			if err = r.ParseForm(); err != nil {
				return annotateError(err, "Something went wrong! Please try again.", http.StatusBadRequest)
			}

			return h(w, r)
		}
	}

	var errorHandling = func(handler func(w http.ResponseWriter, r *http.Request) error) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := handler(w, r); err != nil {
				var errorString string = "Something went wrong! Please try again."
				var errorCode int = 500

				if v, ok := err.(humanReadableError); ok {
					errorString, errorCode = v.humanError(), v.hTTPCode()
				}

				log.Println(err)
				w.Write([]byte(errorString))
				w.WriteHeader(errorCode)
				return
			}
		})
	}

	var handleFunc = func(path string, handler handler) {
		http.Handle(path, errorHandling(middleware(handler)))
	}

	return handleFunc
}

// Function that actually runs the local auth server.
func SetupTwitchAuth(handleFunc func(path string, handler handler), cfg *config.Config, wg *sync.WaitGroup) *http.Server {
	handleFunc("/", handleRoot)
	handleFunc("/login", handleLogin)
	handleFunc("/twitch/redirect", handleOAuth2Callback)

	// TODO: put port here from config
	srv := &http.Server{Addr: ":9001"}

	return srv
}

// Function that starts the server.
// Should always be called as it's own goroutine.
func StartTwitchAuth(srv *http.Server) {
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("[%s] Error starting Twitch auth server: %s\n", logging.ErrorSign, err.Error())
	}
}

// HandleRoot is a Handler that shows a login button. In production, if the frontend is served / generated
// by Go, it should use html/template to prevent XSS attacks.
// TODO: implement template here
func handleRoot(w http.ResponseWriter, r *http.Request) (err error) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<html><body><a href="/login">Login using Twitch</a></body></html>`))

	return
}

// HandleLogin is a Handler that redirects the user to Twitch for login, and provides the 'state'
// parameter which protects against login CSRF.
func handleLogin(w http.ResponseWriter, r *http.Request) (err error) {
	session, err := cookieStore.Get(r, oauthSessionName)
	if err != nil {
		log.Printf("corrupted session %s -- generated new", err)
		err = nil
	}

	var tokenBytes [255]byte
	if _, err := rand.Read(tokenBytes[:]); err != nil {
		return annotateError(err, "Couldn't generate a session!", http.StatusInternalServerError)
	}

	state := hex.EncodeToString(tokenBytes[:])

	session.AddFlash(state, stateCallbackKey)

	if err = session.Save(r, w); err != nil {
		return
	}

	http.Redirect(w, r, oauth2Config.AuthCodeURL(state), http.StatusTemporaryRedirect)

	return
}

// HandleOauth2Callback is a Handler for oauth's 'redirect_uri' endpoint;
// it validates the state token and retrieves an OAuth token from the request parameters.
func handleOAuth2Callback(w http.ResponseWriter, r *http.Request) (err error) {
	session, err := cookieStore.Get(r, oauthSessionName)
	if err != nil {
		log.Printf("corrupted session %s -- generated new", err)
		err = nil
	}

	// ensure we flush the csrf challenge even if the request is ultimately unsuccessful
	defer func() {
		if err := session.Save(r, w); err != nil {
			log.Printf("error saving session: %s", err)
		}
	}()

	switch stateChallenge, state := session.Flashes(stateCallbackKey), r.FormValue("state"); {
	case state == "", len(stateChallenge) < 1:
		err = errors.New("missing state challenge")
	case state != stateChallenge[0]:
		err = fmt.Errorf("invalid oauth state, expected '%s', got '%s'", state, stateChallenge[0])
	}

	if err != nil {
		return annotateError(
			err,
			"Couldn't verify your confirmation, please try again.",
			http.StatusBadRequest,
		)
	}

	token, err := oauth2Config.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		return
	}

	// add the oauth token to session
	session.Values[oauthTokenKey] = token

	AccessToken = token.AccessToken
	RefreshToken = token.RefreshToken
	TokenExpiry = token.Expiry

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)

	return
}

// HumanReadableError represents error information
// that can be fed back to a human user.
//
// This prevents internal state that might be sensitive
// being leaked to the outside world.
type humanReadableError interface {
	humanError() string
	hTTPCode() int
}

// HumanReadableWrapper implements HumanReadableError
type humanReadableWrapper struct {
	ToHuman string
	Code    int
	error
}

func (h humanReadableWrapper) HumanError() string { return h.ToHuman }
func (h humanReadableWrapper) HTTPCode() int      { return h.Code }

// AnnotateError wraps an error with a message that is intended for a human end-user to read,
// plus an associated HTTP error code.
func annotateError(err error, annotation string, code int) error {
	if err == nil {
		return nil
	}
	return humanReadableWrapper{ToHuman: annotation, error: err}
}

type handler func(http.ResponseWriter, *http.Request) error

// Refresh struct.
type TwitchRefreshStruct struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	Scope        []string `json:"scope"`
	TokenType    string   `json:"token_type"`
}
