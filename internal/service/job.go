package service

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/sync/singleflight"
)

type NarrateJobService interface {
	Enqueue(ctx context.Context, url string) (string, error)
	GetJob(jobId string) (*NarrateJob, error)
}

type NarrateJob struct {
	id     string
	url    string
	result string
	err    error
	mu     *sync.RWMutex
}

func (j *NarrateJob) GetResult() (string, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	return j.result, j.err
}

type NarrateJobManager struct {
	narrator *NarratorRegistry
	source   *SourceRegistry
	jobs     *sync.Map
	group    *singleflight.Group
}

func (m *NarrateJobManager) Enqueue(ctx context.Context, url string) (string, error) {
	jobId, err := uuid.NewV7()
	if err != nil {
		return "", err
	}

	job := &NarrateJob{
		id:  jobId.String(),
		url: url,
	}
	m.jobs.Store(jobId, job)

	go func() {
		v, err, _ := m.group.Do(url, func() (any, error) {
			src, err := m.source.GetSource(ctx, url)
			if err != nil {
				return nil, err
			}

			return m.narrator.GetNarration(
				ctx,
				url,
				src.GetContent(),
				WithNarrateLanguage(src.GetLanguage()),
			)
		})

		job.mu.Lock()
		defer job.mu.Unlock()

		if err != nil {
			job.err = err
			return
		}
		if result, ok := v.(string); ok {
			job.result = result
		} else {
			job.err = &UnexpectedTypeError{
				Value:    v,
				Expected: new(string),
			}
		}
	}()

	return job.id, nil
}

func (n *NarrateJobManager) GetJob(jobId string) (*NarrateJob, error) {
	v, ok := n.jobs.Load(jobId)
	if !ok {
		return nil, &JobNotFoundError{JobId: jobId}
	}

	result, ok := v.(*NarrateJob)
	if !ok {
		return nil, &UnexpectedTypeError{
			Value:    v,
			Expected: new(NarrateJob),
		}
	}

	return result, nil
}

var _ NarrateJobService = new(NarrateJobManager)
