package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	katarive "github.com/heptaliane/katarive-go-sdk"
	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
)

type sourceRegistry struct {
	source  katarive.Source
	pattern *regexp.Regexp
}

type SourceManager struct {
	sources []*sourceRegistry
}

func (m *SourceManager) GetSource(
	ctx context.Context,
	url string,
) (*pb.GetSourceResponse, error) {
	for _, s := range m.sources {
		if s.pattern.Match([]byte(url)) {
			return s.source.GetSource(ctx, url)
		}
	}

	return nil, errors.New(fmt.Sprintf("No supported Source plugin found for %s", url))
}

func NewSourceManager(ctx context.Context, sources []katarive.Source) (*SourceManager, error) {
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

	return &SourceManager{
		sources: registries,
	}, nil
}
