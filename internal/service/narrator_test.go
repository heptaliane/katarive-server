package service_test

import (
	"context"
	"errors"
	"fmt"
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

func TestSemaphoreNarratorManager(t *testing.T) {
	t.Parallel()

	validText := "valid"
	invalidText := "invalid"
	invalidReason := "invalid reason"
	narrateError := errors.New("some error")
	validEncoding := pb.AudioEncoding_AUDIO_ENCODING_MP3
	invalidEncoding := pb.AudioEncoding_AUDIO_ENCODING_M4A
	basepath := "dummy"
	validPath := "dummy.mp3"
	options := []*pb.NarratorOption{
		{
			Id:          "id-1",
			Label:       "label-1",
			Description: "description-1",
		},
	}
	gnsmr := &pb.GetNarratorServiceMetadataResponse{
		Name:              "narrator",
		Version:           "v1",
		SupportedEncoding: []pb.AudioEncoding{validEncoding},
		Options:           options,
	}

	narrator := pbmock.NewMockNarratorServiceClient(gomock.NewController(t))
	narrator.EXPECT().GetNarratorServiceMetadata(gomock.Any(), gomock.Any()).
		Return(gnsmr, nil).AnyTimes()
	narrator.EXPECT().Narrate(gomock.Any(), gomock.Any()).
		DoAndReturn(func(
			ctx context.Context,
			req *pb.NarrateRequest,
			opt ...grpc.CallOption,
		) (*pb.NarrateResponse, error) {
			switch req.GetText() {
			case validText:
				return &pb.NarrateResponse{Error: false}, nil
			case invalidText:
				return &pb.NarrateResponse{Error: true, Reason: &invalidReason}, nil
			}
			return nil, narrateError
		}).AnyTimes()

	cases := map[string]struct {
		text                     string
		options                  []service.NarrateOption
		expectedError            error
		expectedName             string
		expectedPath             string
		expectedSupportedOptions []*pb.NarratorOption
	}{
		"success": {
			text: validText,
			options: []service.NarrateOption{
				service.WithNarrateEncoding(validEncoding),
			},
			expectedName:             "narrator:v1",
			expectedPath:             validPath,
			expectedSupportedOptions: options,
		},
		"failed with reason": {
			text: invalidText,
			options: []service.NarrateOption{
				service.WithNarrateEncoding(validEncoding),
			},
			expectedError:            &service.NarrateError{Reason: invalidReason},
			expectedName:             "narrator:v1",
			expectedSupportedOptions: options,
		},
		"failed before pb": {
			text: "error",
			options: []service.NarrateOption{
				service.WithNarrateEncoding(validEncoding),
			},
			expectedError:            narrateError,
			expectedName:             "narrator:v1",
			expectedSupportedOptions: options,
		},
		"failed with encoding": {
			text: validText,
			options: []service.NarrateOption{
				service.WithNarrateEncoding(invalidEncoding),
			},
			expectedError: &service.UnsupportedEncodingError{
				Target:   "narrator:v1",
				Encoding: invalidEncoding.String(),
			},
			expectedName:             "narrator:v1",
			expectedSupportedOptions: options,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			nm, err := service.NewSemaphoreNarratorManager(ctx, narrator)
			if err != nil {
				t.Errorf("Failed to create NarratorManager: %v", err)
				return
			}

			path, err := nm.Do(ctx, basepath, tc.text, tc.options...)
			if tc.expectedError == nil {
				if err != nil {
					t.Errorf("Unexpected Error: %v", err)
					return
				}
			} else {
				if diff := cmp.Diff(err.Error(), tc.expectedError.Error()); diff != "" {
					t.Errorf(
						"Unmatched Narrate result: expected '%v' but got '%v'",
						tc.expectedError,
						err,
					)
					return
				}
			}
			if path != tc.expectedPath {
				t.Errorf("Unmatched path: expected '%s' but got '%s'", tc.expectedPath, path)
				return
			}

			actualName := nm.GetName()
			if actualName != tc.expectedName {
				t.Errorf(
					"Unmatched Name: expected '%s' but got '%s'",
					tc.expectedName,
					actualName,
				)
				return
			}

			options := nm.SupportedOptions()
			opt := cmpopts.IgnoreUnexported(pb.NarratorOption{})
			if diff := cmp.Diff(tc.expectedSupportedOptions, options, opt); diff != "" {
				t.Errorf("Unmatched SupportedOptions (got: -, want: +):\n%s", diff)
				return
			}
		})
	}
}

func TestNarratorRegistry(t *testing.T) {
	t.Parallel()

	basedir := t.TempDir()

	nm := mock.NewMockNarratorManager(gomock.NewController(t))
	nms := []service.NarratorManager{nm}

	text := "text"
	name := "mock"
	url := "http://example.com/1"
	nm.EXPECT().GetName().Return(name).AnyTimes()
	nm.EXPECT().Do(gomock.Any(), gomock.Any(), text).
		DoAndReturn(func(
			ctx context.Context,
			basePath string,
			text string,
			opts ...service.NarrateOption,
		) (string, error) {
			path := fmt.Sprintf("%s.mp3", basePath)
			f, err := service.NewFile(path)
			if err != nil {
				t.Fatalf("Failed to create file: %v", err)
			}
			defer f.Close()
			return path, nil
		}).AnyTimes()

	nr := service.NewFileNarratorRegistry(basedir, nms)

	cases := []struct {
		label         string
		name          string
		text          string
		expectedError error
	}{
		{
			label:         "valid",
			name:          name,
			text:          text,
			expectedError: nil,
		},
		{
			label:         "exists",
			name:          name,
			text:          text,
			expectedError: nil,
		},
		{
			label:         "invalid",
			name:          "unsupported",
			expectedError: service.UnspecifiedNarratorError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			ctx := context.Background()

			nr.Use(tc.name)
			path, err := nr.Do(ctx, url, tc.text)
			if tc.expectedError == nil {
				if err != nil {
					t.Errorf("Unexpceted error: %v", err)
					return
				}
				if path == "" {
					t.Errorf("Blank path")
					return
				}
			} else {
				if !errors.Is(err, tc.expectedError) {
					t.Errorf("Unexpected error: expected %v but got %v", tc.expectedError, err)
					return
				}
			}
		})
	}
}
