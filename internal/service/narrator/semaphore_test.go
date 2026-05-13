package narrator_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/heptaliane/katarive-server/internal/service/narrator"
)

func TestSemaphoreNarratorManagerNarrate(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		text             string
		expectedResponse *pb.NarrateResponse
		isError          bool
	}{
		"valid": {
			text:             VALID_TEXT,
			expectedResponse: nr,
			isError:          false,
		},
		"invalid": {
			text:    "invalid",
			isError: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			client := setupNarratorServiceClient(t)
			snm, err := narrator.NewSemaphoreNarratorManager(ctx, client)
			if err != nil {
				t.Errorf("Failed to create NarratorManager: %v", err)
				return
			}

			req := &pb.NarrateRequest{Text: tc.text}
			res, err := snm.Narrate(ctx, req)
			if tc.isError {
				if err == nil {
					t.Error("Error expected but got nil")
					return
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			diff := cmp.Diff(tc.expectedResponse, res, protocmp.Transform())
			if diff != "" {
				t.Errorf("Narrate mismatch (-want +got):\n%s", diff)
				return
			}
		})
	}
}
