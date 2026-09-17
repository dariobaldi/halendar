package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func (cfg *config) GetVariables() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cfg.env = getEnvString("ENV", "development")

	cfg.port = getEnvInt("PORT", 4000)
	cfg.frontendURL = getEnvString("FRONTEND_URL", "https://frontend.domain")
	cfg.cors.trustedOrigins = getEnvSliceString("TRUSTED_ORIGINS", []string{"frontend.domain"})

	cfg.db.dsn = getEnvString("DB_DSN", "postgres://user:password@localhost/database?sslmode=disable")
	cfg.db.dsnDev = getEnvString("DB_DSN_DEV", "postgres://user:password@localhost/database?sslmode=disable")

	cfg.db.maxOpenConns = getEnvInt("DB_MAX_OPEN_CONNS", 25)
	cfg.db.maxIdleConns = getEnvInt("DB_MAX_IDLE_CONNS", 25)

	idle_time := getEnvString("DB_MAX_IDLE_TIME", "15m")
	cfg.db.maxIdleTime, err = time.ParseDuration(idle_time)
	if err != nil {
		log.Fatal(err)
		cfg.db.maxIdleTime = 15 * time.Minute
	}

	cfg.limiter.requestPerSecond = getEnvFloat64("REQUEST_PER_SECOND", 2)
	cfg.limiter.burstLimit = getEnvInt("BURST_LIMIT", 4)
	cfg.limiter.enabled = getEnvBool("RATE_LIMITER", true)

	cfg.smtp.host = getEnvString("SMTP_HOST", "localhost")
	cfg.smtp.port = getEnvInt("SMTP_PORT", 25)
	cfg.smtp.username = getEnvString("SMTP_USERNAME", "username")
	cfg.smtp.password = getEnvString("SMTP_PASSWORD", "password")
	cfg.smtp.sender = getEnvString("SMTP_SENDER", "Halendar <no-reply@yourdomain.com>")

	cfg.ollama.baseURL = getEnvString("OLLAMA_BASE_URL", "http://ollama:11434")
	cfg.ollama.model = getEnvString("OLLAMA_MODEL", "gemma3:1b")

	// Model used for a user's own Claude API key, once they've connected and
	// activated one (see internal/claude and cmd/api/ai_settings.go) -- unrelated to
	// which Anthropic account or plan owns the key, just which model it's billed to.
	cfg.claude.model = getEnvString("CLAUDE_MODEL", "claude-haiku-4-5-20251001")

	// Same idea for a user's own Gemini API key (see internal/gemini).
	cfg.gemini.model = getEnvString("GEMINI_MODEL", "gemini-3.6-flash")

	cfg.push.projectID = getEnvString("FCM_PROJECT_ID", "")
	cfg.push.serviceAccountFile = getEnvString("FCM_SERVICE_ACCOUNT_FILE", "")

	cfg.security.encryptionKey = getEnvString("ENCRYPTION_KEY", "")

	cfg.google.clientID = getEnvString("GOOGLE_OAUTH_CLIENT_ID", "")
	cfg.google.clientSecret = getEnvString("GOOGLE_OAUTH_CLIENT_SECRET", "")
	cfg.google.redirectURL = getEnvString("GOOGLE_OAUTH_REDIRECT_URL", "")
	cfg.google.calendarRedirectURL = getEnvString("GOOGLE_OAUTH_CALENDAR_REDIRECT_URL", "")

	syncInterval := getEnvString("EMAIL_SYNC_INTERVAL", "1m")
	cfg.emailSync.interval, err = time.ParseDuration(syncInterval)
	if err != nil {
		cfg.emailSync.interval = time.Minute
	}
}

// GETENV HELPERS

func getEnvInt(key string, defaultValue int) int {
	s := os.Getenv(key)
	if s == "" {
		return defaultValue
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}
	return v
}

func getEnvBool(key string, defaultValue bool) bool {
	s := os.Getenv(key)
	if s == "" {
		return defaultValue
	}
	v, err := strconv.ParseBool(s)
	if err != nil {
		return defaultValue
	}
	return v
}

func getEnvFloat64(key string, defaultValue float64) float64 {
	s := os.Getenv(key)
	if s == "" {
		return defaultValue
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return defaultValue
	}
	return v
}

func getEnvString(key string, defaultValue string) string {
	s := os.Getenv(key)
	if s == "" {
		return defaultValue
	}
	return s
}

func getEnvSliceString(key string, defaultValue []string) []string {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	return strings.Fields(val)
}
