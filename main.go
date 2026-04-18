package main

import (
	"log"
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

func main() {
	grpc, err := NewGRPCServer(PLUGIN_DIR, DATA_DIR)
	if err != nil {
		log.Fatalf("Failed to initialize grpc server: %v", err)
	}

	mux := NewHttpServer(map[string]string{
		"/file/":   DATA_DIR,
		"/static/": STATIC_DIR,
	})

	server := NewHttp2Server(ADDR, grpc, mux)

	go func() {
		log.Printf("Start server on %s", ADDR)
		if err := server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Fatalf("Failed to serve: %v", err)
			}
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Printf("Shut down gRPC server")
	server.Close()
	log.Printf("Server stopped")
}
