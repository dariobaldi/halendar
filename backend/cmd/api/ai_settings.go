package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dariobaldi/halendar_back/internal/claude"
	"github.com/dariobaldi/halendar_back/internal/data"
	"github.com/dariobaldi/halendar_back/internal/gemini"
	"github.com/dariobaldi/halendar_back/internal/secretbox"
	"github.com/google/uuid"
)

// getAISettingsHandler returns the current user's AI provider choice and whether a
// Claude/Gemini key is on file for each (never the keys themselves).
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

// verifyAIKey checks a candidate API key against the real provider before it's
// stored, so a typo or revoked key is caught immediately rather than silently
// failing the next background analysis pass.
func verifyAIKey(ctx context.Context, provider, apiKey, model string) error {
	switch provider {
	case data.AIProviderClaude:
		return claude.VerifyKey(ctx, apiKey, model)
	case data.AIProviderGemini:
		return gemini.VerifyKey(ctx, apiKey, model)
	default:
		return fmt.Errorf("%q is not a key-bearing AI provider", provider)
	}
}

// modelFor returns the model name provider's key is sent to, from config.
func (app *app) modelFor(provider string) string {
	if provider == data.AIProviderGemini {
		return app.config.gemini.model
	}
	return app.config.claude.model
}

// connectAIKeyHandler imports (or replaces) the user's API key for :provider (claude
// or gemini). Doesn't itself activate that provider -- see setAIProviderHandler.
func (app *app) connectAIKeyHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	provider := app.readStringParam(r, "provider")
	if provider != data.AIProviderClaude && provider != data.AIProviderGemini {
		app.badRequestResponse(w, r, fmt.Errorf("unknown provider %q", provider))
		return
	}

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
	if err := verifyAIKey(ctx, provider, input.APIKey, app.modelFor(provider)); err != nil {
		app.errorResponse(w, r, http.StatusUnprocessableEntity, fmt.Sprintf("Could not verify this API key with %s: %s", provider, err.Error()))
		return
	}

	encrypted, err := secretbox.Seal(app.encryptionKey, []byte(input.APIKey))
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if err := app.models.AISettings.SetAPIKey(user.ID, provider, encrypted); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.respondAISettings(w, r, user.ID)
}

// disconnectAIKeyHandler removes the stored key for :provider, falling back to the
// local model if it was the active one.
func (app *app) disconnectAIKeyHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	provider := app.readStringParam(r, "provider")
	if provider != data.AIProviderClaude && provider != data.AIProviderGemini {
		app.badRequestResponse(w, r, fmt.Errorf("unknown provider %q", provider))
		return
	}

	if err := app.models.AISettings.ClearAPIKey(user.ID, provider); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	app.respondAISettings(w, r, user.ID)
}

// setAIProviderHandler switches which model future analysis uses for this user.
// Rejected if switching to Claude or Gemini without a key on file for it.
func (app *app) setAIProviderHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	var input struct {
		Provider string `json:"provider"`
	}
	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if input.Provider != data.AIProviderOllama && input.Provider != data.AIProviderClaude && input.Provider != data.AIProviderGemini {
		app.badRequestResponse(w, r, fmt.Errorf("unknown provider %q", input.Provider))
		return
	}

	if input.Provider != data.AIProviderOllama {
		settings, err := app.models.AISettings.Get(user.ID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		hasKey := (input.Provider == data.AIProviderClaude && settings.HasClaudeKey) ||
			(input.Provider == data.AIProviderGemini && settings.HasGeminiKey)
		if !hasKey {
			app.badRequestResponse(w, r, fmt.Errorf("connect a %s API key before activating it", input.Provider))
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
