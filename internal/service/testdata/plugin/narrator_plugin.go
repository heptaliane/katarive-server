package main

import (
	"github.com/hashicorp/go-plugin"
	katarive "github.com/heptaliane/katarive-go-sdk"
	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
)

type N struct {
	pb.UnimplementedNarratorServiceServer
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: katarive.Handshake,
		Plugins: map[string]plugin.Plugin{
			"narrator": &katarive.NarratorPlugin{Impl: &N{}},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
