package main

import (
	"net/http"
	"strings"

	"github.com/hashicorp/go-hclog"
	pb "github.com/heptaliane/katarive-server/gen/pb/api/v1"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"

	"github.com/heptaliane/katarive-server/internal/handler"
	"github.com/heptaliane/katarive-server/internal/service"
)

func newKatariveHandler(
	pluginDir string,
	destDir string,
	interval int,
	logLevel hclog.Level,
) (*handler.KatariveHandlerV1, error) {
	plugin, err := LoadPlugins(pluginDir, logLevel)
	if err != nil {
		return nil, err
	}

	nr, err := NewNarrator(destDir, plugin)
	if err != nil {
		return nil, err
	}

	sr, err := NewSource(destDir, interval, plugin)
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

func NewGRPCServer(
	pluginDir string,
	dataDir string,
	interval int,
	logLevel hclog.Level,
) (*grpc.Server, error) {
	server := grpc.NewServer()

	kh, err := newKatariveHandler(pluginDir, dataDir, interval, logLevel)
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
	grpcwebServer := grpcweb.WrapServer(grpcServer)

	root := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case isGRPC(r):
			grpcServer.ServeHTTP(w, r)
		case grpcwebServer.IsGrpcWebRequest(r):
			grpcwebServer.ServeHTTP(w, r)
		default:
			httpServer.ServeHTTP(w, r)
		}
	})

	c := cors.New(cors.Options{
		AllowedOrigins: []string{
			// Allow access from Vite
			"http://localhost:5173",
		},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{
			"Content-Type",
			"X-User-Agent",
			"X-Grpc-Web",
			"Grpc-Timeout",
		},
		ExposedHeaders: []string{"Grpc-Status", "Grpc-Message"},
		Debug:          false,
	})

	return &http.Server{
		Addr:    addr,
		Handler: h2c.NewHandler(c.Handler(root), &http2.Server{}),
	}
}
