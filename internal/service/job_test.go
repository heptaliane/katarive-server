package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
	"go.uber.org/mock/gomock"

	"github.com/heptaliane/katarive-server/internal/service"
	"github.com/heptaliane/katarive-server/internal/service/mock"
)

func TestNarrateJobManager(t *testing.T) {
	t.Parallel()

	nr := mock.NewMockNarratorRegistry(gomock.NewController(t))
	sr := mock.NewMockSourceRegistry(gomock.NewController(t))
	jm := service.NewNarrateJobManager(nr, sr)

	validUrl := "http://example.com"
	invalidUrl := "http://invalid.com"
	narrateErrorUrl := "http://narrrate-error.com"
	source := &pb.GetSourceResponse{
		Content:  "source_content",
		Language: pb.Language_LANGUAGE_ENGLISH,
	}
	getSourceError := errors.New("Source.Get failed")
	result := "ok-result"
	narrateError := errors.New("Narrator.Do failed")
	jobInterval, err := time.ParseDuration("10ms")
	if err != nil {
		t.Fatalf("Failed to create jobInterval: %v", err)
	}

	sr.EXPECT().Get(gomock.Any(), validUrl).Return(source, nil).AnyTimes()
	sr.EXPECT().Get(gomock.Any(), narrateErrorUrl).Return(source, nil).AnyTimes()
	sr.EXPECT().Get(gomock.Any(), invalidUrl).Return(nil, getSourceError).AnyTimes()
	nr.EXPECT().Do(
		gomock.Any(),
		validUrl,
		source.GetContent(),
		gomock.Any(),
	).Return(result, nil).AnyTimes()
	nr.EXPECT().Do(
		gomock.Any(),
		narrateErrorUrl,
		source.GetContent(),
		gomock.Any(),
	).Return("", narrateError).AnyTimes()

	cases := map[string]struct {
		url            string
		expectedError  error
		expectedResult string
	}{
		"valid": {
			url:            validUrl,
			expectedResult: result,
		},
		"invalidSource": {
			url:           invalidUrl,
			expectedError: getSourceError,
		},
		"invalidNarration": {
			url:           narrateErrorUrl,
			expectedError: narrateError,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			jobId, err := jm.Enqueue(ctx, tc.url)
			if err != nil {
				t.Errorf("Failed to enqueue job: %v", err)
				return
			}

			job, err := jm.GetJob(jobId)
			if err != nil {
				t.Errorf("Failed to get job: %v", err)
				return
			}

			time.Sleep(jobInterval)

			result, err := job.GetResult()
			if err != tc.expectedError {
				t.Errorf("Unexpected error: expected %v but got %v", tc.expectedError, err)
				return
			}
			if result != tc.expectedResult {
				t.Errorf("Unmatched result: expected %s but got %s", tc.expectedResult, result)
				return
			}
		})
	}
}
