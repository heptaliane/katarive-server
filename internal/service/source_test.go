package service_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	pbmock "github.com/heptaliane/katarive-go-sdk/gen/mock/plugin/v1"
	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"

	"github.com/heptaliane/katarive-server/internal/service"
	"github.com/heptaliane/katarive-server/internal/service/mock"
)

func TestSemaphoreSourceManager(t *testing.T) {
	t.Parallel()

	gsr := &pb.GetSourceResponse{
		Title:    "example title",
		Content:  "example content",
		Language: pb.Language_LANGUAGE_ENGLISH,
		NextUrl:  "http://example.com/2",
	}
	gssmr := &pb.GetSourceServiceMetadataResponse{
		Name:             "example",
		Version:          "v1",
		SupportedPattern: `^http://example\.com/.*`,
	}
	supportedUrl := "http://example.com/1"

	source := pbmock.NewMockSourceServiceClient(gomock.NewController(t))

	source.EXPECT().GetSource(gomock.Any(), gomock.Any()).
		DoAndReturn(func(
			ctx context.Context,
			req *pb.GetSourceRequest,
			opt ...grpc.CallOption,
		) (*pb.GetSourceResponse, error) {
			if req.Url == supportedUrl {
				return gsr, nil
			}
			return nil, &service.UnsupportedSourceURLError{URL: req.Url}
		}).AnyTimes()
	source.EXPECT().GetSourceServiceMetadata(gomock.Any(), gomock.Any()).
		Return(gssmr, nil).AnyTimes()

	ctx := context.Background()
	sm, err := service.NewSemaphoreSourceManager(ctx, source)
	if err != nil {
		t.Fatal(err)
	}

	cases := map[string]struct {
		url                    string
		expectedSource         *pb.GetSourceResponse
		expectedIsError        bool
		expectedIsSupportedURL bool
		expectedName           string
	}{
		"supported": {
			url:                    "http://example.com/1",
			expectedSource:         gsr,
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
			actualSource, err := sm.GetSource(ctx, tc.url)
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

func TestFileSourceRegistry(t *testing.T) {
	t.Parallel()

	basedir := t.TempDir()
	source := &pb.GetSourceResponse{
		Title:   "title",
		Content: "content",
		NextUrl: "http://example.com/2",
	}

	sm := mock.NewMockSourceManager(gomock.NewController(t))
	sms := []service.SourceManager{sm}
	sr := service.NewFileSourceRegistry(basedir, sms)

	supportedUrl := "http://example.com/1"
	unsupportedUrl := "http://unsupported.com/1"
	sm.EXPECT().IsSupportedURL(supportedUrl).Return(true).AnyTimes()
	sm.EXPECT().IsSupportedURL(unsupportedUrl).Return(false).AnyTimes()
	sm.EXPECT().GetName().Return("mock").AnyTimes()
	sm.EXPECT().GetSource(gomock.Any(), supportedUrl).Return(source, nil).Times(1)

	cases := []struct {
		name           string
		url            string
		expectedSource *pb.GetSourceResponse
	}{
		{
			name:           "new_file",
			url:            supportedUrl,
			expectedSource: source,
		},
		{
			name:           "exists_file",
			url:            supportedUrl,
			expectedSource: source,
		},
		{
			name:           "unsupported",
			url:            unsupportedUrl,
			expectedSource: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			actualSource, err := sr.Get(ctx, tc.url)
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
