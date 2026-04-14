package handler

import (
	"context"
	"errors"

	pb "github.com/heptaliane/katarive-server/gen/pb/api/v1"
	"github.com/heptaliane/katarive-server/internal/service"
)

type KatariveHandlerV1 struct {
	pb.UnimplementedKatariveServiceServer

	js service.NarrateJobService
}

func NewKatariveHandler(js service.NarrateJobService) *KatariveHandlerV1 {
	return &KatariveHandlerV1{
		js: js,
	}
}

func (h *KatariveHandlerV1) CreateNarration(
	ctx context.Context,
	req *pb.CreateNarrationRequest,
) (*pb.CreateNarrationResponse, error) {
	jobId, err := h.js.Enqueue(ctx, req.GetUrl())
	if err != nil {
		return nil, err
	}

	return &pb.CreateNarrationResponse{
		Id: jobId,
	}, nil
}

func (h *KatariveHandlerV1) GetJobStatus(
	ctx context.Context,
	req *pb.GetJobStatusRequest,
) (*pb.GetJobStatusResponse, error) {
	job, err := h.js.GetJob(req.GetId())
	if err != nil {
		var notFound *service.JobNotFoundError
		if errors.As(err, &notFound) {
			return &pb.GetJobStatusResponse{
				Status: pb.GetJobStatusResponse_STATUS_NOT_FOUND,
			}, nil
		}
		return &pb.GetJobStatusResponse{
			Status: pb.GetJobStatusResponse_STATUS_FAILED,
		}, nil
	}

	result, err := job.GetResult()
	if err != nil {
		return &pb.GetJobStatusResponse{
			Status: pb.GetJobStatusResponse_STATUS_FAILED,
		}, nil
	}
	if result == "" {
		return &pb.GetJobStatusResponse{
			Status: pb.GetJobStatusResponse_STATUS_PROGRESSING,
		}, nil
	}
	return &pb.GetJobStatusResponse{
		Status: pb.GetJobStatusResponse_STATUS_COMPLETED,
		Path:   &result,
	}, nil
}

// Check KatariveServiceServer implementation
var _ pb.KatariveServiceServer = new(KatariveHandlerV1)
