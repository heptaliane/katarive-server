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
	"github.com/heptaliane/katarive-server/internal/service/narrator"
	"github.com/heptaliane/katarive-server/internal/service/source"
)

type PluginNarrationJobQueue struct {
	sr source.SourceRegistry
	nr narrator.NarrateRegistry

	jobs   *sync.Map
	group  *singleflight.Group
	logger *slog.Logger
}

func (q *PluginNarrationJobQueue) Queue(
	ctx context.Context,
	opts ...JobOption[narrationJobOption],
) (string, error) {
	var options narrationJobOption
	for _, opt := range opts {
		opt(&options)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}

	jobId := id.String()
	job := NewPluginJob[string]()
	q.jobs.Store(jobId, job)

	go func() {
		q.logger.InfoContext(ctx, "Start narration job", "id", jobId, "url", options.url)

		v, err, _ := q.group.Do(options.url, func() (any, error) {
			src, err := q.sr.SourceItem(ctx, options.url)
			if err != nil {
				return nil, err
			}

			return q.nr.Do(
				ctx, src,
				narrator.WithSpeaker(options.speakerId),
				narrator.WithNarrator(options.narrator),
				narrator.WithEncoding(options.encoding),
			)
		})

		job.mu.Lock()
		defer job.mu.Unlock()

		if err != nil {
			job.err = err
		} else if result, ok := v.(string); ok {
			q.logger.InfoContext(
				ctx, "Narration job completed",
				"id", jobId,
				"url", options.url,
				"result", result,
			)
			job.data = &result
			job.status = pb.JobStatus_JOB_STATUS_COMPLETED
			return
		} else {
			job.err = &model.UnexpectedTypeError{Value: v, Expected: new(string)}
		}

		q.logger.ErrorContext(
			ctx, "Narration job failed",
			"id", jobId,
			"url", options.url,
			"error", job.err,
		)
		job.status = pb.JobStatus_JOB_STATUS_FAILED
	}()

	return jobId, nil
}
func (q *PluginNarrationJobQueue) Get(id string) (NarrationJob, error) {
	v, ok := q.jobs.Load(id)
	if !ok {
		q.logger.Warn("No such job", "id", id)
		job := NewPluginJob[string]()
		job.err = &model.JobNotFoundError{Id: id}
		job.status = pb.JobStatus_JOB_STATUS_NOT_FOUND
		return job, nil
	}

	job, ok := v.(NarrationJob)
	if !ok {
		q.logger.Error(
			"Unsupported job type",
			"id", id,
			"type", fmt.Sprintf("%T", v),
		)
		return nil, &model.UnexpectedTypeError{Value: v, Expected: new(NarrationJob)}
	}

	return job, nil
}

// Ensure PluginNarrationJobQueue implements NarrationJobQueue
var _ NarrationJobQueue = new(PluginNarrationJobQueue)

// helpers
func NewNarrationJobQueue(
	sr source.SourceRegistry,
	nr narrator.NarrateRegistry,
	group *singleflight.Group,
) *PluginNarrationJobQueue {
	return &PluginNarrationJobQueue{
		sr:     sr,
		nr:     nr,
		jobs:   new(sync.Map),
		group:  group,
		logger: slog.Default(),
	}
}
