package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	mux := httprouter.New()

	mux.NotFound = http.HandlerFunc(app.notFound)
	mux.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowed)

	mux.HandlerFunc("GET", "/status", app.status)

	mux.Handler("GET", "/restricted-basic-auth", app.requireBasicAuthentication(http.HandlerFunc(app.restricted)))

	return app.logRequest(app.recoverPanic(mux))
}
