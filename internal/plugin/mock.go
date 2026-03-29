package plugin

import (
	"context"

	katarive "github.com/heptaliane/katarive-go-sdk"
	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
)

type MockSource struct {
	Patterns []string
	Title    string
	Content  string
	NextUrl  string
}

func (s *MockSource) GetSupportedPatterns(
	ctx context.Context,
) (*pb.GetSupportedPatternsResponse, error) {
	return &pb.GetSupportedPatternsResponse{
		Patterns: s.Patterns,
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
