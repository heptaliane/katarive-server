package job

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	pb "github.com/heptaliane/katarive-server/gen/pb/api/v1"

	"github.com/heptaliane/katarive-server/internal/model"
	"github.com/heptaliane/katarive-server/internal/service/narrator"
	"github.com/heptaliane/katarive-server/internal/service/source"
)

type MutexNarrateJobQueue struct {
	sr source.SourceRegistry
	nr narrator.NarrateRegistry

	jobs   *sync.Map
	logger *slog.Logger
}

func (q *MutexNarrateJobQueue) Queue(
	ctx context.Context,
	opts ...JobOption[narrationJobOption],
) (Job, error) {
	var options narrationJobOption
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
		q.logger.InfoContext(ctx, "Narrate job start", "url", options.url)

		item, err := q.sr.GetItem(options.url)
		if err != nil {
			q.logger.ErrorContext(
				ctx, "Narrate job failed with GetItem",
				"url", options.url,
				"error", err,
			)
			job.set(pb.JobStatus_JOB_STATUS_FAILED, err)
			return
		}
		if item == nil {
			err := q.sr.AddItem(ctx, options.url)
			if err != nil {
				q.logger.ErrorContext(
					ctx, "Narrate job failed with AddItem",
					"url", options.url,
					"error", err,
				)
				job.set(pb.JobStatus_JOB_STATUS_FAILED, err)
				return
			}

			item, err = q.sr.GetItem(options.url)
			if err != nil {
				q.logger.ErrorContext(
					ctx, "Narrate job failed with GetItem",
					"url", options.url,
					"error", err,
				)
				job.set(pb.JobStatus_JOB_STATUS_FAILED, err)
				return
			}
		}

		err = q.nr.Do(
			ctx, item,
			narrator.WithSpeaker(options.speakerId),
			narrator.WithNarrator(options.narrator),
			narrator.WithEncoding(options.encoding),
		)
		if err == nil {
			q.logger.InfoContext(ctx, "Narrate job completed", "url", options.url)
			job.set(pb.JobStatus_JOB_STATUS_COMPLETED, nil)
			q.jobs.Delete(options.url)
			return
		}

		q.logger.ErrorContext(
			ctx, "Narrate job failed",
			"url", options.url,
			"error", err,
		)
		job.set(pb.JobStatus_JOB_STATUS_FAILED, err)
	}()

	return job, nil
}

// Ensure MutexNarrateJobQueue implements NarrateJobQueue
var _ NarrateJobQueue = new(MutexNarrateJobQueue)

// helpers
func (q *MutexNarrateJobQueue) get(url string) (Job, error) {
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
func NewMutexNarrateJobQueue(
	sr source.SourceRegistry,
	nr narrator.NarrateRegistry,
) *MutexNarrateJobQueue {
	return &MutexNarrateJobQueue{
		sr:     sr,
		nr:     nr,
		jobs:   new(sync.Map),
		logger: slog.Default(),
	}
}
