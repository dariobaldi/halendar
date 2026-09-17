package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dariobaldi/halendar_back/internal/calendarimport"
	"github.com/dariobaldi/halendar_back/internal/data"
	"github.com/dariobaldi/halendar_back/internal/emailimport"
	"github.com/dariobaldi/halendar_back/internal/validator"
	"github.com/google/uuid"
	"halendar/calendar"
)

// caldavConfig is the non-secret part of a CalDAV calendar's config (the password
// lives encrypted in calendar_account_credentials, everything else here doesn't need
// to be).
type caldavConfig struct {
	URL      string `json:"url"`
	User     string `json:"user"`
	Calendar string `json:"calendar,omitempty"`
	Timezone string `json:"timezone,omitempty"`
}

// listCalendarAccountsHandler lists the calendars the current user has connected.
func (app *app) listCalendarAccountsHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	accounts, err := app.models.CalendarAccounts.GetForUser(user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if err := app.writeJSON(w, http.StatusOK, envelope{"accounts": accounts}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// connectCalendarAccountHandler returns the URL to open in the system browser to
// start an OAuth2-based provider's consent flow (Google today). CalDAV calendars
// don't go through this -- see connectCaldavCalendarHandler.
func (app *app) connectCalendarAccountHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	providerName := app.readStringParam(r, "provider")
	provider, ok := app.calendarProviders.Get(providerName)
	if !ok {
		app.badRequestResponse(w, r, fmt.Errorf("unknown or unconfigured provider %q", providerName))
		return
	}

	state, err := app.models.Tokens.New(user.ID, oauthStateTTL, data.ScopeOAuthState)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if err := app.writeJSON(w, http.StatusOK, envelope{"auth_url": provider.AuthCodeURL(state.Plaintext)}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// calendarAccountCallbackHandler is where an OAuth2 calendar provider redirects the
// user's browser back to after consent. See emailAccountCallbackHandler for why this
// has no Authorization header and renders a page instead of JSON.
func (app *app) calendarAccountCallbackHandler(w http.ResponseWriter, r *http.Request) {
	providerName := app.readStringParam(r, "provider")
	provider, ok := app.calendarProviders.Get(providerName)
	if !ok {
		app.writeOAuthResult(w, http.StatusBadRequest, false, "This provider isn't configured.")
		return
	}

	qs := r.URL.Query()
	if errParam := qs.Get("error"); errParam != "" {
		app.writeOAuthResult(w, http.StatusOK, false, "Connection cancelled.")
		return
	}
	code := qs.Get("code")
	state := qs.Get("state")
	if code == "" || state == "" {
		app.writeOAuthResult(w, http.StatusBadRequest, false, "Missing authorization code.")
		return
	}

	user, err := app.models.Users.GetForToken(state, data.ScopeOAuthState, true)
	if err != nil {
		app.writeOAuthResult(w, http.StatusBadRequest, false, "This connection link has expired. Please try again from the app.")
		return
	}

	token, displayName, err := provider.Exchange(r.Context(), code)
	if err != nil {
		app.logError(r, err)
		app.writeOAuthResult(w, http.StatusBadGateway, false, "Google did not confirm the connection. Please try again.")
		return
	}

	account := data.CalendarAccount{UserID: user.ID, Provider: provider.Name(), DisplayName: displayName}
	if err := app.models.CalendarAccounts.Upsert(&account); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	sealed, err := app.sealSecret(oauth2Secret{RefreshToken: token.RefreshToken})
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	cred := data.CalendarAccountCredential{CalendarAccountID: account.ID, AuthType: data.CalendarAuthTypeOAuth2, EncryptedSecret: sealed}
	if err := app.models.CalendarAccounts.PutCredential(cred); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.SendToWsUser(user.ID, app.retriveWebSocket("halendar"), envelope{"type": "calendar_accounts", "refresh": true})
	app.writeOAuthResult(w, http.StatusOK, true, displayName+" is now connected.")
}

// connectCaldavCalendarHandler connects a CalDAV calendar (Apple iCloud, a
// groupware's calendar, ...) directly from a submitted URL/username/app password --
// no redirect flow needed. The credentials are tested before anything is stored.
func (app *app) connectCaldavCalendarHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	var input struct {
		URL      string `json:"url"`
		User     string `json:"user"`
		Pass     string `json:"pass"`
		Calendar string `json:"calendar"`
		Timezone string `json:"timezone"`
	}
	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	v.Check(input.URL != "", "url", "must be provided")
	v.Check(input.User != "", "user", "must be provided")
	v.Check(input.Pass != "", "pass", "must be provided")
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	cfg := calendar.Config{URL: input.URL, User: input.User, Pass: input.Pass, Calendar: input.Calendar, Timezone: input.Timezone}
	client := calendar.New(cfg)

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	if err := client.Test(ctx); err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("could not connect: %w", err))
		return
	}

	configJSON, err := json.Marshal(caldavConfig{URL: input.URL, User: input.User, Calendar: input.Calendar, Timezone: input.Timezone})
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	displayName := input.User
	if input.Calendar != "" {
		displayName = input.Calendar + " (" + input.User + ")"
	}
	account := data.CalendarAccount{UserID: user.ID, Provider: "caldav", DisplayName: displayName, Config: configJSON}
	if err := app.models.CalendarAccounts.Upsert(&account); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	sealed, err := app.sealSecret(map[string]string{"password": input.Pass})
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	cred := data.CalendarAccountCredential{CalendarAccountID: account.ID, AuthType: data.CalendarAuthTypeBasic, EncryptedSecret: sealed}
	if err := app.models.CalendarAccounts.PutCredential(cred); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.SendToWsUser(user.ID, app.retriveWebSocket("halendar"), envelope{"type": "calendar_accounts", "refresh": true})
	if err := app.writeJSON(w, http.StatusCreated, envelope{"account": account}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// deleteCalendarAccountHandler disconnects a calendar: its credentials are removed
// along with it (ON DELETE CASCADE).
func (app *app) deleteCalendarAccountHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	id, err := uuid.Parse(app.readStringParam(r, "id"))
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid id parameter"))
		return
	}

	if err := app.models.CalendarAccounts.Delete(id, user.ID); err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	if err := app.writeJSON(w, http.StatusOK, envelope{"status": "deleted"}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// linkGoogleCalendarFromEmail creates (or refreshes) a linked Google calendar account
// from a Gmail connection's own OAuth grant -- gmailScopes already includes Calendar
// access, so no separate consent is needed for the common case of one Google account
// used for both.
func (app *app) linkGoogleCalendarFromEmail(userID uuid.UUID, email string, token emailimport.Token) error {
	account := data.CalendarAccount{UserID: userID, Provider: "google", DisplayName: email}
	if err := app.models.CalendarAccounts.Upsert(&account); err != nil {
		return fmt.Errorf("upserting calendar account: %w", err)
	}
	sealed, err := app.sealSecret(oauth2Secret{RefreshToken: token.RefreshToken})
	if err != nil {
		return fmt.Errorf("sealing credential: %w", err)
	}
	cred := data.CalendarAccountCredential{CalendarAccountID: account.ID, AuthType: data.CalendarAuthTypeOAuth2, EncryptedSecret: sealed}
	if err := app.models.CalendarAccounts.PutCredential(cred); err != nil {
		return fmt.Errorf("storing credential: %w", err)
	}
	return nil
}

// caldavSource adapts *calendar.Client (a CalDAV connection) to calendarimport.Source.
type caldavSource struct{ client *calendar.Client }

func (s caldavSource) Busy(ctx context.Context, start, end time.Time) (bool, error) {
	busy, _, err := s.client.Busy(ctx, start, end)
	return busy, err
}

func (s caldavSource) Timezone() *time.Location { return s.client.Timezone() }

// calendarSourceFor builds a calendarimport.Source for a connected calendar account,
// refreshing its OAuth token or decrypting its CalDAV password as needed.
func (app *app) calendarSourceFor(ctx context.Context, account data.CalendarAccount) (calendarimport.Source, error) {
	cred, err := app.models.CalendarAccounts.GetCredential(account.ID)
	if err != nil {
		return nil, fmt.Errorf("loading credential: %w", err)
	}

	switch account.Provider {
	case "caldav":
		var cfg caldavConfig
		if err := json.Unmarshal(account.Config, &cfg); err != nil {
			return nil, fmt.Errorf("parsing config: %w", err)
		}
		var secret struct {
			Password string `json:"password"`
		}
		if err := app.openSecret(cred.EncryptedSecret, &secret); err != nil {
			return nil, fmt.Errorf("decrypting credential: %w", err)
		}
		client := calendar.New(calendar.Config{URL: cfg.URL, User: cfg.User, Pass: secret.Password, Calendar: cfg.Calendar, Timezone: cfg.Timezone})
		return caldavSource{client: client}, nil

	default:
		provider, ok := app.calendarProviders.Get(account.Provider)
		if !ok {
			return nil, fmt.Errorf("provider %q not configured", account.Provider)
		}
		var secret oauth2Secret
		if err := app.openSecret(cred.EncryptedSecret, &secret); err != nil {
			return nil, fmt.Errorf("decrypting credential: %w", err)
		}
		token, err := provider.Refresh(ctx, secret.RefreshToken)
		if err != nil {
			return nil, fmt.Errorf("refreshing token: %w", err)
		}
		return provider.Source(token), nil
	}
}
