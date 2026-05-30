package source_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/heptaliane/katarive-server/internal/model"
	"github.com/heptaliane/katarive-server/internal/service/source"
)

func TestSemaphoreSourceManagerName(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ssc := setupSourceServiceClient(t)
	sm, err := source.NewSemaphoreSourceManager(ctx, ssc)
	if err != nil {
		t.Fatalf("Failed to create SemaphoreSourceManager: %v", err)
	}

	expectedName := "example-name.v1"
	actualName := sm.Name()
	if actualName != expectedName {
		t.Errorf(
			"Unmatched Name: expected %s but got %s",
			expectedName,
			actualName,
		)
		return
	}
}
func TestSemaphoreSourceManagerIsSupportedItem(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ssc := setupSourceServiceClient(t)
	sm, err := source.NewSemaphoreSourceManager(ctx, ssc)
	if err != nil {
		t.Fatalf("Failed to create SemaphoreSourceManager: %v", err)
	}

	cases := map[string]struct {
		url      string
		expected bool
	}{
		"supported": {
			url:      "http://example.com/item/001",
			expected: true,
		},
		"unsupported": {
			url:      "http://unsupported.com/item/001",
			expected: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
		})

		actual := sm.IsSupportedItem(tc.url)
		if actual != tc.expected {
			t.Errorf("Unmatched IsSupported: expected %t but got %t", tc.expected, actual)
			return
		}
	}
}
func TestSemaphoreSourceManagerIsSupportedCollection(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ssc := setupSourceServiceClient(t)
	sm, err := source.NewSemaphoreSourceManager(ctx, ssc)
	if err != nil {
		t.Fatalf("Failed to create SemaphoreSourceManager: %v", err)
	}

	cases := map[string]struct {
		url      string
		expected bool
	}{
		"supported": {
			url:      "http://example.com/collection/001",
			expected: true,
		},
		"unsupported": {
			url:      "http://unsupported.com/collection/001",
			expected: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
		})

		actual := sm.IsSupportedCollection(tc.url)
		if actual != tc.expected {
			t.Errorf("Unmatched IsSupported: expected %t but got %t", tc.expected, actual)
			return
		}
	}
}
func TestSemaphoreSourceManagerGetSourceItem(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ssc := setupSourceServiceClient(t)
	sm, err := source.NewSemaphoreSourceManager(ctx, ssc)
	if err != nil {
		t.Fatalf("Failed to create SemaphoreSourceManager: %v", err)
	}

	cases := map[string]struct {
		url              string
		expectedResponse *pb.GetSourceItemResponse
		expectedError    error
	}{
		"supported": {
			url:              "http://example.com/item/001",
			expectedResponse: &gsir,
		},
		"unsupported": {
			url:           "http://unsupported.com/item/001",
			expectedError: &model.UnsupportedSourceURLError{Url: "http://unsupported.com/item/001"},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			req := &pb.GetSourceItemRequest{Url: tc.url}
			actual, err := sm.GetSourceItem(ctx, req)
			if tc.expectedError == nil {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				diff := cmp.Diff(tc.expectedResponse, actual, protocmp.Transform())
				if diff != "" {
					t.Errorf("response mismatch (-want +got):\n%s", diff)
				}
			} else {
				if err == nil {
					t.Error("expected an error but got nil")
					return
				}
				if diff := cmp.Diff(tc.expectedError.Error(), err.Error()); diff != "" {
					t.Errorf("error message mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
func TestSemaphoreSourceManagerGetSourceCollection(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ssc := setupSourceServiceClient(t)
	sm, err := source.NewSemaphoreSourceManager(ctx, ssc)
	if err != nil {
		t.Fatalf("Failed to create SemaphoreSourceManager: %v", err)
	}

	cases := map[string]struct {
		url              string
		expectedResponse *pb.GetSourceCollectionResponse
		expectedError    error
	}{
		"supported": {
			url:              "http://example.com/collection/001",
			expectedResponse: &gscr,
		},
		"unsupported": {
			url:           "http://unsupported.com/collection/001",
			expectedError: &model.UnsupportedSourceURLError{Url: "http://unsupported.com/collection/001"},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			req := &pb.GetSourceCollectionRequest{Url: tc.url}
			actual, err := sm.GetSourceCollection(ctx, req)
			if tc.expectedError == nil {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				diff := cmp.Diff(tc.expectedResponse, actual, protocmp.Transform())
				if diff != "" {
					t.Errorf("response mismatch (-want +got):\n%s", diff)
				}
			} else {
				if err == nil {
					t.Error("expected an error but got nil")
					return
				}
				if diff := cmp.Diff(tc.expectedError.Error(), err.Error()); diff != "" {
					t.Errorf("error message mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
