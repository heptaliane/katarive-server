package handler

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	pb "github.com/heptaliane/katarive-server/gen/pb/api/v1"
	"github.com/heptaliane/katarive-server/internal/service"
)

type KatariveHandlerV1 struct {
	pb.UnimplementedKatariveServiceServer

	js service.NarrateJobService
	pm PathModifier
}

func NewKatariveHandler(
	js service.NarrateJobService,
	pm PathModifier,
) *KatariveHandlerV1 {
	return &KatariveHandlerV1{
		js: js,
		pm: pm,
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
	result = h.pm.Do(result)
	return &pb.GetJobStatusResponse{
		Status: pb.GetJobStatusResponse_STATUS_COMPLETED,
		Path:   &result,
	}, nil
}

// Check KatariveServiceServer implementation
var _ pb.KatariveServiceServer = new(KatariveHandlerV1)

// -----------------
// helper components
// -----------------
type PathModifier interface {
	Do(path string) string
}
type BasePathModifier struct {
	rules []basePathModificationRule
}

func (m *BasePathModifier) Do(path string) string {
	p := []byte(path)
	for _, rule := range m.rules {
		p = rule.source.ReplaceAll(p, rule.dest)
	}
	return string(p)
}

// Ensure PathModifier implementation
var _ PathModifier = new(BasePathModifier)

type basePathModificationRule struct {
	source *regexp.Regexp
	dest   []byte
}

type BasePathModifierOption = func(m *BasePathModifier)

func WithPathRule(sourcePrefix string, destPrefix string) BasePathModifierOption {
	return func(m *BasePathModifier) {
		m.rules = append(m.rules, basePathModificationRule{
			source: regexp.MustCompile(fmt.Sprintf("^%s", sourcePrefix)),
			dest:   []byte(destPrefix),
		})
	}
}
func NewBasePathModifier(opts ...BasePathModifierOption) *BasePathModifier {
	m := new(BasePathModifier)
	for _, opt := range opts {
		opt(m)
	}
	return m
}
