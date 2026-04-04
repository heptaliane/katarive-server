package service_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"

	"github.com/heptaliane/katarive-server/internal/plugin"
	"github.com/heptaliane/katarive-server/internal/service"
)

func TestSemaphoreSourceManager(t *testing.T) {
	t.Parallel()

	source := &plugin.MockSource{
		Name:             "example",
		Version:          "v1",
		SupportedPattern: `^http://example\.com/.*`,
		Title:            "example title",
		Content:          "example content",
		NextUrl:          "http://example.com/2",
	}
	ctx := context.Background()
	sm, err := service.NewSemaphoreSourceManager(ctx, source)
	if err != nil {
		t.Fatal(err)
	}

	cases := map[string]struct {
		url                    string
		expectedSource         *pb.GetSourceResponse
		expectedIsSupportedURL bool
		expectedName           string
	}{
		"supported": {
			url: "http://example.com/1",
			expectedSource: &pb.GetSourceResponse{
				Title:   source.Title,
				Content: source.Content,
				NextUrl: source.NextUrl,
			},
			expectedIsSupportedURL: true,
			expectedName:           source.Name,
		},
		"unsupported": {
			url:                    "http://unsupported.com/1",
			expectedSource:         nil,
			expectedIsSupportedURL: false,
			expectedName:           source.Name,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			actualSource, _ := sm.GetSource(ctx, tc.url)
			if diff := cmp.Diff(actualSource, tc.expectedSource); diff != "" {
				t.Errorf("Unmatched GetSource result (want: -, got: +): %s", diff)
				return
			}
			actualIsSupportedURL := sm.IsSupportedURL(tc.url)
			if actualIsSupportedURL != tc.expectedIsSupportedURL {
				t.Errorf(
					"Expceted %s but got %s for IsSupportedURL",
					tc.expectedIsSupportedURL,
					actualIsSupportedURL,
				)
				return
			}
			actualName := sm.GetName()
			if actualName != tc.expectedName {
				t.Errorf(
					"Expceted %s but got %s for GetName",
					tc.expectedName,
					actualName,
				)
				return
			}
		})
	}
}
