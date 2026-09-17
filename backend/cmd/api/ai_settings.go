package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dariobaldi/halendar_back/internal/claude"
	"github.com/dariobaldi/halendar_back/internal/data"
	"github.com/dariobaldi/halendar_back/internal/secretbox"
	"github.com/google/uuid"
)

// getAISettingsHandler returns the current user's AI provider choice and whether a
// Claude key is on file (never the key itself).
func (app *app) getAISettingsHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	settings, err := app.models.AISettings.Get(user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if err := app.writeJSON(w, http.StatusOK, envelope{"ai_settings": settings}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// connectClaudeHandler imports (or replaces) the user's Anthropic API key. The key is
// verified with a minimal live request before being stored, so a typo or revoked key
// is caught immediately rather than silently failing the next background analysis
// pass. Doesn't itself activate Claude -- see setAIProviderHandler.
func (app *app) connectClaudeHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	var input struct {
		APIKey string `json:"api_key"`
	}
	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if input.APIKey == "" {
		app.badRequestResponse(w, r, errors.New(`"api_key" field missing`))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	if err := claude.VerifyKey(ctx, input.APIKey, app.config.claude.model); err != nil {
		app.errorResponse(w, r, http.StatusUnprocessableEntity, "Could not verify this API key with Claude: "+err.Error())
		return
	}

	encrypted, err := secretbox.Seal(app.encryptionKey, []byte(input.APIKey))
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if err := app.models.AISettings.SetAPIKey(user.ID, encrypted); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.respondAISettings(w, r, user.ID)
}

// disconnectClaudeHandler removes the stored key and falls back to the local model.
func (app *app) disconnectClaudeHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	if err := app.models.AISettings.ClearAPIKey(user.ID); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	app.respondAISettings(w, r, user.ID)
}

// setAIProviderHandler switches which model future analysis uses for this user.
// Rejected if switching to Claude without a key on file.
func (app *app) setAIProviderHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	var input struct {
		Provider string `json:"provider"`
	}
	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if input.Provider != data.AIProviderOllama && input.Provider != data.AIProviderClaude {
		app.badRequestResponse(w, r, fmt.Errorf("unknown provider %q", input.Provider))
		return
	}

	if input.Provider == data.AIProviderClaude {
		settings, err := app.models.AISettings.Get(user.ID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		if !settings.HasAPIKey {
			app.badRequestResponse(w, r, errors.New("connect a Claude API key before activating it"))
			return
		}
	}

	if err := app.models.AISettings.SetProvider(user.ID, input.Provider); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	app.respondAISettings(w, r, user.ID)
}

func (app *app) respondAISettings(w http.ResponseWriter, r *http.Request, userID uuid.UUID) {
	settings, err := app.models.AISettings.Get(userID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if err := app.writeJSON(w, http.StatusOK, envelope{"ai_settings": settings}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
