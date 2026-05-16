package handler_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	pb "github.com/heptaliane/katarive-server/gen/pb/api/v1"
	"github.com/heptaliane/katarive-server/internal/handler"
)

func TestKatariveHandlerV1Narration(t *testing.T) {
	t.Parallel()

	p := VALID_PATH

	sr := newSourceRegistry(t)
	nr := newNarrateRegistry(t)
	pm := handler.NewBasePathModifier()
	h := handler.NewKatariveHandlerV1(sr, nr, pm)

	cases := map[string]struct {
		url              string
		expectedResponse *pb.GetNarrationResponse
		expectedJobError error
	}{
		"valid_url": {
			url: VALID_URL,
			expectedResponse: &pb.GetNarrationResponse{
				Status: pb.JobStatus_JOB_STATUS_COMPLETED,
				Path:   &p,
			},
		},
		"invalid_url": {
			url:              VALID_URL,
			expectedJobError: sie,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			qreq := &pb.QueueNarrationRequest{Url: tc.url}
			qres, err := h.QueueNarration(ctx, qreq)
			if err != nil {
				t.Errorf("QueueNarration returns unexpected error: %v", err)
				return
			}

			jreq := &pb.GetNarrationRequest{Id: qres.GetId()}
			jres, err := h.GetNarration(ctx, jreq)
			if tc.expectedJobError != nil {
				if err == nil {
					t.Errorf("GetNarration doesn't return error")
					return
				}
				if diff := cmp.Diff(tc.expectedJobError.Error(), err.Error()); diff != "" {
					t.Errorf("GetNarration returns unexpected error (-want +got):\n%s", diff)
				}
				return
			}
			if err != nil {
				t.Errorf("GetNarration returns unexpected error: %v", err)
				return
			}
			diff := cmp.Diff(tc.expectedResponse, jres, protocmp.Transform())
			if diff != "" {
				t.Errorf("GetNarration returns unexpected response (-want +got):\n%s", diff)
				return
			}
		})
	}
}
