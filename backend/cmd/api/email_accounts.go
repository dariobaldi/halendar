package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dariobaldi/halendar_back/internal/data"
	"github.com/google/uuid"
)

// oauthStateTTL is how long a "connect" link stays valid before the user must retry.
const oauthStateTTL = 10 * time.Minute

// oauth2Secret is the JSON shape encrypted into email_account_credentials for
// AuthTypeOAuth2. Only the refresh token needs to survive between syncs -- a fresh
// access token is minted from it right before each use.
type oauth2Secret struct {
	RefreshToken string `json:"refresh_token"`
}

// listEmailAccountsHandler lists the accounts the current user has connected.
func (app *app) listEmailAccountsHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	accounts, err := app.models.EmailAccounts.GetForUser(user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if err := app.writeJSON(w, http.StatusOK, envelope{"accounts": accounts}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// connectEmailAccountHandler returns the URL the client should open in the system
// browser (not an embedded webview -- Google's OAuth policy blocks signing in from
// one) to start that provider's OAuth2 consent flow.
func (app *app) connectEmailAccountHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	providerName := app.readStringParam(r, "provider")
	provider, ok := app.emailProviders.Get(providerName)
	if !ok {
		app.badRequestResponse(w, r, fmt.Errorf("unknown or unconfigured provider %q", providerName))
		return
	}

	// The state token both proves this request came from an authenticated session and
	// carries no data of its own -- the callback looks the user back up from it,
	// consuming it so the link can't be replayed.
	state, err := app.models.Tokens.New(user.ID, oauthStateTTL, data.ScopeOAuthState)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if err := app.writeJSON(w, http.StatusOK, envelope{"auth_url": provider.AuthCodeURL(state.Plaintext)}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// emailAccountCallbackHandler is where the provider redirects the user's browser back
// to after consent. It has no Authorization header -- the state token is what ties the
// request back to the user who started it. Not JSON: it renders a small page for the
// browser tab the user opened for the flow.
func (app *app) emailAccountCallbackHandler(w http.ResponseWriter, r *http.Request) {
	providerName := app.readStringParam(r, "provider")
	provider, ok := app.emailProviders.Get(providerName)
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

	token, email, err := provider.Exchange(r.Context(), code)
	if err != nil {
		app.logError(r, err)
		app.writeOAuthResult(w, http.StatusBadGateway, false, "Google did not confirm the connection. Please try again.")
		return
	}

	account := data.EmailAccount{UserID: user.ID, Provider: provider.Name(), EmailAddress: email}
	if err := app.models.EmailAccounts.Upsert(&account); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	sealed, err := app.sealSecret(oauth2Secret{RefreshToken: token.RefreshToken})
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	cred := data.EmailAccountCredential{EmailAccountID: account.ID, AuthType: data.AuthTypeOAuth2, EncryptedSecret: sealed}
	if err := app.models.EmailAccounts.PutCredential(cred); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Gmail's OAuth scope list includes Calendar access (see gmailScopes), so this one
	// grant covers both -- link the same Google account as a calendar too, rather than
	// making the user connect it again separately.
	if err := app.linkGoogleCalendarFromEmail(user.ID, email, token); err != nil {
		app.logger.Error("email callback: linking calendar account: " + err.Error())
	}

	// Establish the UID baseline right away rather than waiting for the next tick (up
	// to a minute away): NewSince(ctx, 0) never imports the existing backlog on this
	// first pass, but running it now closes the window where mail arriving before the
	// next tick would otherwise be silently folded into the baseline and missed.
	app.background(func() { app.syncEmailAccount(account) })

	app.SendToWsUser(user.ID, app.retriveWebSocket("halendar"), envelope{"type": "email_accounts", "refresh": true})
	app.writeOAuthResult(w, http.StatusOK, true, email+" is now connected.")
}

// deleteEmailAccountHandler disconnects an account: its credentials and imported
// message history are removed along with it (ON DELETE CASCADE).
func (app *app) deleteEmailAccountHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	id, err := uuid.Parse(app.readStringParam(r, "id"))
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid id parameter"))
		return
	}

	if err := app.models.EmailAccounts.Delete(id, user.ID); err != nil {
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

// writeOAuthResult renders the small page shown in the system browser once the
// provider redirects back, since consent happens outside the app (Google disallows
// signing in from an embedded webview). It shows the outcome briefly, then sends the
// browser on to the web app -- on a phone this still leaves the user in the browser
// rather than back in the native app (see the doc comment on FrontendURL), but at
// least it's not left stranded on a bare confirmation page.
func (app *app) writeOAuthResult(w http.ResponseWriter, status int, success bool, message string) {
	icon, title := "✓", "Connected"
	if !success {
		icon, title = "✕", "Couldn't connect"
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	fmt.Fprintf(w, `<!doctype html>
<html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<title>%s</title>
<meta http-equiv="refresh" content="2;url=%s">
<style>
body{font-family:-apple-system,system-ui,sans-serif;display:flex;align-items:center;justify-content:center;height:100vh;margin:0;background:#f4f4f5;color:#18181b}
.card{text-align:center;padding:2rem;max-width:22rem}
.icon{font-size:2.5rem;margin-bottom:.5rem}
a{color:inherit}
</style></head>
<body><div class="card"><div class="icon">%s</div><h2>%s</h2><p>%s</p><p><a href="%s">Continue</a></p></div></body>
<script>setTimeout(function(){ location.replace(%q); }, 2000);</script>
</html>
`, title, app.config.frontendURL, icon, title, message, app.config.frontendURL, app.config.frontendURL)
}
