package job

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	pb "github.com/heptaliane/katarive-server/gen/pb/api/v1"
	"golang.org/x/sync/singleflight"

	"github.com/heptaliane/katarive-server/internal/model"
	"github.com/heptaliane/katarive-server/internal/service/source"
)

type PluginSourceItemsJobQueue struct {
	sr source.SourceRegistry

	jobs   *sync.Map
	group  *singleflight.Group
	logger *slog.Logger
}

func (q *PluginSourceItemsJobQueue) Queue(
	ctx context.Context,
	opts ...JobOption[sourceItemsJobOption],
) (string, error) {
	var options sourceItemsJobOption
	for _, opt := range opts {
		opt(&options)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}

	jobId := id.String()
	job := NewPluginJob[[]*model.SourceSummary]()
	q.jobs.Store(jobId, job)

	go func() {
		q.logger.InfoContext(ctx, "Start sourceItems job", "id", jobId, "url", options.url)

		v, err, _ := q.group.Do(options.url, func() (any, error) {
			return q.sr.SourceItems(ctx, options.url)
		})

		job.mu.Lock()
		defer job.mu.Unlock()

		if err != nil {
			job.err = err
		} else if result, ok := v.([]*model.SourceSummary); ok {
			q.logger.InfoContext(
				ctx, "SourceItems job completed",
				"id", jobId,
				"url", options.url,
			)
			job.data = &result
			job.status = pb.JobStatus_JOB_STATUS_COMPLETED
			return
		} else {
			job.err = &model.UnexpectedTypeError{Value: v, Expected: new([]*model.SourceSummary)}
		}

		q.logger.ErrorContext(
			ctx, "SourceItems job failed",
			"id", jobId,
			"url", options.url,
			"error", job.err,
		)
		job.status = pb.JobStatus_JOB_STATUS_FAILED
	}()

	return jobId, nil
}
func (q *PluginSourceItemsJobQueue) Get(id string) (SourceItemsJob, error) {
	v, ok := q.jobs.Load(id)
	if !ok {
		q.logger.Warn("No such job", "id", id)
		job := NewPluginJob[[]*model.SourceSummary]()
		job.err = &model.JobNotFoundError{Id: id}
		job.status = pb.JobStatus_JOB_STATUS_NOT_FOUND
		return job, nil
	}

	job, ok := v.(SourceItemsJob)
	if !ok {
		q.logger.Error(
			"Unsupported job type",
			"id", id,
			"type", fmt.Sprintf("%T", v),
		)
		return nil, &model.UnexpectedTypeError{Value: v, Expected: new(SourceItemsJob)}
	}

	return job, nil
}

// Ensure PluginSourceItemsJobQueue implements SourceItemsJobQueue
var _ SourceItemsJobQueue = new(PluginSourceItemsJobQueue)

// helpers
func NewSourceItemsJobQueue(
	sr source.SourceRegistry,
	group *singleflight.Group,
) *PluginSourceItemsJobQueue {
	return &PluginSourceItemsJobQueue{
		sr:     sr,
		jobs:   new(sync.Map),
		group:  group,
		logger: slog.Default(),
	}
}
