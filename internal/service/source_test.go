package service_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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
			expectedName:           "example:v1",
		},
		"unsupported": {
			url:                    "http://unsupported.com/1",
			expectedSource:         nil,
			expectedIsSupportedURL: false,
			expectedName:           "example:v1",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.expectedIsSupportedURL {
				ctx := context.Background()
				actualSource, _ := sm.GetSource(ctx, tc.url)
				opt := cmpopts.IgnoreUnexported(pb.GetSourceResponse{})
				if diff := cmp.Diff(actualSource, tc.expectedSource, opt); diff != "" {
					t.Errorf("Unmatched GetSource result (got: -, want: +):\n%s", diff)
					return
				}
			}
			actualIsSupportedURL := sm.IsSupportedURL(tc.url)
			if actualIsSupportedURL != tc.expectedIsSupportedURL {
				t.Errorf(
					"Expceted %t but got %t for IsSupportedURL",
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

func TestSourceRegistry(t *testing.T) {
	t.Parallel()

	basedir := t.TempDir()

	source := &pb.GetSourceResponse{
		Title:   "title",
		Content: "content",
		NextUrl: "http://example.com/2",
	}

	sms := []service.SourceManager{
		&service.MockSourceManager{
			Source:       source,
			SupportedURL: regexp.MustCompile(`http://example\.com/.*`),
			Name:         "name",
		},
	}
	sr := service.NewSourceRegistry(basedir, sms)

	cases := []struct {
		name           string
		url            string
		expectedSource *pb.GetSourceResponse
	}{
		{
			name:           "new_file",
			url:            "http://example.com/1",
			expectedSource: source,
		},
		{
			name:           "exists_file",
			url:            "http://example.com/1",
			expectedSource: source,
		},
		{
			name:           "unsupported",
			url:            "http://unsupported.com",
			expectedSource: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			actualSource, err := sr.GetSource(ctx, tc.url)
			if tc.expectedSource == nil {
				if err == nil {
					t.Errorf("Expect error but got nil")
					return
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				opt := cmpopts.IgnoreUnexported(pb.GetSourceResponse{})
				if diff := cmp.Diff(actualSource, tc.expectedSource, opt); diff != "" {
					t.Errorf("Unmatched GetSource result (got: -, want: +):\n%s", diff)
					return
				}
			}
		})
	}
}
