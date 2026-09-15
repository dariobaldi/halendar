package main

import (
	"fmt"
	"net/http"
)

func (app *application) backgroundTask(r *http.Request, fn func() error) {
	app.wg.Go(func() {
		defer func() {
			pv := recover()
			if pv != nil {
				app.reportServerError(r, fmt.Errorf("%v", pv))
			}
		}()

		err := fn()
		if err != nil {
			app.reportServerError(r, err)
		}
	})
}
