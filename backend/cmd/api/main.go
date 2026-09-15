package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"sync"

	"github.com/dariobaldi/halendar/internal/database"
	"github.com/dariobaldi/halendar/internal/env"
	"github.com/dariobaldi/halendar/internal/version"

	"github.com/lmittmann/tint"
)

func main() {
	logger := slog.New(tint.NewTextHandler(os.Stdout, &tint.Options{Level: slog.LevelDebug}))

	err := run(logger)
	if err != nil {
		trace := string(debug.Stack())
		logger.Error(err.Error(), "trace", trace)
		os.Exit(1)
	}
}

type config struct {
	baseURL   string
	httpPort  int
	basicAuth struct {
		username       string
		hashedPassword string
	}
	autoHTTPS struct {
		domain  string
		email   string
		staging bool
	}
	db struct {
		dsn         string
		automigrate bool
	}
}

type application struct {
	config config
	db     *database.DB
	logger *slog.Logger
	wg     sync.WaitGroup
}

func run(logger *slog.Logger) error {
	var cfg config

	cfg.baseURL = env.GetString("BASE_URL", "http://localhost:5051")
	cfg.httpPort = env.GetInt("HTTP_PORT", 5051)
	cfg.autoHTTPS.domain = env.GetString("AUTO_HTTPS_DOMAIN", "")
	cfg.autoHTTPS.email = env.GetString("AUTO_HTTPS_EMAIL", "admin@example.com")
	cfg.autoHTTPS.staging = env.GetBool("AUTO_HTTPS_STAGING", false)
	cfg.basicAuth.username = env.GetString("BASIC_AUTH_USERNAME", "admin")
	cfg.basicAuth.hashedPassword = env.GetString("BASIC_AUTH_HASHED_PASSWORD", "$2a$10$jRb2qniNcoCyQM23T59RfeEQUbgdAXfR6S0scynmKfJa5Gj3arGJa")
	cfg.db.dsn = env.GetString("DB_DSN", "user:pass@localhost:5432/db")
	cfg.db.automigrate = env.GetBool("DB_AUTOMIGRATE", true)

	showVersion := flag.Bool("version", false, "display version and exit")

	flag.Parse()

	if *showVersion {
		fmt.Printf("version: %s\n", version.Get())
		return nil
	}

	db, err := database.New(cfg.db.dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	if cfg.db.automigrate {
		err = db.MigrateUp()
		if err != nil {
			return err
		}
	}

	app := &application{
		config: cfg,
		db:     db,
		logger: logger,
	}

	if cfg.autoHTTPS.domain != "" {
		return app.serveAutoHTTPS()
	}

	return app.serveHTTP()
}
