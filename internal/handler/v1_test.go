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
			},
		},
		"invalid_url": {
			url:              "http://invalid.com",
			expectedJobError: sie,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			req := &pb.GetNarrationRequest{Url: tc.url}
			_, err := h.GetNarration(ctx, req)
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

			time.Sleep(interval)

			res, err := h.GetNarration(ctx, req)
			if tc.expectedJobError != nil {
				if err == nil {
					t.Errorf("GetNarration doesn't return error")
					return
				}
				if diff := cmp.Diff(tc.expectedJobError.Error(), err.Error()); diff != "" {
					t.Errorf("GetNarration returns unexpected error (-want +got):\n%s", diff)
				}
				return
			} else if err != nil {
				t.Errorf("GetNarration returns unexpected error: %v", err)
			}

			diff := cmp.Diff(tc.expectedResponse, res, protocmp.Transform())
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

	cases := map[string]struct {
		url              string
		expectedResponse *pb.GetSourceItemResponse
		expectedJobError error
	}{
		"valid_url": {
			url: VALID_URL,
			expectedResponse: &pb.GetSourceItemResponse{
				Status: pb.JobStatus_JOB_STATUS_COMPLETED,
				Item: &pb.SourceItem{
					Id:      si.GetId(),
					Url:     si.GetUrl(),
					Title:   si.GetTitle(),
					Content: si.GetContent(),
				},
				Collection: &pb.SourceCollection{
					Id:          scs[0].GetId(),
					Url:         scs[0].GetUrl(),
					Title:       scs[0].GetTitle(),
					Description: scs[0].GetDescription(),
					Author:      scs[0].GetAuthor(),
					Tags:        scs[0].GetTags(),
				},
			},
		},
		"invalid_url": {
			url:              "http://invalid.com",
			expectedJobError: sie,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			req := &pb.GetSourceItemRequest{Url: tc.url}
			_, err := h.GetSourceItem(ctx, req)
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

			time.Sleep(interval)

			res, err := h.GetSourceItem(ctx, req)
			if tc.expectedJobError != nil {
				if err == nil {
					t.Errorf("GetSourceItem doesn't return error")
					return
				}
				if diff := cmp.Diff(tc.expectedJobError.Error(), err.Error()); diff != "" {
					t.Errorf("GetSourceItem returns unexpected error (-want +got):\n%s", diff)
				}
				return
			} else if err != nil {
				t.Errorf("GetSourceItem returns unexpected error: %v", err)
			}

			diff := cmp.Diff(tc.expectedResponse, res, protocmp.Transform())
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
				Items: []*pb.SourceSummary{
					{Id: "item1", Title: "title1", Url: "http://valid.com/1"},
					{Id: "item2", Title: "title2", Url: "http://valid.com/2"},
				},
			},
		},
		"invalid_url": {
			url:              "http://invalid.com",
			expectedJobError: sce,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			req := &pb.GetSourceCollectionRequest{Url: tc.url}
			_, err := h.GetSourceCollection(ctx, req)
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

			time.Sleep(interval)

			res, err := h.GetSourceCollection(ctx, req)
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
				return
			} else if err != nil {
				t.Errorf("GetSourceCollection returns unexpected error: %v", err)
			}

			diff := cmp.Diff(tc.expectedResponse, res, protocmp.Transform())
			if diff != "" {
				t.Errorf(
					"GetSourceCollection returns unexpected response (-want +got):\n%s", diff,
				)
				return
			}
		})
	}
}
