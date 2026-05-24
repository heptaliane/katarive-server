package handler

import (
	"context"
	"fmt"
	"regexp"

	ppb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
	apb "github.com/heptaliane/katarive-server/gen/pb/api/v1"
	"golang.org/x/sync/singleflight"

	"github.com/heptaliane/katarive-server/internal/service/job"
	"github.com/heptaliane/katarive-server/internal/service/narrator"
	"github.com/heptaliane/katarive-server/internal/service/source"
)

type KatariveHandlerV1 struct {
	apb.UnimplementedKatariveServiceServer

	sr  source.SourceRegistry
	nr  narrator.NarrateRegistry
	nq  job.NarrationJobQueue
	scq job.SourceCollectionJobQueue
	siq job.SourceItemJobQueue

	pm PathModifier
}

func (h *KatariveHandlerV1) QueueNarration(
	ctx context.Context,
	req *apb.QueueNarrationRequest,
) (*apb.QueueNarrationResponse, error) {
	ctx = context.WithoutCancel(ctx)
	jobId, err := h.nq.Queue(
		ctx,
		job.WithNarrationUrl(req.GetUrl()),
		job.WithNarrationNarrator(req.GetNarrator()),
		job.WithNarrationSpeakerId(req.GetSpeakerId()),
		job.WithNarrationEncoding(ppb.AudioEncoding_AUDIO_ENCODING_MP3),
		// TODO: Allow disabling cache
	)
	return &apb.QueueNarrationResponse{Id: jobId}, err
}
func (h *KatariveHandlerV1) GetNarration(
	ctx context.Context,
	req *apb.GetNarrationRequest,
) (*apb.GetNarrationResponse, error) {
	job, err := h.nq.Get(req.GetId())
	if err != nil {
		return nil, err
	}

	result := job.Result()

	var path *string
	var source *apb.SourceSummary
	if result != nil {
		p := h.pm.Do(result.Path)
		path = &p
		source = &apb.SourceSummary{
			Id:    result.Source.GetId(),
			Url:   result.Source.GetUrl(),
			Title: result.Source.GetTitle(),
		}
	}

	return &apb.GetNarrationResponse{
		Status: job.Status(),
		Path:   path,
		Source: source,
	}, job.Error()
}
func (h *KatariveHandlerV1) GetNarrators(
	ctx context.Context,
	req *apb.GetNarratorsRequest,
) (*apb.GetNarratorsResponse, error) {
	narrators := h.nr.Metadata()

	var res apb.GetNarratorsResponse
	for _, n := range narrators {
		var speakers []*apb.Speaker
		for _, s := range n.Speakers {
			speakers = append(speakers, &apb.Speaker{
				Id:    s.GetId(),
				Label: s.GetName(),
			})
		}
		res.Narrator = append(res.Narrator, &apb.Narrator{
			Name:     n.Name,
			Speakers: speakers,
		})
	}

	return &res, nil
}
func (h *KatariveHandlerV1) QueueSourceItem(
	ctx context.Context,
	req *apb.QueueSourceItemRequest,
) (*apb.QueueSourceItemResponse, error) {
	ctx = context.WithoutCancel(ctx)

	jobId, err := h.siq.Queue(
		ctx,
		job.WithSourceItemUrl(req.GetUrl()),
		job.WithoutSourceItemCache(req.GetDisableCache()),
	)
	return &apb.QueueSourceItemResponse{Id: jobId}, err
}
func (h *KatariveHandlerV1) GetSourceItem(
	ctx context.Context,
	req *apb.GetSourceItemRequest,
) (*apb.GetSourceItemResponse, error) {
	job, err := h.siq.Get(req.GetId())
	if err != nil {
		return nil, err
	}

	result := job.Result()
	var metadata *apb.SourceSummary
	var contentPtr *string
	if result != nil {
		metadata = &apb.SourceSummary{
			Id:    result.GetId(),
			Url:   result.GetUrl(),
			Title: result.GetTitle(),
		}
		content := result.GetContent()
		contentPtr = &content
	}

	return &apb.GetSourceItemResponse{
		Status:   job.Status(),
		Metadata: metadata,
		Content:  contentPtr,
	}, job.Error()
}
func (h *KatariveHandlerV1) QueueSourceCollection(
	ctx context.Context,
	req *apb.QueueSourceCollectionRequest,
) (*apb.QueueSourceCollectionResponse, error) {
	ctx = context.WithoutCancel(ctx)
	jobId, err := h.scq.Queue(
		ctx,
		job.WithSourceCollectionUrl(req.GetUrl()),
		job.WithoutSourceCollectionCache(req.GetDisableCache()),
	)
	return &apb.QueueSourceCollectionResponse{Id: jobId}, err
}
func (h *KatariveHandlerV1) GetSourceCollection(
	ctx context.Context,
	req *apb.GetSourceCollectionRequest,
) (*apb.GetSourceCollectionResponse, error) {
	job, err := h.scq.Get(req.GetId())
	if err != nil {
		return nil, err
	}

	result := job.Result()
	var collection *apb.SourceCollection
	var sources []*apb.SourceSummary
	if result != nil {
		collection = &apb.SourceCollection{
			Id:          result.Collection.GetId(),
			Url:         result.Collection.GetUrl(),
			Title:       result.Collection.GetTitle(),
			Description: result.Collection.GetDescription(),
			Author:      result.Collection.GetAuthor(),
			Tags:        result.Collection.GetTags(),
		}
		for _, s := range result.Sources {
			sources = append(sources, &apb.SourceSummary{
				Id:    s.GetId(),
				Title: s.GetTitle(),
				Url:   s.GetUrl(),
			})
		}
	}

	// TODO: Set Sources

	return &apb.GetSourceCollectionResponse{
		Status:     job.Status(),
		Collection: collection,
		Sources:    sources,
	}, job.Error()
}
func (h *KatariveHandlerV1) GetSourceCollections(
	ctx context.Context,
	req *apb.GetSourceCollectionsRequest,
) (*apb.GetSourceCollectionsResponse, error) {
	cs, err := h.sr.SourceCollections()
	if err != nil {
		return nil, err
	}

	var collections []*apb.SourceCollection
	for _, c := range cs {
		collections = append(collections, &apb.SourceCollection{
			Id:          c.GetId(),
			Url:         c.GetUrl(),
			Title:       c.GetTitle(),
			Description: c.GetDescription(),
			Author:      c.GetAuthor(),
			Tags:        c.GetTags(),
		})
	}

	return &apb.GetSourceCollectionsResponse{Collection: collections}, nil
}

// Check KatariveServiceServer implementation
var _ apb.KatariveServiceServer = new(KatariveHandlerV1)

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

	return &KatariveHandlerV1{
		sr:  sr,
		nr:  nr,
		nq:  job.NewNarrationJobQueue(sr, nr, ngrp),
		scq: job.NewSourceCollectionJobQueue(sr, scgrp),
		siq: job.NewSourceItemJobQueue(sr, sigrp),
		pm:  pm,
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
