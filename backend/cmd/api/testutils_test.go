package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"net/url"

	"github.com/dariobaldi/halendar/internal/database"
)

func newTestApplication(t *testing.T) *application {
	app := new(application)

	app.config.baseURL = "https://www.example.com"

	app.logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	app.db = newTestDB(t)

	return app
}

func newTestDB(t *testing.T) *database.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DB_DSN")

	if dsn == "" {
		t.Fatal("TEST_DB_DSN environment variable must be set in the format user:pass@localhost:port/db")
	}

	schemaName := fmt.Sprintf("test_schema_%d", time.Now().UnixNano())
	dsn, err := appendParam(dsn, "search_path", schemaName)
	if err != nil {
		t.Fatal(err)
	}

	db, err := database.New(dsn)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		defer db.Close()

		_, err = db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))
		if err != nil {
			t.Error(err)
		}
	})

	_, err = db.Exec(fmt.Sprintf("CREATE SCHEMA %s", schemaName))
	if err != nil {
		t.Fatal(err)
	}

	err = db.MigrateUp()
	if err != nil {
		t.Fatal(err)
	}

	return db
}

func appendParam(dsn, key, value string) (string, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return "", fmt.Errorf("failed to parse DSN: %w", err)
	}

	q := u.Query()
	q.Set(key, value)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func newTestRequest(t *testing.T, method, path string, data map[string]any) *http.Request {
	if data == nil {
		req, err := http.NewRequest(method, path, nil)
		if err != nil {
			t.Fatal(err)
		}
		return req
	}

	js, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(method, path, bytes.NewBuffer(js))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")
	return req
}

type testResponse struct {
	*http.Response
	BodyFields map[string]any
}

func send(t *testing.T, req *http.Request, h http.Handler) testResponse {
	if len(req.PostForm) > 0 {
		body := req.PostForm.Encode()
		req.Body = io.NopCloser(strings.NewReader(body))
		req.ContentLength = int64(len(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	res := rec.Result()

	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	fields := map[string]any{}

	if len(resBody) > 0 {
		err := json.Unmarshal(resBody, &fields)
		if err != nil {
			t.Fatal(err)
		}
	}

	return testResponse{
		Response:   res,
		BodyFields: fields,
	}
}
