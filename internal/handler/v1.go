package handler

import (
	"context"

	pb "github.com/heptaliane/katarive-server/gen/pb/api/v1"
)

type KatariveHandlerV1 struct {
	pb.UnimplementedKatariveServiceServer
}

func (h *KatariveHandlerV1) CreateNarration(
	ctx context.Context,
	req *pb.CreateNarrationRequest,
) (*pb.CreateNarrationResponse, error) {
	// TODO: Implementation
	return &pb.CreateNarrationResponse{}, nil
}

func (h *KatariveHandlerV1) GetTask(
	ctx context.Context,
	req *pb.GetJobStatusRequest,
) (*pb.GetJobStatusResponse, error) {
	// TODO: Implementation
	return &pb.GetJobStatusResponse{}, nil
}

// Check KatariveServiceServer implementation
var _ pb.KatariveServiceServer = new(KatariveHandlerV1)
