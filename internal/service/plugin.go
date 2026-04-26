package service

import (
	"context"
	"os"
	"os/exec"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	katarive "github.com/heptaliane/katarive-go-sdk"
	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
	"google.golang.org/grpc/reflection/grpc_reflection_v1"
)

type PluginRegistry struct {
	sources   []pb.SourceServiceClient
	narrators []pb.NarratorServiceClient
	clients   []*plugin.Client
	logger    hclog.Logger
}

func (r *PluginRegistry) Load(path string) error {
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  katarive.Handshake,
		Plugins:          katarive.PluginMap,
		Cmd:              exec.Command(path),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Logger:           r.logger,
		Managed:          true,
	})
	r.clients = append(r.clients, client)

	rpcClient, err := client.Client()
	if err != nil {
		return err
	}
	grpcClient := rpcClient.(*plugin.GRPCClient)
	names, err := grpcServices(grpcClient)
	if err != nil {
		return err
	}

	for _, name := range names {
		if name == pb.SourceService_ServiceDesc.ServiceName {
			rawSource, err := rpcClient.Dispense("source")
			if err != nil {
				return err
			}
			source := rawSource.(pb.SourceServiceClient)
			r.sources = append(r.sources, source)
		}
		if name == pb.NarratorService_ServiceDesc.ServiceName {
			rawNarrator, err := rpcClient.Dispense("narrator")
			if err != nil {
				return err
			}
			narrator := rawNarrator.(pb.NarratorServiceClient)
			r.narrators = append(r.narrators, narrator)
		}
	}

	return nil
}
func (r *PluginRegistry) GetSources() []pb.SourceServiceClient {
	return r.sources
}
func (r *PluginRegistry) GetNarrators() []pb.NarratorServiceClient {
	return r.narrators
}

// Helper functions
func grpcServices(client *plugin.GRPCClient) ([]string, error) {
	conn := client.Conn

	reflectionClient := grpc_reflection_v1.NewServerReflectionClient(conn)
	ctx := context.Background()
	stream, err := reflectionClient.ServerReflectionInfo(ctx)
	if err != nil {
		return []string{}, err
	}

	stream.Send(&grpc_reflection_v1.ServerReflectionRequest{
		MessageRequest: &grpc_reflection_v1.ServerReflectionRequest_ListServices{},
	})
	resp, err := stream.Recv()
	if err != nil {
		return []string{}, err
	}

	var names []string
	for _, svc := range resp.GetListServicesResponse().Service {
		names = append(names, svc.GetName())
	}
	return names, nil
}

func NewPluginRegistry(level hclog.Level) *PluginRegistry {
	return &PluginRegistry{
		logger: hclog.New(&hclog.LoggerOptions{
			Level:      level,
			JSONFormat: true,
			Output:     os.Stdout,
		}),
	}
}
