package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const ADDR string = ":9421"
const HTTP_PORT string = ":9422"
const PLUGIN_DIR string = "plugins"
const DATA_DIR string = "data"
const STATIC_DIR string = "web"
const INTERVAL int = 1000
const LOG_LEVEL slog.Level = slog.LevelDebug

func main() {
	SetupLogger(LOG_LEVEL)

	grpc, err := NewGRPCServer(PLUGIN_DIR, DATA_DIR, INTERVAL)
	if err != nil {
		slog.Error("Failed to initialize grpc server", "error", err)
		os.Exit(1)
	}

	mux := NewHttpServer(map[string]string{
		"/file/":   DATA_DIR,
		"/static/": STATIC_DIR,
	})

	server := NewHttp2Server(ADDR, grpc, mux)

	go func() {
		slog.Info(fmt.Sprintf("Start server on %s", ADDR))
		if err := server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				slog.Error("Failed to serve", "error", err)
				os.Exit(1)
			}
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	slog.Info("Shut down server")
	server.Close()
	slog.Info("Server stopped")
}
