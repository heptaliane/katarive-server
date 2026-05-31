package job_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/heptaliane/katarive-server/internal/service/job"
)

func TestMutexSourceItemJobQueue(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		url           string
		expectedError error
	}{
		"job_success": {
			url:           VALID_URL,
			expectedError: nil,
		},
		"job_failed": {
			url:           "http://invalid.com",
			expectedError: sie,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			jq := job.NewMutexSourceItemJobQueue(
				setupSourceRegistry(t),
			)

			ctx := context.Background()
			job, err := jq.Queue(ctx, job.WithSourceItemUrl(tc.url))
			if err != nil {
				t.Errorf("Queue returns unexpected error: %v", err)
				return
			}

			if err != nil {
				t.Errorf("Get returns unexpected error: %v", err)
				return
			}

			time.Sleep(interval)

			err = job.Error()
			if tc.expectedError == nil {
				if err != nil {
					t.Errorf("Job returns unexpected error: %v", err)
					return
				}
			} else {
				if diff := cmp.Diff(tc.expectedError.Error(), err.Error()); diff != "" {
					t.Errorf("Job returns unexpected error (-want +got):\n%s", diff)
					return
				}
			}
		})
	}
}
