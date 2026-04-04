package service

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	katarive "github.com/heptaliane/katarive-go-sdk"
	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
)

type SourceManager interface {
	GetSource(ctx context.Context, url string) (*pb.GetSourceResponse, error)
	IsSupportedURL(url string) bool
	GetName() string
}

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

type SemaphoreSourceManager struct {
	source katarive.Source

	pattern *regexp.Regexp
	name    string
	version string

	mu      *sync.RWMutex
	options *semaphoreSourceManagerOptions
}

func (s *SemaphoreSourceManager) GetSource(
	ctx context.Context,
	url string,
) (*pb.GetSourceResponse, error) {
	s.mu.Lock()
	defer func() {
		time.Sleep(s.options.interval)
		s.mu.Unlock()
	}()

	return s.source.GetSource(ctx, url)
}
func (s *SemaphoreSourceManager) IsSupportedURL(url string) bool {
	return s.pattern.Match([]byte(url))
}
func (s *SemaphoreSourceManager) GetName() string {
	return fmt.Sprintf("%s:%s", s.name, s.version)
}

// Ensure SemaphoreSourceManager implements SourceManager
var _ SourceManager = new(SemaphoreSourceManager)

func NewSemaphoreSourceManager(
	ctx context.Context,
	source katarive.Source,
	opts ...SemaphoreSourceManagerOption,
) (*SemaphoreSourceManager, error) {
	var options semaphoreSourceManagerOptions
	for _, opt := range opts {
		opt(&options)
	}

	res, err := source.GetSourceServiceMetadata(ctx)
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

type MockSourceManager struct {
	Source       *pb.GetSourceResponse
	SupportedURL *regexp.Regexp
	Name         string
}

func (m *MockSourceManager) GetSource(
	ctx context.Context,
	url string,
) (*pb.GetSourceResponse, error) {
	return m.Source, nil
}
func (m *MockSourceManager) IsSupportedURL(url string) bool {
	return m.SupportedURL.Match([]byte(url))
}
func (m *MockSourceManager) GetName() string {
	return m.Name
}

// Ensure MockSourceManager implements SourceManager
var _ SourceManager = new(MockSourceManager)

type SourceRegistry struct {
	basedir string
	sources []SourceManager
}

func (s *SourceRegistry) GetSource(
	ctx context.Context,
	url string,
) (*pb.GetSourceResponse, error) {
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

	filename := fmt.Sprintf("%s.json", url2filename(url))
	path := filepath.Join(s.basedir, sm.GetName(), filename)
	if Exists(path) {
		return LoadJson[pb.GetSourceResponse](path)
	}

	res, err := sm.GetSource(ctx, url)
	if err != nil {
		return nil, err
	}

	err = DumpJson(path, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func NewSourceRegistry(
	basedir string,
	sources []SourceManager,
) *SourceRegistry {
	return &SourceRegistry{
		basedir: basedir,
		sources: sources,
	}
}
