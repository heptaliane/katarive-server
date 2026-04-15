package main

import (
	"github.com/hashicorp/go-plugin"
	katarive "github.com/heptaliane/katarive-go-sdk"
	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
)

type S struct {
	pb.UnimplementedSourceServiceServer
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: katarive.Handshake,
		Plugins: map[string]plugin.Plugin{
			"source": &katarive.SourcePlugin{Impl: &S{}},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
