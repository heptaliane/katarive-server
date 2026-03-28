package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	pb "github.com/heptaliane/katarive-server/gen/pb/api/v1"
	"github.com/heptaliane/katarive-server/internal/handler"
)

const PORT string = ":9421"

func main() {
	katariveHandler := &handler.KatariveHandlerV1{}

	server := grpc.NewServer()
	pb.RegisterKatariveServiceServer(server, katariveHandler)

	listener, err := net.Listen("tcp", PORT)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	go func() {
		log.Printf("Start gRPC server on %s", PORT)
		if err := server.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Printf("Shut down gRPC server")
	server.GracefulStop()
	log.Printf("Server stopped")
}
