package service

import (
	"context"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
	"golang.org/x/sync/singleflight"
)

// ==================================
// Interfaces for NarrateJob handlers
// ==================================

//go:generate mockgen -source=$GOFILE -destination=mock/mock_$GOFILE -package=mock
type NarrateJobService interface {
	Enqueue(ctx context.Context, url string) (string, error)
	GetJob(jobId string) (NarrateJob, error)
}

//go:generate mockgen -source=$GOFILE -destination=mock/mock_$GOFILE -package=mock
type NarrateJob interface {
	GetResult() (string, error)
}

// ================================
// NarrateJobService Implementation
// ================================

// -----------------
// NarrateJobManager
// -----------------

type NarrateJobManager struct {
	narrator NarratorRegistry
	source   SourceRegistry
	jobs     *sync.Map
	group    *singleflight.Group
	logger   *slog.Logger
}

func (m *NarrateJobManager) Enqueue(ctx context.Context, url string) (string, error) {
	jobId, err := uuid.NewV7()
	if err != nil {
		return "", err
	}

	job := &SemaphoreNarrateJob{
		id:  jobId.String(),
		url: url,
		mu:  new(sync.RWMutex),
	}
	m.jobs.Store(job.id, job)

	go func() {
		m.logger.InfoContext(ctx, "Start narrate job", "id", jobId, "url", url)
		v, err, _ := m.group.Do(url, func() (any, error) {
			ctx = context.WithoutCancel(ctx)
			src, err := m.source.Get(ctx, url)
			if err != nil {
				return nil, err
			}

			return m.narrator.Do(
				ctx,
				url,
				src.GetContent(),
				WithNarrateLanguage(src.GetLanguage()),
				// TODO: Allow encoding selection
				WithNarrateEncoding(pb.AudioEncoding_AUDIO_ENCODING_MP3),
			)
		})

		job.mu.Lock()
		defer job.mu.Unlock()

		if err != nil {
			job.err = err
		} else if result, ok := v.(string); ok {
			m.logger.InfoContext(
				ctx, "Narrate job successed",
				"id", jobId,
				"url", url,
			)
			job.result = result
		} else {
			job.err = &UnexpectedTypeError{
				Value:    v,
				Expected: new(string),
			}
		}
		m.logger.ErrorContext(
			ctx, "Narrate job failed",
			"id", jobId,
			"url", url,
			"error", job.err,
		)
	}()

	return job.id, nil
}

func (m *NarrateJobManager) GetJob(jobId string) (NarrateJob, error) {
	v, ok := m.jobs.Load(jobId)
	if !ok {
		m.logger.Warn("No such job", "id", jobId)
		return nil, &JobNotFoundError{JobId: jobId}
	}

	result, ok := v.(*SemaphoreNarrateJob)
	if !ok {
		return nil, &UnexpectedTypeError{
			Value:    v,
			Expected: new(SemaphoreNarrateJob),
		}
	}

	return result, nil
}

// Ensure NarrateJobManager implements NarrateJobService
var _ NarrateJobService = new(NarrateJobManager)

// -----------------
// Helper components
// -----------------

func NewNarrateJobManager(
	narrator NarratorRegistry,
	source SourceRegistry,
) *NarrateJobManager {
	return &NarrateJobManager{
		narrator: narrator,
		source:   source,
		jobs:     new(sync.Map),
		group:    new(singleflight.Group),
		logger:   slog.Default(),
	}
}

// =========================
// NarrateJob Implementation
// =========================

// -------------------
// SemaphoreNarrateJob
// -------------------

type SemaphoreNarrateJob struct {
	id     string
	url    string
	result string
	err    error
	mu     *sync.RWMutex
}

func (j *SemaphoreNarrateJob) GetResult() (string, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	return j.result, j.err
}

// Ensure SemaphoreNarrateJob implements NarrateJob
var _ NarrateJob = new(SemaphoreNarrateJob)
