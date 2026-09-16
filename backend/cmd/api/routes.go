package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

const (
	UserLevel  = 1
	MidLevel   = 5
	AdminLevel = 10
	DevLevel   = 100
)

func (app *app) routes() http.Handler {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// Admin
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodGet, "/v1/requests_stats", app.requirePermission(AdminLevel, app.requestsStatsHandler))
	
	// Users
	router.HandlerFunc(http.MethodGet, "/v1/users", app.requirePermission(AdminLevel, app.getUsersHandler))
	router.HandlerFunc(http.MethodPost, "/v1/users", app.requirePermission(AdminLevel, app.registerUserHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/user", app.requirePermission(AdminLevel, app.updateUserHandler))
	router.HandlerFunc(http.MethodPut, "/v1/users/activate", app.requirePermission(AdminLevel, app.activateUserHandler))
	router.HandlerFunc(http.MethodPost, "/v1/users/authentication", app.createAuthenticationToken)
	router.HandlerFunc(http.MethodPost, "/v1/users/change_password", app.requirePermission(UserLevel, app.changePasswordHandler))

	// Websocket
	router.HandlerFunc(http.MethodPost, "/v1/websocket/token", app.requirePermission(UserLevel, app.createWsToken))
	router.HandlerFunc(http.MethodGet, "/v1/ws/:channel/:token", app.WebSocketHandler)

	return app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(app.requestsSlog(router)))))
}
