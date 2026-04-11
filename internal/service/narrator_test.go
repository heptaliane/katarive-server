package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
	"go.uber.org/mock/gomock"

	"github.com/heptaliane/katarive-server/internal/plugin"
	"github.com/heptaliane/katarive-server/internal/service"
	"github.com/heptaliane/katarive-server/internal/service/mock"
)

func TestSemaphoreNarratorManager(t *testing.T) {
	t.Parallel()

	options := []*pb.NarratorOption{
		{
			Id:          "id-1",
			Label:       "label-1",
			Description: "description-1",
		},
	}
	narratorName := "narrator"
	version := "v1"
	reason := "error reason"

	cases := map[string]struct {
		err                      error
		reason                   *string
		expectedError            error
		expectedName             string
		expectedSupportedOptions []*pb.NarratorOption
	}{
		"success": {
			expectedName:             "narrator:v1",
			expectedSupportedOptions: options,
		},
		"failed with reason": {
			err:                      nil,
			reason:                   &reason,
			expectedError:            &service.NarrateError{Reason: "fail"},
			expectedName:             "narrator:v1",
			expectedSupportedOptions: options,
		},
		"failed before pb": {
			err:                      errors.New("some error"),
			reason:                   &reason,
			expectedError:            errors.New("some error"),
			expectedName:             "narrator:v1",
			expectedSupportedOptions: options,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			narrator := &plugin.MockNarrator{
				NarrateError: tc.err,
				NarrateResponse: &pb.NarrateResponse{
					Error:  tc.err != nil,
					Reason: tc.reason,
				},
				GetNarratorServiceMetadataResponse: &pb.GetNarratorServiceMetadataResponse{
					Name:    narratorName,
					Version: version,
					Options: options,
				},
			}

			ctx := context.Background()
			nm, err := service.NewSemaphoreNarratorManager(ctx, narrator)
			if err != nil {
				t.Errorf("Failed to create NarratorManager: %v", err)
				return
			}

			err = nm.Do(ctx, "dummy.wav", "text")
			if errors.Is(err, tc.expectedError) && err != tc.expectedError {
				t.Errorf(
					"Unmatched Narrate result: expected %v but got %v",
					tc.expectedError,
					err,
				)
				return
			}

			actualName := nm.GetName()
			if actualName != tc.expectedName {
				t.Errorf("Unmatched Name: expected %s but got %s", tc.expectedName, actualName)
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
	nm.EXPECT().Do(gomock.Any(), gomock.Any(), text).Return(nil).
		Do(func(ctx context.Context, path string, text string, opts ...service.NarrateOption) {
			f, err := service.NewFile(path)
			if err != nil {
				t.Fatalf("Failed to create file: %v", err)
			}
			defer f.Close()
		}).Times(1)

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
