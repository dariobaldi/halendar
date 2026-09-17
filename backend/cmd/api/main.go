package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/dariobaldi/halendar_back/internal/calendarimport"
	"github.com/dariobaldi/halendar_back/internal/data"
	"github.com/dariobaldi/halendar_back/internal/emailimport"
	"github.com/dariobaldi/halendar_back/internal/mailer"
	"github.com/dariobaldi/halendar_back/internal/ollama"
	"github.com/dariobaldi/halendar_back/internal/push"
	"github.com/dariobaldi/halendar_back/internal/secretbox"
	"github.com/dariobaldi/halendar_back/internal/vcs"
	"github.com/dariobaldi/halendar_back/internal/websocket"
	_ "github.com/lib/pq"
	"golang.org/x/time/rate"
	"halendar/calendar"
	"halendar/mail"
)

var (
	version = vcs.Version()
)

type config struct {
	port        int
	env         string
	frontendURL string // where an OAuth connect flow sends the browser once it's done
	cors        struct {
		trustedOrigins []string
	}
	db struct {
		dsn          string
		dsnDev       string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}
	limiter struct {
		requestPerSecond float64
		burstLimit       int
		enabled          bool
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	ollama struct {
		baseURL string
		model   string
	}
	claude struct {
		model string // which Claude model a user's own API key is sent to, see internal/claude
	}
	gemini struct {
		model string // which Gemini model a user's own API key is sent to, see internal/gemini
	}
	push struct {
		projectID          string
		serviceAccountFile string
	}
	security struct {
		encryptionKey string // base64 AES-256 key; secretbox.ParseKey decodes it
	}
	google struct {
		clientID            string
		clientSecret        string
		redirectURL         string // .../email-accounts/gmail/callback (also covers Calendar, bundled into the same grant)
		calendarRedirectURL string // .../calendar-accounts/google/callback (standalone Calendar-only connect)
	}
	emailSync struct {
		interval time.Duration
	}
}

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type app struct {
	clientsIPs        map[string]*client
	config            config
	logger            *slog.Logger
	fileLogger        *slog.Logger
	mailer            mailer.Mailer
	mailbox           *mail.Mailbox
	calendar          *calendar.Client
	ollama            *ollama.Client
	push              *push.Client
	models            data.Models
	emailProviders    emailimport.Registry
	calendarProviders calendarimport.Registry
	encryptionKey     secretbox.Key
	mu                sync.Mutex
	websockets        map[string]*websocket.Hub
	wg                sync.WaitGroup
	summaryUpdateCh   chan struct{}
	analysisSem       chan struct{} // caps concurrent analyzeEmailMessage calls, see its doc comment
}

func main() {
	var cfg config
	cfg.GetVariables()

	// Override values with flags
	flag.StringVar(&cfg.env, "env", cfg.env, "Environment (development|staging|production)")
	flag.IntVar(&cfg.port, "port", cfg.port, "API server port")
	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	flag.StringVar(&cfg.db.dsn, "dsn", cfg.db.dsn, "PostgreSQL DSN")

	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", cfg.db.maxOpenConns, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", cfg.db.maxIdleConns, "PostgreSQL max idle connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", cfg.db.maxIdleTime, "PostgreSQL max connection idle time")

	flag.Float64Var(&cfg.limiter.requestPerSecond, "limiter-request-per-second", cfg.limiter.requestPerSecond, "Rate limiter requests per second")
	flag.IntVar(&cfg.limiter.burstLimit, "limiter-burst-limit", cfg.limiter.burstLimit, "Rate limiter burst limit")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", cfg.limiter.enabled, "Rate limiter enabled")

	flag.StringVar(&cfg.smtp.host, "smtp-host", cfg.smtp.host, "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", cfg.smtp.port, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", cfg.smtp.username, "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", cfg.smtp.password, "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", cfg.smtp.sender, "SMTP sender")

	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	if *displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		os.Exit(0)
	}

	if cfg.env != "production" {
		cfg.db.dsn = cfg.db.dsnDev
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logFile, err := os.OpenFile("data/halendar_log.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	var loggerFile *slog.Logger
	if err != nil {
		logger.Info("failed opening logFile: " + err.Error())
		loggerFile = logger
	} else {
		defer logFile.Close()
		loggerFile = slog.New(slog.NewJSONHandler(logFile, nil))
	}

	db, err := openDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	defer db.Close()
	logger.Info("database connection pool established")

	var fcmServiceAccount []byte
	if cfg.push.serviceAccountFile != "" {
		fcmServiceAccount, err = os.ReadFile(cfg.push.serviceAccountFile)
		if err != nil {
			logger.Error("could not read FCM_SERVICE_ACCOUNT_FILE: " + err.Error())
		}
	}

	// A missing or invalid ENCRYPTION_KEY disables connecting email accounts (there
	// would be nothing safe to encrypt their credentials with) rather than the whole
	// API, in keeping with how a missing mail/calendar/push config degrades below.
	encryptionKey, keyErr := secretbox.ParseKey(cfg.security.encryptionKey)
	emailProviders := emailimport.Registry{}
	calendarProviders := calendarimport.Registry{}
	switch {
	case keyErr != nil:
		logger.Error("invalid or missing ENCRYPTION_KEY: " + keyErr.Error() + " (generate one with: openssl rand -base64 32); connecting email/calendar accounts is disabled")
	case cfg.google.clientID == "" || cfg.google.clientSecret == "":
		logger.Info("GOOGLE_OAUTH_CLIENT_ID/SECRET not set: Gmail and Google Calendar account connection is disabled")
	default:
		emailProviders["gmail"] = emailimport.NewGmailProvider(cfg.google.clientID, cfg.google.clientSecret, cfg.google.redirectURL)
		if cfg.google.calendarRedirectURL != "" {
			calendarProviders["google"] = calendarimport.NewGoogleProvider(cfg.google.clientID, cfg.google.clientSecret, cfg.google.calendarRedirectURL)
		} else {
			logger.Info("GOOGLE_OAUTH_CALENDAR_REDIRECT_URL not set: standalone Google Calendar connection is disabled (connecting Gmail still links a calendar automatically)")
		}
	}

	app := &app{
		clientsIPs:        make(map[string]*client),
		config:            cfg,
		logger:            logger,
		fileLogger:        loggerFile,
		models:            data.NewModels(db),
		mailer:            mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
		mailbox:           mail.New(mail.ConfigFromEnv()),
		calendar:          calendar.New(calendar.ConfigFromEnv()),
		ollama:            ollama.New(cfg.ollama.baseURL, cfg.ollama.model),
		push:              push.New(push.Config{ProjectID: cfg.push.projectID, ServiceAccountJSON: fcmServiceAccount}),
		websockets:        make(map[string]*websocket.Hub),
		emailProviders:    emailProviders,
		calendarProviders: calendarProviders,
		encryptionKey:     encryptionKey,
		analysisSem:       make(chan struct{}, analysisConcurrency),
	}

	app.backgroudProcess()
	app.emailSyncLoop()
	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// Set configuration values.
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	db.SetConnMaxIdleTime(cfg.db.maxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check that the connection is successful
	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
