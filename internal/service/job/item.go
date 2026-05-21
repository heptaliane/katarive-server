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

type PluginSourceItemJobQueue struct {
	sr source.SourceRegistry

	jobs   *sync.Map
	group  *singleflight.Group
	logger *slog.Logger
}

func (q *PluginSourceItemJobQueue) Queue(
	ctx context.Context,
	opts ...JobOption[sourceItemJobOption],
) (string, error) {
	var options sourceItemJobOption
	for _, opt := range opts {
		opt(&options)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}

	jobId := id.String()
	job := NewPluginJob[model.SourceItem]()
	q.jobs.Store(jobId, job)

	go func() {
		q.logger.InfoContext(ctx, "Start sourceItem job", "id", jobId, "url", options.url)

		v, err, _ := q.group.Do(options.url, func() (any, error) {
			opts := []source.SourceOption{
				source.WithoutCache(options.disableCache),
			}

			return q.sr.SourceItem(ctx, options.url, opts...)
		})

		job.mu.Lock()
		defer job.mu.Unlock()

		if err != nil {
			job.err = err
		} else if result, ok := v.(*model.SourceItem); ok {
			q.logger.InfoContext(
				ctx, "SourceItem job completed",
				"id", jobId,
				"url", options.url,
			)
			job.data = result
			job.status = pb.JobStatus_JOB_STATUS_COMPLETED
			return
		} else {
			job.err = &model.UnexpectedTypeError{Value: v, Expected: new(model.SourceItem)}
		}

		q.logger.ErrorContext(
			ctx, "SourceItem job failed",
			"id", jobId,
			"url", options.url,
			"error", job.err,
		)
		job.status = pb.JobStatus_JOB_STATUS_FAILED
	}()

	return jobId, nil
}
func (q *PluginSourceItemJobQueue) Get(id string) (SourceItemJob, error) {
	v, ok := q.jobs.Load(id)
	if !ok {
		q.logger.Warn("No such job", "id", id)
		job := NewPluginJob[model.SourceItem]()
		job.err = &model.JobNotFoundError{Id: id}
		job.status = pb.JobStatus_JOB_STATUS_NOT_FOUND
		return job, nil
	}

	job, ok := v.(SourceItemJob)
	if !ok {
		q.logger.Error(
			"Unsupported job type",
			"id", id,
			"type", fmt.Sprintf("%T", v),
		)
		return nil, &model.UnexpectedTypeError{Value: v, Expected: new(SourceItemJob)}
	}

	return job, nil
}

// Ensure PluginSourceItemJobQueue implements SourceItemJobQueue
var _ SourceItemJobQueue = new(PluginSourceItemJobQueue)

// helpers
func NewSourceItemJobQueue(
	sr source.SourceRegistry,
	group *singleflight.Group,
) *PluginSourceItemJobQueue {
	return &PluginSourceItemJobQueue{
		sr:     sr,
		jobs:   new(sync.Map),
		group:  group,
		logger: slog.Default(),
	}
}
