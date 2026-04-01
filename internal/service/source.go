package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"

	katarive "github.com/heptaliane/katarive-go-sdk"
	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
)

type SourceManager struct {
	mu       *sync.RWMutex
	interval time.Duration
	source   katarive.Source
	name     string
	version  string
	pattern  *regexp.Regexp
}

func (m *SourceManager) GetSource(
	ctx context.Context,
	url string,
) (*pb.GetSourceResponse, error) {
	m.mu.Lock()
	defer func() {
		time.Sleep(m.interval)
		m.mu.Unlock()
	}()

	return m.source.GetSource(ctx, url)
}
func (m *SourceManager) GetName() string {
	return fmt.Sprintf("%s:%s", m.name, m.version)
}

type SourceRepository struct {
	sources []*SourceManager
}

func (m *SourceRepository) GetSource(
	url string,
) (*SourceManager, error) {

	for _, s := range m.sources {
		if s.pattern.Match([]byte(url)) {
			return s, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("No supported Source plugin found for %s", url))
}

func NewSourceRepository(
	ctx context.Context,
	sources []katarive.Source,
	interval_ms int,
) (*SourceRepository, error) {
	duration, err := time.ParseDuration(fmt.Sprintf("%dms", interval_ms))
	if err != nil {
		return nil, err
	}
	mu := new(sync.RWMutex)

	var sm []*SourceManager
	for _, source := range sources {
		res, err := source.GetSourceServiceMetadata(ctx)
		if err != nil {
			return nil, err
		}

		sm = append(sm, &SourceManager{
			mu:       mu,
			interval: duration,
			source:   source,
			name:     res.GetName(),
			version:  res.GetVersion(),
			pattern:  regexp.MustCompile(res.GetSupportedPattern()),
		})
	}

	return &SourceRepository{
		sources: sm,
	}, nil
}
