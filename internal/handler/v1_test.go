package handler_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.uber.org/mock/gomock"

	pb "github.com/heptaliane/katarive-server/gen/pb/api/v1"
	"github.com/heptaliane/katarive-server/internal/handler"
	"github.com/heptaliane/katarive-server/internal/service"
	"github.com/heptaliane/katarive-server/internal/service/mock"
)

func TestKatariveHandlerV1CreateNarration(t *testing.T) {
	t.Parallel()

	js := mock.NewMockNarrateJobService(gomock.NewController(t))
	pm := handler.NewBasePathModifier()
	kh := handler.NewKatariveHandler(js, pm)

	validUrl := "http://valid.com"
	invalidUrl := "http://invalid.com"
	validJobId := "jobId"
	invalidError := errors.New("invalid job")
	js.EXPECT().Enqueue(gomock.Any(), validUrl, gomock.Any(), gomock.Any()).
		Return(validJobId, nil).AnyTimes()
	js.EXPECT().Enqueue(gomock.Any(), invalidUrl, gomock.Any(), gomock.Any()).
		Return("", invalidError).AnyTimes()

	cases := map[string]struct {
		url              string
		expectedResponse *pb.CreateNarrationResponse
		expectedError    error
	}{
		"valid": {
			url:              validUrl,
			expectedResponse: &pb.CreateNarrationResponse{Id: validJobId},
		},
		"invalid": {
			url:           invalidUrl,
			expectedError: invalidError,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			req := &pb.CreateNarrationRequest{Url: tc.url}
			res, err := kh.CreateNarration(ctx, req)
			if err != tc.expectedError {
				t.Errorf("Unmatched error: expected %v but got %v", err, tc.expectedError)
				return
			}
			opt := cmpopts.IgnoreUnexported(pb.CreateNarrationResponse{})
			if diff := cmp.Diff(tc.expectedResponse, res, opt); diff != "" {
				t.Errorf("Unmatched CrateNarrationResponse (got: -, want: +):\n%s", diff)
				return
			}
		})
	}
}

func TestKatariveHandlerV1GetJobStatus(t *testing.T) {
	t.Parallel()

	completedJob := mock.NewMockNarrateJob(gomock.NewController(t))
	failedJob := mock.NewMockNarrateJob(gomock.NewController(t))
	progressJob := mock.NewMockNarrateJob(gomock.NewController(t))
	js := mock.NewMockNarrateJobService(gomock.NewController(t))
	pm := handler.NewBasePathModifier()
	kh := handler.NewKatariveHandler(js, pm)

	validJobId := "valid"
	notFoundJobId := "not_found"
	getJobFailedJobId := "job_failed"
	getResultFailedJobId := "result_failed"
	progressJobId := "progress"
	notFoundError := &service.JobNotFoundError{JobId: notFoundJobId}
	validPath := "/path/to/file"

	js.EXPECT().GetJob(validJobId).Return(completedJob, nil).AnyTimes()
	js.EXPECT().GetJob(notFoundJobId).Return(nil, notFoundError).AnyTimes()
	js.EXPECT().GetJob(getJobFailedJobId).Return(nil, errors.New("some error")).AnyTimes()
	js.EXPECT().GetJob(getResultFailedJobId).Return(failedJob, nil).AnyTimes()
	js.EXPECT().GetJob(progressJobId).Return(progressJob, nil).AnyTimes()
	completedJob.EXPECT().GetResult().Return(validPath, nil).Times(1)
	failedJob.EXPECT().GetResult().Return("", errors.New("some error")).Times(1)
	progressJob.EXPECT().GetResult().Return("", nil).Times(1)

	cases := map[string]struct {
		jobId            string
		expectedResponse *pb.GetJobStatusResponse
	}{
		"valid": {
			jobId: validJobId,
			expectedResponse: &pb.GetJobStatusResponse{
				Status: pb.GetJobStatusResponse_STATUS_COMPLETED,
				Path:   &validPath,
			},
		},
		"notFound": {
			jobId: notFoundJobId,
			expectedResponse: &pb.GetJobStatusResponse{
				Status: pb.GetJobStatusResponse_STATUS_NOT_FOUND,
			},
		},
		"getJob-failed": {
			jobId: getJobFailedJobId,
			expectedResponse: &pb.GetJobStatusResponse{
				Status: pb.GetJobStatusResponse_STATUS_FAILED,
			},
		},
		"getResult-failed": {
			jobId: getResultFailedJobId,
			expectedResponse: &pb.GetJobStatusResponse{
				Status: pb.GetJobStatusResponse_STATUS_FAILED,
			},
		},
		"progress": {
			jobId: progressJobId,
			expectedResponse: &pb.GetJobStatusResponse{
				Status: pb.GetJobStatusResponse_STATUS_PROGRESSING,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			req := &pb.GetJobStatusRequest{Id: tc.jobId}
			res, err := kh.GetJobStatus(ctx, req)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			opt := cmpopts.IgnoreUnexported(pb.GetJobStatusResponse{})
			if diff := cmp.Diff(res, tc.expectedResponse, opt); diff != "" {
				t.Errorf("Unmatched GetJobStatusResponse (got: -, want: +):\n%s", diff)
				return
			}
		})
	}
}
