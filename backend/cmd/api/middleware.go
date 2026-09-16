package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/dariobaldi/halendar_back/internal/data"
	"github.com/dariobaldi/halendar_back/internal/validator"
	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
)

type RunTime struct {
	Hour   int
	Minute int
}

func (app *app) rateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.limiter.enabled {
			ip := realip.FromRequest(r)

			app.mu.Lock()
			if _, found := app.clientsIPs[ip]; !found {
				app.clientsIPs[ip] = &client{limiter: rate.NewLimiter(rate.Limit(app.config.limiter.requestPerSecond), app.config.limiter.burstLimit)}
			}

			app.clientsIPs[ip].lastSeen = time.Now()
			app.mu.Unlock()

			if !app.clientsIPs[ip].limiter.Allow() {
				app.rateLimitExceededResponse(w, r)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (app *app) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *app) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		token := headerParts[1]

		v := validator.New()
		data.ValidateTokenPlaintext(v, token)
		if !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		user, err := app.models.Users.GetForToken(token, data.ScopeAuthentication, false)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		r = app.contextSetUser(r, user)

		next.ServeHTTP(w, r)
	})
}

func (app *app) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}

		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *app) requirePermission(accessLevel int, next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		if user.AccessLevel < accessLevel {
			app.notPermittedResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	}

	return app.requireActivatedUser(fn)
}

func (app *app) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		w.Header().Add("Vary", "Access-Control-Request-Method")

		origin := r.Header.Get("Origin")
		isPreflight := r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != ""

		for i := range app.config.cors.trustedOrigins {
			if origin == app.config.cors.trustedOrigins[i] {
				w.Header().Set("Access-Control-Allow-Origin", origin)

				if isPreflight {
					w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
					w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
					w.WriteHeader(http.StatusOK)
					return
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (app *app) requestsSlog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)

		if duration > SlowThreshold {
			app.fileLogger.Warn("Slow request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Duration("duration", duration),
				slog.Time("timestamp", time.Now()),
			)
		}
	})
}

func calculateNextRun(now time.Time, runTimes []RunTime) time.Time {
	// Find the next occurrence of these times
	var nextRun time.Time
	for _, rt := range runTimes {
		candidate := time.Date(now.Year(), now.Month(), now.Day(), rt.Hour, rt.Minute, 0, 0, now.Location())
		if candidate.After(now) {
			if nextRun.IsZero() || candidate.Before(nextRun) {
				nextRun = candidate
			}
		}
	}

	// If no run time is found today, pick the earliest for the next day
	if nextRun.IsZero() {
		nextDay := now.Add(24 * time.Hour)
		nextRun = time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), runTimes[0].Hour, runTimes[0].Minute, 0, 0, now.Location())
	}

	return nextRun
}
