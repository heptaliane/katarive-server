package service_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	katarive "github.com/heptaliane/katarive-go-sdk"

	"github.com/heptaliane/katarive-server/internal/plugin"
	"github.com/heptaliane/katarive-server/internal/service"
)

func TestSourceManager(t *testing.T) {
	t.Parallel()

	sources := []katarive.Source{
		&plugin.MockSource{
			Name:             "example",
			Version:          "v1",
			SupportedPattern: `^http://example\.com/.*`,
		},
	}
	ctx := context.Background()
	sm, err := service.NewSourceRepository(ctx, sources, 10)
	if err != nil {
		t.Fatal(err)
	}

	cases := map[string]struct {
		url  string
		name string
	}{
		"normal": {
			url:  "http://example.com/1",
			name: "example:v1",
		},
		"not_found": {
			url:  "http://not_found.com/1",
			name: "",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx = context.Background()
			sm, err := sm.GetSource(tc.url)

			if tc.name == "" {
				if err == nil {
					t.Errorf("Error expected, but got nil")
					return
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			if diff := cmp.Diff(sm.GetName(), tc.name); diff != "" {
				t.Errorf("Unexpected SourceManager name: %s", diff)
			}

			_, err = sm.GetSource(ctx, tc.url)
			if err != nil {
				t.Errorf("Unexpected error in GetSource: %v", err)
				return
			}
		})
	}
}
