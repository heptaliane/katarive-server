package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"google.golang.org/grpc"

	pb "github.com/heptaliane/katarive-server/gen/pb/api/v1"
	"github.com/heptaliane/katarive-server/internal/handler"
	"github.com/heptaliane/katarive-server/internal/service"
)

const PORT string = ":9421"
const PLUGIN_DIR string = "plugins"
const DATA_DIR string = "data"

func loadPlugins(pluginDir string) (*service.PluginRegistry, error) {
	files, err := os.ReadDir(pluginDir)
	if err != nil {
		return nil, err
	}

	pr := new(service.PluginRegistry)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		path := filepath.Join(pluginDir, file.Name())
		err := pr.Load(path)
		if err != nil {
			return nil, err
		}
	}
	return pr, nil
}

func newNarrator(destDir string, plugins *service.PluginRegistry) (service.NarratorRegistry, error) {
	ctx := context.Background()

	var narrators []service.NarratorManager
	for _, rawNarrator := range plugins.GetNarrators() {
		narrator, err := service.NewSemaphoreNarratorManager(ctx, rawNarrator)
		if err != nil {
			return nil, err
		}
		narrators = append(narrators, narrator)
	}

	return service.NewFileNarratorRegistry(destDir, narrators), nil
}

func newSource(destDir string, plugins *service.PluginRegistry) (service.SourceRegistry, error) {
	ctx := context.Background()

	var sources []service.SourceManager
	for _, rawSource := range plugins.GetSources() {
		source, err := service.NewSemaphoreSourceManager(ctx, rawSource)
		if err != nil {
			return nil, err
		}
		sources = append(sources, source)
	}

	return service.NewFileSourceRegistry(destDir, sources), nil
}

func newKatariveHandler(pluginDir string, destDir string) (*handler.KatariveHandlerV1, error) {
	plugin, err := loadPlugins(pluginDir)
	if err != nil {
		return nil, err
	}

	nr, err := newNarrator(destDir, plugin)
	if err != nil {
		return nil, err
	}

	sr, err := newSource(destDir, plugin)
	if err != nil {
		return nil, err
	}

	js := service.NewNarrateJobManager(nr, sr)
	return handler.NewKatariveHandler(js), nil
}

func main() {
	katariveHandler, err := newKatariveHandler(PLUGIN_DIR, DATA_DIR)
	if err != nil {
		log.Fatalf("Failed to initialize handler: %v", err)
	}

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
