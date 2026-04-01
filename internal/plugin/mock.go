package plugin

import (
	"context"

	katarive "github.com/heptaliane/katarive-go-sdk"
	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
)

type MockSource struct {
	Name             string
	Version          string
	SupportedPattern string

	Title   string
	Content string
	NextUrl string
}

func (s *MockSource) GetSourceServiceMetadata(
	ctx context.Context,
) (*pb.GetSourceServiceMetadataResponse, error) {
	return &pb.GetSourceServiceMetadataResponse{
		Name:             s.Name,
		Version:          s.Version,
		SupportedPattern: s.SupportedPattern,
	}, nil
}

func (s *MockSource) GetSource(ctx context.Context, url string) (*pb.GetSourceResponse, error) {
	return &pb.GetSourceResponse{
		Title:   s.Title,
		Content: s.Content,
		NextUrl: s.NextUrl,
	}, nil
}

// Make sure to impplement Source
var _ katarive.Source = new(MockSource)
