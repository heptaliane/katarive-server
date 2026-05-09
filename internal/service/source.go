package service

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
)

// ==============================
// Interfaces for Source handlers
// ==============================

//go:generate mockgen -source=$GOFILE -destination=mock/mock_$GOFILE -package=mock
type SourceManager interface {
	GetSourceItem(ctx context.Context, url string) (*pb.GetSourceItemResponse, error)
	GetSourceCollection(ctx context.Context, url string) (*pb.GetSourceCollectionResponse, error)
	IsSupportedURL(url string) bool
	GetName() string
}

//go:generate mockgen -source=$GOFILE -destination=mock/mock_$GOFILE -package=mock
type SourceRegistry interface {
	SourceItem(ctx context.Context, url string) (*pb.GetSourceItemResponse, error)
	SourceCollection(ctx context.Context, url string) (*pb.GetSourceCollectionResponse, error)
}

// ============================
// SourceManager Implementation
// ============================

// ----------------------
// SemaphoreSourceManager
// ----------------------

type SemaphoreSourceManager struct {
	source pb.SourceServiceClient

	pattern *regexp.Regexp
	name    string
	version string

	mu      *sync.RWMutex
	options *semaphoreSourceManagerOptions
}

func (s *SemaphoreSourceManager) GetSourceItem(
	ctx context.Context,
	url string,
) (*pb.GetSourceItemResponse, error) {
	s.mu.Lock()
	defer func() {
		time.Sleep(s.options.interval)
		s.mu.Unlock()
	}()

	req := &pb.GetSourceItemRequest{
		Url: url,
	}
	return s.source.GetSourceItem(ctx, req)
}
func (s *SemaphoreSourceManager) GetSourceCollection(
	ctx context.Context,
	url string,
) (*pb.GetSourceCollectionResponse, error) {
	s.mu.Lock()
	defer func() {
		time.Sleep(s.options.interval)
		s.mu.Unlock()
	}()

	req := &pb.GetSourceCollectionRequest{
		Url: url,
	}
	return s.source.GetSourceCollection(ctx, req)
}
func (s *SemaphoreSourceManager) IsSupportedURL(url string) bool {
	return s.pattern.Match([]byte(url))
}
func (s *SemaphoreSourceManager) GetName() string {
	return fmt.Sprintf("%s:%s", s.name, s.version)
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
	source pb.SourceServiceClient,
	opts ...SemaphoreSourceManagerOption,
) (*SemaphoreSourceManager, error) {
	var options semaphoreSourceManagerOptions
	for _, opt := range opts {
		opt(&options)
	}

	req := &pb.GetSourceServiceMetadataRequest{}
	res, err := source.GetSourceServiceMetadata(ctx, req)
	if err != nil {
		return nil, err
	}

	return &SemaphoreSourceManager{
		source:  source,
		pattern: regexp.MustCompile(res.GetSupportedPattern()),
		name:    res.GetName(),
		version: res.GetVersion(),
		mu:      new(sync.RWMutex),
		options: &options,
	}, nil
}

// =============================
// SourceRegistry Implementation
// =============================

// ------------------
// FileSourceRegistry
// ------------------

type FileSourceRegistry struct {
	basedir string
	sources []SourceManager
	logger  *slog.Logger
}

func (s *FileSourceRegistry) SourceItem(
	ctx context.Context,
	url string,
) (*pb.GetSourceItemResponse, error) {
	// Find supported SourceManager
	var sm SourceManager
	for _, source := range s.sources {
		if source.IsSupportedURL(url) {
			sm = source
			break
		}
	}
	if sm == nil {
		return nil, &UnsupportedSourceURLError{URL: url}
	}
	s.logger.DebugContext(ctx, "SourceManager found", "type", sm.GetName(), "url", url)

	filename := fmt.Sprintf("%s.json", url2filename(url))
	path := filepath.Join(s.basedir, sm.GetName(), filename)
	if Exists(path) {
		s.logger.DebugContext(ctx, "Source item cache hit", "url", url, "path", path)
		return LoadJson[pb.GetSourceItemResponse](path)
	}

	res, err := sm.GetSourceItem(ctx, url)
	if err != nil {
		return nil, err
	}
	s.logger.DebugContext(ctx, "Source item fetched", "url", url)

	err = DumpJson(path, res)
	if err != nil {
		return nil, err
	}
	s.logger.DebugContext(ctx, "Source item saved", "url", url, "path", path)
	return res, nil
}
func (s *FileSourceRegistry) SourceCollection(
	ctx context.Context,
	url string,
) (*pb.GetSourceCollectionResponse, error) {
	// Find supported SourceManager
	var sm SourceManager
	for _, source := range s.sources {
		if source.IsSupportedURL(url) {
			sm = source
			break
		}
	}
	if sm == nil {
		return nil, &UnsupportedSourceURLError{URL: url}
	}
	s.logger.DebugContext(ctx, "SourceManager found", "type", sm.GetName(), "url", url)

	filename := fmt.Sprintf("%s.collection.json", url2filename(url))
	path := filepath.Join(s.basedir, sm.GetName(), filename)
	if Exists(path) {
		s.logger.DebugContext(ctx, "Source collection cache hit", "url", url, "path", path)
		return LoadJson[pb.GetSourceCollectionResponse](path)
	}

	res, err := sm.GetSourceCollection(ctx, url)
	if err != nil {
		return nil, err
	}
	s.logger.DebugContext(ctx, "Source item fetched", "url", url)

	err = DumpJson(path, res)
	if err != nil {
		return nil, err
	}
	s.logger.DebugContext(ctx, "Source item saved", "url", url, "path", path)
	return res, nil
}

// Ensure FileSourceRegistry implements SourceRegistry
var _ SourceRegistry = new(FileSourceRegistry)

// -----------------
// Helper components
// -----------------

func NewFileSourceRegistry(
	basedir string,
	sources []SourceManager,
) *FileSourceRegistry {
	return &FileSourceRegistry{
		basedir: basedir,
		sources: sources,
		logger:  slog.Default(),
	}
}
