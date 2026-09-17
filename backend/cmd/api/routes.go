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

	// Email accounts: connect messaging accounts (Gmail today, more providers later),
	// imported in the background once connected. The callback has no Authorization
	// header -- the provider's redirect calls it directly.
	router.HandlerFunc(http.MethodGet, "/v1/email-accounts", app.requirePermission(UserLevel, app.listEmailAccountsHandler))
	router.HandlerFunc(http.MethodGet, "/v1/email-accounts/:provider/connect", app.requirePermission(UserLevel, app.connectEmailAccountHandler))
	router.HandlerFunc(http.MethodGet, "/v1/email-accounts/:provider/callback", app.emailAccountCallbackHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/email-accounts/:id", app.requirePermission(UserLevel, app.deleteEmailAccountHandler))

	// Email messages: re-run AI analysis on everything already imported, e.g. after
	// improving the extraction prompt.
	router.HandlerFunc(http.MethodPost, "/v1/email-messages/reanalyze", app.requirePermission(UserLevel, app.reanalyzeEmailMessagesHandler))

	// Calendar accounts: connect a calendar to check availability against (Google
	// Calendar and CalDAV -- Apple iCloud, and any groupware that speaks it -- today,
	// more providers later). Connecting Gmail links a Google Calendar account too
	// (see linkGoogleCalendarFromEmail), so this is mainly for a calendar on its own.
	router.HandlerFunc(http.MethodGet, "/v1/calendar-accounts", app.requirePermission(UserLevel, app.listCalendarAccountsHandler))
	router.HandlerFunc(http.MethodGet, "/v1/calendar-accounts/:provider/connect", app.requirePermission(UserLevel, app.connectCalendarAccountHandler))
	router.HandlerFunc(http.MethodGet, "/v1/calendar-accounts/:provider/callback", app.calendarAccountCallbackHandler)
	router.HandlerFunc(http.MethodPost, "/v1/calendar-accounts/caldav", app.requirePermission(UserLevel, app.connectCaldavCalendarHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/calendar-accounts/:id", app.requirePermission(UserLevel, app.deleteCalendarAccountHandler))

	// Proposals: the extracted-event review screen (email + slots + calendar
	// availability + drafted reply). Sending/booking isn't wired up yet -- confirm
	// and reject just move a proposal to "history" for now.
	router.HandlerFunc(http.MethodGet, "/v1/proposals", app.requirePermission(UserLevel, app.listProposalsHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/proposals/:id/slot", app.requirePermission(UserLevel, app.selectProposalSlotHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/proposals/:id/draft", app.requirePermission(UserLevel, app.updateProposalDraftHandler))
	router.HandlerFunc(http.MethodPost, "/v1/proposals/:id/confirm", app.requirePermission(UserLevel, app.confirmProposalHandler))
	router.HandlerFunc(http.MethodPost, "/v1/proposals/:id/reject", app.requirePermission(UserLevel, app.rejectProposalHandler))
	router.HandlerFunc(http.MethodPost, "/v1/proposals/:id/reanalyze", app.requirePermission(UserLevel, app.reanalyzeProposalHandler))

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

	// AI settings: which model email analysis uses for this user -- the shared local
	// Ollama instance by default, or the user's own Claude/Gemini API key once
	// they've connected and activated one.
	// setAIProviderHandler is PATCH, not PUT, so its static "/provider" segment
	// doesn't collide with the wildcard ":provider" below -- httprouter keeps a
	// separate route tree per method, but panics at startup if the same method has
	// both a static and a wildcard child at the same position.
	router.HandlerFunc(http.MethodGet, "/v1/ai-settings", app.requirePermission(UserLevel, app.getAISettingsHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/ai-settings/provider", app.requirePermission(UserLevel, app.setAIProviderHandler))
	router.HandlerFunc(http.MethodPut, "/v1/ai-settings/:provider/key", app.requirePermission(UserLevel, app.connectAIKeyHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/ai-settings/:provider/key", app.requirePermission(UserLevel, app.disconnectAIKeyHandler))

	// Devices: register a push token so notifications can be sent to it
	router.HandlerFunc(http.MethodGet, "/v1/devices", app.requirePermission(UserLevel, app.listDevicesHandler))
	router.HandlerFunc(http.MethodPost, "/v1/devices", app.requirePermission(UserLevel, app.registerDeviceHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/devices", app.requirePermission(UserLevel, app.unregisterDeviceHandler))

	// Push: test connectivity to Firebase / send a test notification
	router.HandlerFunc(http.MethodGet, "/v1/push/health", app.requirePermission(UserLevel, app.pushHealthHandler))
	router.HandlerFunc(http.MethodPost, "/v1/push/test", app.requirePermission(UserLevel, app.testPushHandler))

	return app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(app.requestsSlog(router)))))
}
