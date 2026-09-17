// Package gemini is a minimal client for Google's Generative Language API, used as
// another alternative to the local Ollama model (internal/ollama) alongside
// internal/claude, when a user has connected their own Gemini API key and activated
// it. Exposes the same Generate/GenerateDeterministic shape as the other two clients
// so cmd/api's email-analysis pipeline can use whichever one interchangeably -- see
// aiClient in cmd/api/email_import.go.
package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultBaseURL = "https://generativelanguage.googleapis.com"
	// maxOutputTokens caps the reply, not the prompt -- see internal/claude's
	// equivalent constant for why.
	maxOutputTokens = 1024
)

// Client talks to the Gemini API using one caller-supplied API key and model. Built
// fresh per request from a user's own decrypted key, same as internal/claude.Client --
// see cmd/api's aiClientFor.
type Client struct {
	baseURL string
	apiKey  string
	model   string
	http    *http.Client
}

// New returns a Client for apiKey using model (e.g. "gemini-3.6-flash").
func New(apiKey, model string) *Client {
	return &Client{
		baseURL: defaultBaseURL,
		apiKey:  apiKey,
		model:   model,
		http:    &http.Client{Timeout: 2 * time.Minute},
	}
}

// Model returns the model name this client sends prompts to.
func (c *Client) Model() string { return c.model }

type part struct {
	Text string `json:"text"`
}

type content struct {
	Role  string `json:"role"`
	Parts []part `json:"parts"`
}

type generationConfig struct {
	Temperature     *float64 `json:"temperature,omitempty"`
	MaxOutputTokens int      `json:"maxOutputTokens"`
}

type generateContentRequest struct {
	Contents         []content        `json:"contents"`
	GenerationConfig generationConfig `json:"generationConfig"`
}

type generateContentResponse struct {
	Candidates []struct {
		Content content `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Generate sends prompt to the model and returns its full response text, using the
// model's default sampling.
func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	return c.generate(ctx, prompt, nil)
}

// GenerateDeterministic is like Generate but disables sampling (temperature 0), for
// tasks that need consistent, well-formed output rather than varied phrasing -- see
// the equivalent method on internal/ollama.Client for why that matters here.
func (c *Client) GenerateDeterministic(ctx context.Context, prompt string) (string, error) {
	temp := 0.0
	return c.generate(ctx, prompt, &temp)
}

func (c *Client) generate(ctx context.Context, prompt string, temperature *float64) (string, error) {
	body, err := json.Marshal(generateContentRequest{
		Contents: []content{{Role: "user", Parts: []part{{Text: prompt}}}},
		GenerationConfig: generationConfig{
			Temperature:     temperature,
			MaxOutputTokens: maxOutputTokens,
		},
	})
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent", c.baseURL, c.model)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("gemini: could not reach %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("gemini: reading response: %w", err)
	}

	var out generateContentResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", fmt.Errorf("gemini: invalid response (%s): %w", resp.Status, err)
	}
	if resp.StatusCode != http.StatusOK {
		if out.Error != nil {
			return "", fmt.Errorf("gemini: %s", out.Error.Message)
		}
		return "", fmt.Errorf("gemini: %s", resp.Status)
	}
	if len(out.Candidates) == 0 || len(out.Candidates[0].Content.Parts) == 0 {
		return "", nil
	}
	return out.Candidates[0].Content.Parts[0].Text, nil
}

// VerifyKey makes one minimal, cheap request to confirm apiKey is valid for model.
// Used when a user connects a new key, so a typo or revoked key is caught immediately
// instead of silently falling back or failing on the next background analysis pass.
func VerifyKey(ctx context.Context, apiKey, model string) error {
	_, err := New(apiKey, model).GenerateDeterministic(ctx, "Reply with only the word OK.")
	return err
}
