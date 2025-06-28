package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/romshark/ebiten-experiment/httpserve"
)

func main() {
	fHost := flag.String("host", "localhost:8080", "host address")
	flag.Parse()

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	srv := &http.Server{
		Addr:    *fHost,
		Handler: httpserve.NewServer(log),
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	go func() {
		log.Info("listening", slog.String("host", *fHost))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("serving http", slog.Any("err", err))
		}
	}()

	<-ctx.Done()
	log.Info("shutting down server")

	if err := srv.Shutdown(context.Background()); err != nil {
		log.Error("error shutting down server", slog.Any("err", err))
	}
	log.Info("stopped listening")
}
