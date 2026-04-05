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

// Ensure MockSource implements Source
var _ katarive.Source = new(MockSource)

type MockNarrator struct {
	Error  error
	Reason *string

	Name    string
	Version string
	Options []*pb.NarratorOption
}

func (n *MockNarrator) Narrate(
	ctx context.Context,
	path string,
	text string,
	options map[string]string,
) (*pb.NarrateResponse, error) {
	if n.Error != nil {
		return nil, n.Error
	}
	return &pb.NarrateResponse{
		Error:  n.Reason != nil,
		Reason: n.Reason,
	}, nil
}
func (n *MockNarrator) GetNarratorServiceMetadata(
	ctx context.Context,
) (*pb.GetNarratorServiceMetadataResponse, error) {
	return &pb.GetNarratorServiceMetadataResponse{
		Name:    n.Name,
		Version: n.Version,
		Options: n.Options,
	}, nil
}

// Ensure MockNarrator implements Narrator
var _ katarive.Narrator = new(MockNarrator)
