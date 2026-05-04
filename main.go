package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/hashicorp/go-hclog"

	"github.com/heptaliane/katarive-server/internal/handler"
)

const ADDR string = ":9421"
const PLUGIN_DIR string = "plugins"
const DATA_DIR string = "data"
const STATIC_DIR string = "web"
const INTERVAL int = 1000
const LOG_LEVEL slog.Level = slog.LevelDebug
const PLUGIN_LOG_LEVEL hclog.Level = hclog.Info

func main() {
	SetupLogger(LOG_LEVEL)

	pm := handler.NewBasePathModifier(
		handler.WithPathRule(DATA_DIR, "file"),
		handler.WithPathRule(STATIC_DIR, "static"),
	)
	grpc, err := NewGRPCServer(PLUGIN_DIR, DATA_DIR, INTERVAL, pm, PLUGIN_LOG_LEVEL)
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
