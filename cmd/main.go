package main

import (
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"parce/internal/config"
	"parce/internal/httpapi"
)

func main() {
	if err := config.LoadEnv(".env"); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatal("load .env:", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           httpapi.NewMux(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	slog.Info("starting server", "port", port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("server:", err)
	}
}
