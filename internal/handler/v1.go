package handler

import (
	"context"
	"fmt"
	"regexp"

	ppb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
	apb "github.com/heptaliane/katarive-server/gen/pb/api/v1"

	"github.com/heptaliane/katarive-server/internal/service/job"
	"github.com/heptaliane/katarive-server/internal/service/narrator"
	"github.com/heptaliane/katarive-server/internal/service/source"
)

type KatariveHandlerV1 struct {
	apb.UnimplementedKatariveServiceServer

	sr  source.SourceRegistry
	nr  narrator.NarrateRegistry
	nq  job.NarrateJobQueue
	scq job.SourceCollectionJobQueue
	siq job.SourceItemJobQueue

	pm PathModifier
}

func (h *KatariveHandlerV1) GetNarration(
	ctx context.Context,
	req *apb.GetNarrationRequest,
) (*apb.GetNarrationResponse, error) {
	item, err := h.sr.GetItem(req.GetUrl())
	if err != nil {
		return nil, err
	}

	opts := []narrator.NarrateOption{
		narrator.WithSpeaker(req.GetSpeakerId()),
		narrator.WithNarrator(req.GetNarrator()),
		narrator.WithEncoding(ppb.AudioEncoding_AUDIO_ENCODING_MP3),
	}

	var path *string
	if item != nil {
		result := h.nr.Get(item, opts...)
		if result != nil {
			path = h.pm.Do(result.Path)
		}
	}

	status := apb.JobStatus_JOB_STATUS_COMPLETED
	if path == nil {
		ctx = context.WithoutCancel(ctx)
		job, err := h.nq.Queue(
			ctx,
			job.WithNarrationUrl(req.GetUrl()),
			job.WithNarrationSpeakerId(req.GetSpeakerId()),
			job.WithNarrationNarrator(req.GetNarrator()),
			job.WithNarrationEncoding(ppb.AudioEncoding_AUDIO_ENCODING_MP3),
		)
		if err != nil {
			return nil, err
		}
		status = job.Status()
		err = job.Error()
	}

	return &apb.GetNarrationResponse{
		Status: status,
		Path:   path,
	}, err
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
func (h *KatariveHandlerV1) GetSourceItem(
	ctx context.Context,
	req *apb.GetSourceItemRequest,
) (*apb.GetSourceItemResponse, error) {
	var item *apb.SourceItem
	var collection *apb.SourceCollection
	var err error
	if !req.GetDisableCache() {
		i, err := h.sr.GetItem(req.GetUrl())
		if err != nil {
			return nil, err
		}

		cs, err := h.sr.GetCollections(source.WithItemUrl(req.GetUrl()))
		if err != nil {
			return nil, err
		}

		if i != nil {
			item = &apb.SourceItem{
				Id:      i.GetId(),
				Url:     i.GetUrl(),
				Title:   i.GetTitle(),
				Content: i.GetContent(),
			}
		}
		if len(cs) > 0 {
			collection = &apb.SourceCollection{
				Id:          cs[0].GetId(),
				Url:         cs[0].GetUrl(),
				Title:       cs[0].GetTitle(),
				Description: cs[0].GetDescription(),
				Author:      cs[0].GetAuthor(),
				Tags:        cs[0].GetTags(),
			}
		}
	}

	status := apb.JobStatus_JOB_STATUS_COMPLETED
	if item == nil {
		ctx = context.WithoutCancel(ctx)
		job, err := h.siq.Queue(ctx, job.WithSourceItemUrl(req.GetUrl()))
		if err != nil {
			return nil, err
		}
		status = job.Status()
		err = job.Error()
	}

	return &apb.GetSourceItemResponse{
		Status:     status,
		Item:       item,
		Collection: collection,
	}, err
}
func (h *KatariveHandlerV1) GetSourceCollection(
	ctx context.Context,
	req *apb.GetSourceCollectionRequest,
) (*apb.GetSourceCollectionResponse, error) {
	var collection *apb.SourceCollection
	var items []*apb.SourceSummary
	var err error
	if !req.GetDisableCache() {
		c, err := h.sr.GetCollection(req.GetUrl())
		if err != nil {
			return nil, err
		}
		is, err := h.sr.GetItems(source.WithCollectionUrl(req.GetUrl()))
		if err != nil {
			return nil, err
		}

		if c != nil {
			collection = &apb.SourceCollection{
				Id:          c.GetId(),
				Url:         c.GetUrl(),
				Title:       c.GetTitle(),
				Description: c.GetDescription(),
				Author:      c.GetAuthor(),
				Tags:        c.GetTags(),
			}
		}
		if len(is) > 0 {
			for _, i := range is {
				items = append(items, &apb.SourceSummary{
					Id:    i.GetId(),
					Url:   i.GetUrl(),
					Title: i.GetTitle(),
				})
			}
		}
	}

	status := apb.JobStatus_JOB_STATUS_COMPLETED
	if collection == nil {
		ctx = context.WithoutCancel(ctx)
		job, err := h.scq.Queue(ctx, job.WithSourceCollectionUrl(req.GetUrl()))
		if err != nil {
			return nil, err
		}
		status = job.Status()
		err = job.Error()
	}

	return &apb.GetSourceCollectionResponse{
		Status:     status,
		Collection: collection,
		Items:      items,
	}, err
}
func (h *KatariveHandlerV1) GetSourceCollections(
	ctx context.Context,
	req *apb.GetSourceCollectionsRequest,
) (*apb.GetSourceCollectionsResponse, error) {
	cs, err := h.sr.GetCollections()
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
	return &KatariveHandlerV1{
		sr:  sr,
		nr:  nr,
		nq:  job.NewMutexNarrateJobQueue(sr, nr),
		scq: job.NewMutexSourceCollectionJobQueue(sr),
		siq: job.NewMutexSourceItemJobQueue(sr),
		pm:  pm,
	}
}

type PathModifier interface {
	Do(path string) *string
}
type BasePathModifier struct {
	rules []basePathModificationRule
}

func (m *BasePathModifier) Do(path string) *string {
	p := []byte(path)
	for _, rule := range m.rules {
		p = rule.source.ReplaceAll(p, rule.dest)
	}
	path = string(p)
	return &path
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
