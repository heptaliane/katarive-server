package service_test

import (
	"testing"
	"context"

	katarive "github.com/heptaliane/katarive-go-sdk"
	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
	"github.com/google/go-cmp/cmp"

	"github.com/heptaliane/katarive-server/internal/plugin"
	"github.com/heptaliane/katarive-server/internal/service"
)

func TestSourceManager(t *testing.T) {
	t.Parallel()

	sources := []katarive.Source{
		&plugin.MockSource {
			Patterns: []string{`^http://example\.com/.*`},
			Title: "example",
			Content: "This is the example",
			NextUrl: "http://example.com/2",
		},
	}
	ctx := context.Background()
	sm, err := service.NewSourceManager(ctx, sources)
	if err != nil {
		t.Fatal(err)
	}

	cases := map[string]struct {
		url string
		res *pb.GetSourceResponse
	} {
		"normal": {
			url: "http://example.com/1",
			res: &pb.GetSourceResponse {
				Title: "example",
				Content: "This is the example",
				NextUrl: "http://example.com/2",
			},
		},
		"not_found": {
			url: "http://not_found.com/1",
			res: nil,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx = context.Background()
			src, err := sm.GetSource(ctx, tc.url)

			if tc.res == nil {
				if err == nil {
					t.Errorf("Error expected, but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			if diff := cmp.Diff(src.GetTitle(), tc.res.GetTitle()); diff != "" {
				t.Errorf("Title unmatched: %s", diff)
				return
			}
			if diff := cmp.Diff(src.GetContent(), tc.res.GetContent()); diff != "" {
				t.Errorf("Content unmatched: %s", diff)
				return
			}
			if diff := cmp.Diff(src.GetNextUrl(), tc.res.GetNextUrl()); diff != "" {
				t.Errorf("NextUrl unmatched: %s", diff)
				return
			}
		})
	}
}
