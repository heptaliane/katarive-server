package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	pbmock "github.com/heptaliane/katarive-go-sdk/gen/mock/plugin/v1"
	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/heptaliane/katarive-server/internal/service"
	"github.com/heptaliane/katarive-server/internal/service/mock"
)

func TestSemaphoreSourceManager(t *testing.T) {
	t.Parallel()

	collectionId := "collection-id"
	supportedUrl := "http://example.com/1"
	gsr := &pb.GetSourceItemResponse{
		Item: &pb.SourceItem{
			Id:           "id",
			CollectionId: &collectionId,
			Url:          supportedUrl,
			Title:        "example title",
			Content:      "example content",
			Language:     pb.Language_LANGUAGE_ENGLISH,
		},
	}
	gssmr := &pb.GetSourceServiceMetadataResponse{
		Name:             "example",
		Version:          "v1",
		SupportedPattern: `^http://example\.com/.*`,
	}
	cr := &pb.GetSourceCollectionResponse{
		Sources: []*pb.SourceSummary{
			{
				Id:    "1",
				Title: "example title",
				Url:   "http://example.com/1",
			},
		},
	}

	source := pbmock.NewMockSourceServiceClient(gomock.NewController(t))

	source.EXPECT().GetSourceItem(gomock.Any(), gomock.Any()).
		DoAndReturn(func(
			ctx context.Context,
			req *pb.GetSourceItemRequest,
			opt ...grpc.CallOption,
		) (*pb.GetSourceItemResponse, error) {
			if req.Url == supportedUrl {
				return gsr, nil
			}
			return nil, &service.UnsupportedSourceURLError{URL: req.Url}
		}).AnyTimes()
	source.EXPECT().GetSourceServiceMetadata(gomock.Any(), gomock.Any()).
		Return(gssmr, nil).AnyTimes()
	source.EXPECT().GetSourceCollection(gomock.Any(), gomock.Any()).
		DoAndReturn(func(
			ctx context.Context,
			req *pb.GetSourceCollectionRequest,
			opt ...grpc.CallOption,
		) (*pb.GetSourceCollectionResponse, error) {
			if req.Url == supportedUrl {
				return cr, nil
			}
			return nil, &service.UnsupportedSourceURLError{URL: req.Url}
		}).AnyTimes()

	ctx := context.Background()
	sm, err := service.NewSemaphoreSourceManager(ctx, source)
	if err != nil {
		t.Fatal(err)
	}

	cases := map[string]struct {
		url                    string
		expectedSource         *pb.GetSourceItemResponse
		expectedList           *pb.GetSourceCollectionResponse
		expectedIsError        bool
		expectedIsSupportedURL bool
		expectedName           string
	}{
		"supported": {
			url:                    "http://example.com/1",
			expectedSource:         gsr,
			expectedList:           cr,
			expectedIsError:        false,
			expectedIsSupportedURL: true,
			expectedName:           "example:v1",
		},
		"unsupported": {
			url:                    "http://unsupported.com/1",
			expectedIsError:        true,
			expectedIsSupportedURL: false,
			expectedName:           "example:v1",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			actualSource, err := sm.GetSourceItem(ctx, tc.url)
			if err != nil {
				if !tc.expectedIsError {
					t.Errorf("Unexpected error: %v", err)
					return
				}
			} else {
				if tc.expectedIsError {
					t.Errorf("Error expected but got nil")
					return
				}
				diff := cmp.Diff(actualSource, tc.expectedSource, protocmp.Transform())
				if diff != "" {
					t.Errorf("Unmatched GetSource result (got: -, want: +):\n%s", diff)
					return
				}
			}
			actualList, err := sm.GetSourceCollection(ctx, tc.url)
			if err != nil {
				if !tc.expectedIsError {
					t.Errorf("Unexpected error: %v", err)
					return
				}
			} else {
				if tc.expectedIsError {
					t.Errorf("Error expected but got nil")
					return
				}
				diff := cmp.Diff(actualList, tc.expectedList, protocmp.Transform())
				if diff != "" {
					t.Errorf("Unmatched ListSources result (got: -, want: +):\n%s", diff)
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

func TestFileSourceRegistry(t *testing.T) {
	t.Parallel()

	basedir := t.TempDir()
	source := &pb.GetSourceItemResponse{
		Item: &pb.SourceItem{
			Id:      "id",
			Title:   "title",
			Content: "content",
		},
	}
	collection := &pb.GetSourceCollectionResponse{
		Sources: []*pb.SourceSummary{
			{
				Id:    "1",
				Title: "example title",
				Url:   "http://example.com/1",
			},
		},
	}

	sm := mock.NewMockSourceManager(gomock.NewController(t))
	sms := []service.SourceManager{sm}
	sr := service.NewFileSourceRegistry(basedir, sms)

	supportedUrl := "http://example.com/1"
	unsupportedUrl := "http://unsupported.com/1"
	sm.EXPECT().IsSupportedURL(supportedUrl).Return(true).AnyTimes()
	sm.EXPECT().IsSupportedURL(unsupportedUrl).Return(false).AnyTimes()
	sm.EXPECT().GetName().Return("mock").AnyTimes()
	sm.EXPECT().GetSourceItem(gomock.Any(), supportedUrl).Return(source, nil).Times(1)
	sm.EXPECT().GetSourceCollection(gomock.Any(), supportedUrl).
		Return(collection, nil).Times(1)

	cases := []struct {
		name               string
		url                string
		expectedSource     *pb.GetSourceItemResponse
		expectedCollection *pb.GetSourceCollectionResponse
	}{
		{
			name:               "new_file",
			url:                supportedUrl,
			expectedSource:     source,
			expectedCollection: collection,
		},
		{
			name:               "exists_file",
			url:                supportedUrl,
			expectedSource:     source,
			expectedCollection: collection,
		},
		{
			name:               "unsupported",
			url:                unsupportedUrl,
			expectedSource:     nil,
			expectedCollection: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var errs error
			ctx := context.Background()

			actualItem, err := sr.SourceItem(ctx, tc.url)
			errs = errors.Join(errs, err)

			actualCollection, err := sr.SourceCollection(ctx, tc.url)
			errs = errors.Join(errs, err)

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
				diff := cmp.Diff(actualItem, tc.expectedSource, protocmp.Transform())
				if diff != "" {
					t.Errorf("Unmatched SourceItem result (got: -, want: +):\n%s", diff)
					return
				}
				diff = cmp.Diff(actualCollection, tc.expectedCollection, protocmp.Transform())
				if diff != "" {
					t.Errorf("Unmatched SourceCollection result (got: -, want: +):\n%s", diff)
					return
				}
			}
		})
	}
}
