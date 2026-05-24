package handler_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	pb "github.com/heptaliane/katarive-server/gen/pb/api/v1"
)

func TestKatariveHandlerV1Narration(t *testing.T) {
	t.Parallel()

	p := VALID_PATH

	h := newKatariveHandlerV1(t)

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
				Source: &pb.SourceSummary{
					Id:    "item-id",
					Url:   VALID_URL,
					Title: "item-title",
				},
			},
		},
		"invalid_url": {
			url: "http://invalid.com",
			expectedResponse: &pb.GetNarrationResponse{
				Status: pb.JobStatus_JOB_STATUS_FAILED,
			},
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

			time.Sleep(interval)

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
			} else if err != nil {
				t.Errorf("GetNarration returns unexpected error: %v", err)
			}

			diff := cmp.Diff(tc.expectedResponse, jres, protocmp.Transform())
			if diff != "" {
				t.Errorf("GetNarration returns unexpected response (-want +got):\n%s", diff)
				return
			}
		})
	}
}
func TestKatariveHandlerV1GetNarrators(t *testing.T) {
	t.Parallel()

	h := newKatariveHandlerV1(t)

	cases := map[string]struct {
		expected *pb.GetNarratorsResponse
	}{
		"valid": {
			expected: &pb.GetNarratorsResponse{
				Narrator: []*pb.Narrator{
					{
						Name: "narrator1",
						Speakers: []*pb.Speaker{
							{Id: 1, Label: "narrator1-name1"},
						},
					},
					{
						Name: "narrator2",
						Speakers: []*pb.Speaker{
							{Id: 1, Label: "narrator2-name1"},
							{Id: 2, Label: "narrator2-name2"},
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			req := &pb.GetNarratorsRequest{}
			res, err := h.GetNarrators(ctx, req)
			if err != nil {
				t.Errorf("GetNarrators returns unexpected error: %v", err)
				return
			}
			diff := cmp.Diff(tc.expected, res, protocmp.Transform())
			if diff != "" {
				t.Errorf("GetNarrators unmatch (-want +got):\n%s", diff)
				return
			}
		})
	}
}
func TestKatariveHandlerV1SourceItem(t *testing.T) {
	t.Parallel()

	h := newKatariveHandlerV1(t)

	c := si.GetContent()

	cases := map[string]struct {
		url              string
		expectedResponse *pb.GetSourceItemResponse
		expectedJobError error
	}{
		"valid_url": {
			url: VALID_URL,
			expectedResponse: &pb.GetSourceItemResponse{
				Status: pb.JobStatus_JOB_STATUS_COMPLETED,
				Metadata: &pb.SourceSummary{
					Id:    si.GetId(),
					Url:   si.GetUrl(),
					Title: si.GetTitle(),
				},
				Content: &c,
			},
		},
		"invalid_url": {
			url: "http://invalid.com",
			expectedResponse: &pb.GetSourceItemResponse{
				Status: pb.JobStatus_JOB_STATUS_FAILED,
			},
			expectedJobError: sie,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			qreq := &pb.QueueSourceItemRequest{Url: tc.url}
			qres, err := h.QueueSourceItem(ctx, qreq)
			if err != nil {
				t.Errorf("QueueSourceItem returns unexpected error: %v", err)
				return
			}

			time.Sleep(interval)

			jreq := &pb.GetSourceItemRequest{Id: qres.GetId()}
			jres, err := h.GetSourceItem(ctx, jreq)
			if tc.expectedJobError != nil {
				if err == nil {
					t.Errorf("GetSourceItem doesn't return error")
					return
				}
				if diff := cmp.Diff(tc.expectedJobError.Error(), err.Error()); diff != "" {
					t.Errorf("GetSourceItem returns unexpected error (-want +got):\n%s", diff)
				}
			} else if err != nil {
				t.Errorf("GetSourceItem returns unexpected error: %v", err)
			}

			diff := cmp.Diff(tc.expectedResponse, jres, protocmp.Transform())
			if diff != "" {
				t.Errorf("GetSourceItem returns unexpected response (-want +got):\n%s", diff)
				return
			}
		})
	}
}
func TestKatariveHandlerV1SourceCollection(t *testing.T) {
	t.Parallel()

	h := newKatariveHandlerV1(t)

	cases := map[string]struct {
		url              string
		expectedResponse *pb.GetSourceCollectionResponse
		expectedJobError error
	}{
		"valid_url": {
			url: VALID_URL,
			expectedResponse: &pb.GetSourceCollectionResponse{
				Status: pb.JobStatus_JOB_STATUS_COMPLETED,
				Collection: &pb.SourceCollection{
					Id:          sc.GetId(),
					Url:         sc.GetUrl(),
					Title:       sc.GetTitle(),
					Description: sc.GetDescription(),
					Author:      sc.GetAuthor(),
					Tags:        sc.GetTags(),
				},
				Sources: []*pb.SourceSummary{
					{Id: "item1", Title: "title1", Url: "http://valid.com/1"},
					{Id: "item2", Title: "title2", Url: "http://valid.com/2"},
				},
			},
		},
		"invalid_url": {
			url: "http://invalid.com",
			expectedResponse: &pb.GetSourceCollectionResponse{
				Status: pb.JobStatus_JOB_STATUS_FAILED,
			},
			expectedJobError: sce,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			qreq := &pb.QueueSourceCollectionRequest{Url: tc.url}
			qres, err := h.QueueSourceCollection(ctx, qreq)
			if err != nil {
				t.Errorf("QueueSourceCollection returns unexpected error: %v", err)
				return
			}

			time.Sleep(interval)

			jreq := &pb.GetSourceCollectionRequest{Id: qres.GetId()}
			jres, err := h.GetSourceCollection(ctx, jreq)
			if tc.expectedJobError != nil {
				if err == nil {
					t.Errorf("GetSourceCollection doesn't return error")
					return
				}
				if diff := cmp.Diff(tc.expectedJobError.Error(), err.Error()); diff != "" {
					t.Errorf(
						"GetSourceCollection returns unexpected error (-want +got):\n%s", diff,
					)
				}
			} else if err != nil {
				t.Errorf("GetSourceCollection returns unexpected error: %v", err)
			}

			diff := cmp.Diff(tc.expectedResponse, jres, protocmp.Transform())
			if diff != "" {
				t.Errorf(
					"GetSourceCollection returns unexpected response (-want +got):\n%s", diff,
				)
				return
			}
		})
	}
}
