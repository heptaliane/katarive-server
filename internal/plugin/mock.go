package plugin

import (
	"context"

	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
)

type MockSource struct {
	pb.UnimplementedSourceServiceServer

	GetSourceServiceMetadataResponse *pb.GetSourceServiceMetadataResponse
	GetSourceResponse                *pb.GetSourceResponse
}

func (s *MockSource) GetSourceServiceMetadata(
	ctx context.Context,
	req *pb.GetSourceServiceMetadataRequest,
) (*pb.GetSourceServiceMetadataResponse, error) {
	return s.GetSourceServiceMetadataResponse, nil
}

func (s *MockSource) GetSource(
	ctx context.Context,
	req *pb.GetSourceRequest,
) (*pb.GetSourceResponse, error) {
	return s.GetSourceResponse, nil
}

// Ensure MockSource implements Source
var _ pb.SourceServiceServer = new(MockSource)

type MockNarrator struct {
	pb.UnimplementedNarratorServiceServer

	NarrateError                       error
	NarrateResponse                    *pb.NarrateResponse
	GetNarratorServiceMetadataResponse *pb.GetNarratorServiceMetadataResponse
}

func (n *MockNarrator) Narrate(
	ctx context.Context,
	req *pb.NarrateRequest,
) (*pb.NarrateResponse, error) {
	if n.NarrateError != nil {
		return nil, n.NarrateError
	}
	return n.NarrateResponse, nil
}
func (n *MockNarrator) GetNarratorServiceMetadata(
	ctx context.Context,
	req *pb.GetNarratorServiceMetadataRequest,
) (*pb.GetNarratorServiceMetadataResponse, error) {
	return n.GetNarratorServiceMetadataResponse, nil
}

// Ensure MockNarrator implements Narrator
var _ pb.NarratorServiceServer = new(MockNarrator)
