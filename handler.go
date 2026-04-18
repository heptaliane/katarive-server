package main

import (
	"net/http"
	"strings"

	pb "github.com/heptaliane/katarive-server/gen/pb/api/v1"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"

	"github.com/heptaliane/katarive-server/internal/handler"
	"github.com/heptaliane/katarive-server/internal/service"
)

func newKatariveHandler(pluginDir string, destDir string) (*handler.KatariveHandlerV1, error) {
	plugin, err := LoadPlugins(pluginDir)
	if err != nil {
		return nil, err
	}

	nr, err := NewNarrator(destDir, plugin)
	if err != nil {
		return nil, err
	}

	sr, err := NewSource(destDir, plugin)
	if err != nil {
		return nil, err
	}

	js := service.NewNarrateJobManager(nr, sr)
	return handler.NewKatariveHandler(js), nil
}
func isGRPC(r *http.Request) bool {
	contentType := r.Header.Get("Content-Type")
	return strings.HasPrefix(contentType, "application/grpc") && r.ProtoMajor == 2
}

func NewGRPCServer(pluginDir string, dataDir string) (*grpc.Server, error) {
	server := grpc.NewServer()

	kh, err := newKatariveHandler(pluginDir, dataDir)
	if err != nil {
		return nil, err
	}

	pb.RegisterKatariveServiceServer(server, kh)

	return server, nil
}
func NewHttpServer(paths map[string]string) *http.ServeMux {
	mux := http.NewServeMux()

	for prefix, path := range paths {
		fs := http.FileServer(http.Dir(path))
		mux.Handle(prefix, http.StripPrefix(prefix, fs))
	}

	return mux
}
func NewHttp2Server(
	addr string,
	grpcServer *grpc.Server,
	httpServer *http.ServeMux,
) *http.Server {
	root := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isGRPC(r) {
			grpcServer.ServeHTTP(w, r)
		} else {
			httpServer.ServeHTTP(w, r)
		}
	})

	return &http.Server{
		Addr:    addr,
		Handler: h2c.NewHandler(root, &http2.Server{}),
	}
}
