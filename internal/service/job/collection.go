package job

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	pb "github.com/heptaliane/katarive-server/gen/pb/api/v1"

	"github.com/heptaliane/katarive-server/internal/model"
	"github.com/heptaliane/katarive-server/internal/service/source"
)

type MutexSourceCollectionJobQueue struct {
	sr source.SourceRegistry

	jobs   *sync.Map
	logger *slog.Logger
}

func (q *MutexSourceCollectionJobQueue) Queue(
	ctx context.Context,
	opts ...JobOption[sourceCollectionJobOption],
) (Job, error) {
	var options sourceCollectionJobOption
	for _, opt := range opts {
		opt(&options)
	}

	job, err := q.get(options.url)
	if job != nil || err != nil {
		return job, err
	}

	job = NewMutexJob()
	q.jobs.Store(options.url, job)

	go func() {
		q.logger.InfoContext(ctx, "SourceCollection job start", "url", options.url)

		err := q.sr.AddCollection(ctx, options.url)
		if err == nil {
			q.logger.InfoContext(ctx, "SourceCollection job completed", "url", options.url)
			job.set(pb.JobStatus_JOB_STATUS_COMPLETED, nil)
			q.jobs.Delete(options.url)
			return
		}

		q.logger.ErrorContext(
			ctx, "SourceCollection job failed",
			"url", options.url,
			"error", err,
		)
		job.set(pb.JobStatus_JOB_STATUS_FAILED, err)
	}()

	return job, nil
}

// Ensure MutexSourceCollectionJobQueue implements SourceCollectionJobQueue
var _ SourceCollectionJobQueue = new(MutexSourceCollectionJobQueue)

// helpers
func (q *MutexSourceCollectionJobQueue) get(url string) (Job, error) {
	v, ok := q.jobs.Load(url)
	if !ok {
		return nil, nil
	}

	job, ok := v.(Job)
	if !ok {
		q.logger.Error(
			"Unsupported job type",
			"url", url,
			"type", fmt.Sprintf("%T", v),
		)
		return nil, &model.UnexpectedTypeError{Value: v, Expected: new(Job)}
	}

	return job, nil
}
func NewMutexSourceCollectionJobQueue(sr source.SourceRegistry) *MutexSourceCollectionJobQueue {
	return &MutexSourceCollectionJobQueue{
		sr:     sr,
		jobs:   new(sync.Map),
		logger: slog.Default(),
	}
}
