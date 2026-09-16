package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *app) serve() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		app.logger.Info("Shutting down server", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		app.logger.Info("Closing WebSocket connections")
		for _, ws := range app.websockets {
			for client := range ws.Clients {
				ws.Unregister <- client
				client.Conn.Close()
			}
		}

		// TODO : test graceful shutdown of background tasks
		// app.logger.Info("completing background tasks", "addr", srv.Addr)
		// isDone := waitTimeout(&app.wg, 2*time.Second)
		// if !isDone {
		// 	app.logger.Info("Background tasks timed out")
		// }
		shutdownError <- nil
	}()

	app.logger.Info("starting server", "address", srv.Addr, "env", app.config.env, "version", version)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	app.logger.Info("stopped server", "address", srv.Addr)
	return nil
}
