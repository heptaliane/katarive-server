package handler

import (
	"context"
	"fmt"
	"regexp"

	pb "github.com/heptaliane/katarive-server/gen/pb/api/v1"
	"golang.org/x/sync/singleflight"

	"github.com/heptaliane/katarive-server/internal/service/job"
	"github.com/heptaliane/katarive-server/internal/service/narrator"
	"github.com/heptaliane/katarive-server/internal/service/source"
)

type KatariveHandlerV1 struct {
	pb.UnimplementedKatariveServiceServer

	sr   source.SourceRegistry
	nr   narrator.NarrateRegistry
	nq   job.NarrationJobQueue
	scq  job.SourceCollectionJobQueue
	siq  job.SourceItemJobQueue
	sisq job.SourceItemsJobQueue

	pm PathModifier
}

func (h *KatariveHandlerV1) QueueNarration(
	ctx context.Context,
	req *pb.QueueNarrationRequest,
) (*pb.QueueNarrationResponse, error) {
	ctx = context.WithoutCancel(ctx)
	jobId, err := h.nq.Queue(
		ctx,
		job.WithNarrationUrl(req.GetUrl()),
		job.WithNarrationNarrator(req.GetNarrator()),
		job.WithNarrationSpeakerId(req.GetSpeakerId()),
	)
	return &pb.QueueNarrationResponse{Id: jobId}, err
}
func (h *KatariveHandlerV1) GetNarration(
	ctx context.Context,
	req *pb.GetNarrationRequest,
) (*pb.GetNarrationResponse, error) {
	job, err := h.nq.Get(req.GetId())
	if err != nil {
		return nil, err
	}

	result := job.Result()
	var path *string
	if result != nil {
		modified := h.pm.Do(*path)
		path = &modified
	}

	return &pb.GetNarrationResponse{
		Status: job.Status(),
		Path:   result,
	}, job.Error()
}
func (h *KatariveHandlerV1) GetNarrators(
	ctx context.Context,
	req *pb.GetNarratorsRequest,
) (*pb.GetNarratorsResponse, error) {
	narrators := h.nr.Metadata()

	var res pb.GetNarratorsResponse
	for _, n := range narrators {
		var speakers []*pb.Speaker
		for _, s := range n.Speakers {
			speakers = append(speakers, &pb.Speaker{
				Id:    s.GetId(),
				Label: s.GetName(),
			})
		}
		res.Narrator = append(res.Narrator, &pb.Narrator{
			Name:     n.Name,
			Speakers: speakers,
		})
	}

	return &res, nil
}
func (h *KatariveHandlerV1) QueueSourceItem(
	ctx context.Context,
	req *pb.QueueSourceItemRequest,
) (*pb.QueueSourceItemResponse, error) {
	ctx = context.WithoutCancel(ctx)
	jobId, err := h.siq.Queue(ctx, job.WithSourceItemUrl(req.GetUrl()))
	return &pb.QueueSourceItemResponse{Id: jobId}, err
}
func (h *KatariveHandlerV1) GetSourceItem(
	ctx context.Context,
	req *pb.GetSourceItemRequest,
) (*pb.GetSourceItemResponse, error) {
	job, err := h.siq.Get(req.GetId())
	if err != nil {
		return nil, err
	}

	result := job.Result()
	var metadata *pb.SourceSummary
	var contentPtr *string
	if result != nil {
		metadata = &pb.SourceSummary{
			Id:    result.GetId(),
			Url:   result.GetUrl(),
			Title: result.GetTitle(),
		}
		content := result.GetContent()
		contentPtr = &content
	}

	return &pb.GetSourceItemResponse{
		Status:   job.Status(),
		Metadata: metadata,
		Content:  contentPtr,
	}, job.Error()
}
func (h *KatariveHandlerV1) QueueSourceCollection(
	ctx context.Context,
	req *pb.QueueSourceCollectionRequest,
) (*pb.QueueSourceCollectionResponse, error) {
	ctx = context.WithoutCancel(ctx)
	jobId, err := h.scq.Queue(
		ctx,
		job.WithSourceCollectionUrl(req.GetUrl()),
	)
	return &pb.QueueSourceCollectionResponse{Id: jobId}, err
}
func (h *KatariveHandlerV1) GetSourceCollection(
	ctx context.Context,
	req *pb.GetSourceCollectionRequest,
) (*pb.GetSourceCollectionResponse, error) {
	job, err := h.scq.Get(req.GetId())
	if err != nil {
		return nil, err
	}

	result := job.Result()
	var collection *pb.SourceCollection
	var sources []*pb.SourceSummary
	if result != nil {
		collection = &pb.SourceCollection{
			Id:          result.GetId(),
			Url:         result.GetUrl(),
			Title:       result.GetTitle(),
			Description: result.GetDescription(),
			Author:      result.GetAuthor(),
			Tags:        result.GetTags(),
		}
	}

	// TODO: Set Sources

	return &pb.GetSourceCollectionResponse{
		Status:     job.Status(),
		Collection: collection,
		Sources:    sources,
	}, job.Error()
}

// Check KatariveServiceServer implementation
var _ pb.KatariveServiceServer = new(KatariveHandlerV1)

// -----------------
// helper components
// -----------------
func NewKatariveHandlerV1(
	sr source.SourceRegistry,
	nr narrator.NarrateRegistry,
	pm PathModifier,
) *KatariveHandlerV1 {
	ngrp := new(singleflight.Group)
	scgrp := new(singleflight.Group)
	sigrp := new(singleflight.Group)
	sisgrp := scgrp

	return &KatariveHandlerV1{
		sr:   sr,
		nr:   nr,
		nq:   job.NewNarrationJobQueue(sr, nr, ngrp),
		scq:  job.NewSourceCollectionJobQueue(sr, scgrp),
		siq:  job.NewSourceItemJobQueue(sr, sigrp),
		sisq: job.NewSourceItemsJobQueue(sr, sisgrp),
		pm:   pm,
	}
}

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
