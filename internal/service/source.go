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

type sourceRegistry struct {
	source  katarive.Source
	pattern *regexp.Regexp
}

type SourceManager struct {
	mu       sync.RWMutex
	interval time.Duration
	sources  []*sourceRegistry
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

	for _, s := range m.sources {
		if s.pattern.Match([]byte(url)) {
			return s.source.GetSource(ctx, url)
		}
	}

	return nil, errors.New(fmt.Sprintf("No supported Source plugin found for %s", url))
}

func NewSourceManager(
	ctx context.Context,
	sources []katarive.Source,
	interval int,
) (*SourceManager, error) {
	var registries []*sourceRegistry
	for _, source := range sources {
		res, err := source.GetSupportedPatterns(ctx)
		if err != nil {
			return nil, err
		}

		for _, pattern := range res.GetPatterns() {
			registries = append(registries, &sourceRegistry{
				source:  source,
				pattern: regexp.MustCompile(pattern),
			})
		}
	}

	duration, err := time.ParseDuration(fmt.Sprintf("%dms", interval))
	if err != nil {
		return nil, err
	}

	return &SourceManager{
		interval: duration,
		sources:  registries,
	}, nil
}
