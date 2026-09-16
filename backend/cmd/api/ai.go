package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// promptHandler sends an arbitrary prompt straight to the configured Ollama model and
// returns its response. Useful to check the "ollama" container is reachable and serving.
func (app *app) promptHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Prompt string `json:"prompt"`
	}
	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if input.Prompt == "" {
		app.badRequestResponse(w, r, errors.New(`"prompt" field missing`))
		return
	}

	response, err := app.ollama.Generate(r.Context(), input.Prompt)
	if err != nil {
		app.errorResponse(w, r, http.StatusServiceUnavailable, err.Error())
		return
	}

	if err := app.writeJSON(w, http.StatusOK, envelope{
		"model":    app.ollama.Model(),
		"prompt":   input.Prompt,
		"response": response,
	}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// exampleEmail and exampleInstruction are the defaults testPromptHandler falls back to
// when the caller doesn't supply their own, so the endpoint works out of the box.
const (
	exampleEmail = `From: Claire Martin <claire@example.com>
Subject: Project sync

Hi team,

Could we sync on the roadmap this week? I'm free either:
2026-09-18 14:00-14:30
2026-09-19 10:00-10:30

Let me know what works.

Claire`

	exampleInstruction = `You are an assistant that reads emails and helps schedule meetings.
Read the email below and:
1. List every proposed date/time slot you find.
2. Write a short, friendly reply confirming the first slot works.`
)

// testPromptHandler demonstrates the app's actual use case: an email plus an instruction
// prompt, combined and sent to Ollama, returning its response. Both "email" and "prompt"
// in the (optional) JSON body override the built-in example, so it also works with no
// body at all, e.g. `curl -X POST .../v1/ai/test`.
func (app *app) testPromptHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, BodyMaxBytes))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	input := struct {
		Email  string `json:"email"`
		Prompt string `json:"prompt"`
	}{Email: exampleEmail, Prompt: exampleInstruction}

	if trimmed := bytes.TrimSpace(body); len(trimmed) > 0 {
		var override struct {
			Email  string `json:"email"`
			Prompt string `json:"prompt"`
		}
		if err := json.Unmarshal(trimmed, &override); err != nil {
			app.badRequestResponse(w, r, err)
			return
		}
		if override.Email != "" {
			input.Email = override.Email
		}
		if override.Prompt != "" {
			input.Prompt = override.Prompt
		}
	}

	prompt := input.Prompt + "\n\nEmail:\n\"\"\"\n" + input.Email + "\n\"\"\""

	response, err := app.ollama.Generate(r.Context(), prompt)
	if err != nil {
		app.errorResponse(w, r, http.StatusServiceUnavailable, err.Error())
		return
	}

	if err := app.writeJSON(w, http.StatusOK, envelope{
		"model":    app.ollama.Model(),
		"email":    input.Email,
		"prompt":   prompt,
		"response": response,
	}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
