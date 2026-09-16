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

	// Mail
	router.HandlerFunc(http.MethodGet, "/v1/mail/health", app.requirePermission(UserLevel, app.mailHealthHandler))
	router.HandlerFunc(http.MethodGet, "/v1/mail/folders", app.requirePermission(UserLevel, app.listFoldersHandler))
	router.HandlerFunc(http.MethodGet, "/v1/mail/messages", app.requirePermission(UserLevel, app.listMessagesHandler))
	router.HandlerFunc(http.MethodGet, "/v1/mail/messages/:uid", app.requirePermission(UserLevel, app.readMessageHandler))
	router.HandlerFunc(http.MethodPost, "/v1/mail/messages/:uid/reply", app.requirePermission(UserLevel, app.replyMessageHandler))
	router.HandlerFunc(http.MethodPost, "/v1/mail/send", app.requirePermission(UserLevel, app.sendMessageHandler))
	router.HandlerFunc(http.MethodPost, "/v1/mail/draft", app.requirePermission(UserLevel, app.draftMessageHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/mail/read", app.requirePermission(UserLevel, app.markMessagesReadHandler))
	router.HandlerFunc(http.MethodPost, "/v1/mail/move", app.requirePermission(UserLevel, app.moveMessagesHandler))

	// Calendar
	router.HandlerFunc(http.MethodGet, "/v1/calendar/health", app.requirePermission(UserLevel, app.calendarHealthHandler))
	router.HandlerFunc(http.MethodGet, "/v1/calendar/calendars", app.requirePermission(UserLevel, app.listCalendarsHandler))
	router.HandlerFunc(http.MethodGet, "/v1/calendar/events", app.requirePermission(UserLevel, app.listEventsHandler))
	router.HandlerFunc(http.MethodPost, "/v1/calendar/events", app.requirePermission(UserLevel, app.addEventsHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/calendar/events/:uid", app.requirePermission(UserLevel, app.deleteEventHandler))
	router.HandlerFunc(http.MethodGet, "/v1/calendar/busy", app.requirePermission(UserLevel, app.checkBusyHandler))

	// Schedule: propose a slot from a mail, book it only once the caller confirms
	router.HandlerFunc(http.MethodGet, "/v1/schedule/propose/:uid", app.requirePermission(UserLevel, app.proposeScheduleHandler))
	router.HandlerFunc(http.MethodPost, "/v1/schedule/confirm", app.requirePermission(UserLevel, app.confirmScheduleHandler))

	// AI: test connectivity to the local Ollama model
	router.HandlerFunc(http.MethodPost, "/v1/ai/prompt", app.requirePermission(UserLevel, app.promptHandler))
	router.HandlerFunc(http.MethodPost, "/v1/ai/test", app.requirePermission(UserLevel, app.testPromptHandler))

	return app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(app.requestsSlog(router)))))
}
