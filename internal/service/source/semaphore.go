package source

import (
	"context"
	"fmt"
	"regexp"
	"sync"
	"time"

	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
	"github.com/heptaliane/katarive-server/internal/model"
)

type SemaphoreSourceManager struct {
	client pb.SourceServiceClient

	itemPattern       *regexp.Regexp
	collectionPattern *regexp.Regexp
	name              string
	version           string
	interval          time.Duration

	mu *sync.RWMutex
}

func (s *SemaphoreSourceManager) Name() string {
	return fmt.Sprintf("%s.%s", s.name, s.version)
}
func (s *SemaphoreSourceManager) IsSupportedItem(url string) bool {
	return s.itemPattern.MatchString(url)
}
func (s *SemaphoreSourceManager) IsSupportedCollection(url string) bool {
	return s.collectionPattern.MatchString(url)
}
func (s *SemaphoreSourceManager) GetSourceItem(
	ctx context.Context,
	req *pb.GetSourceItemRequest,
) (*pb.GetSourceItemResponse, error) {
	s.mu.Lock()
	defer func() {
		time.Sleep(s.interval)
		s.mu.Unlock()
	}()

	url := req.GetUrl()
	if !s.IsSupportedItem(url) {
		return nil, &model.UnsupportedSourceURLError{Url: url}
	}

	ir := &pb.GetSourceItemRequest{Url: url}
	return s.client.GetSourceItem(ctx, ir)
}
func (s *SemaphoreSourceManager) GetSourceCollection(
	ctx context.Context,
	req *pb.GetSourceCollectionRequest,
) (*pb.GetSourceCollectionResponse, error) {
	s.mu.Lock()
	defer func() {
		time.Sleep(s.interval)
		s.mu.Unlock()
	}()

	url := req.GetUrl()
	if !s.IsSupportedCollection(url) {
		return nil, &model.UnsupportedSourceURLError{Url: url}
	}

	cr := &pb.GetSourceCollectionRequest{Url: url}
	return s.client.GetSourceCollection(ctx, cr)
}

// Ensure SemaphoreSourceManager implements SourceManager
var _ SourceManager = new(SemaphoreSourceManager)

// -----------------
// Helper components
// -----------------

type semaphoreSourceManagerOptions struct {
	interval time.Duration
}
type SemaphoreSourceManagerOption func(*semaphoreSourceManagerOptions)

func WithInterval(interval_ms int) SemaphoreSourceManagerOption {
	return func(opt *semaphoreSourceManagerOptions) {
		t, err := time.ParseDuration(fmt.Sprintf("%dms", interval_ms))
		if err == nil {
			opt.interval = t
		}
	}
}

func NewSemaphoreSourceManager(
	ctx context.Context,
	client pb.SourceServiceClient,
	opts ...SemaphoreSourceManagerOption,
) (*SemaphoreSourceManager, error) {
	var options semaphoreSourceManagerOptions
	for _, opt := range opts {
		opt(&options)
	}

	req := &pb.GetSourceServiceMetadataRequest{}
	res, err := client.GetSourceServiceMetadata(ctx, req)
	if err != nil {
		return nil, err
	}

	return &SemaphoreSourceManager{
		client:            client,
		itemPattern:       regexp.MustCompile(res.GetSupportedItemPattern()),
		collectionPattern: regexp.MustCompile(res.GetSupportedCollectionPattern()),
		name:              res.GetName(),
		version:           res.GetVersion(),
		interval:          options.interval,
		mu:                new(sync.RWMutex),
	}, nil
}
