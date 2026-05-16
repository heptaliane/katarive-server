package job

import (
	"sync"

	pb "github.com/heptaliane/katarive-server/gen/pb/api/v1"

	"github.com/heptaliane/katarive-server/internal/model"
)

type PluginJob[T any] struct {
	data   *T
	status pb.JobStatus
	err    error

	mu *sync.RWMutex
}

func (j *PluginJob[T]) Result() *T {
	j.mu.RLock()
	defer j.mu.RUnlock()

	return j.data
}
func (j *PluginJob[T]) Status() pb.JobStatus {
	j.mu.RLock()
	defer j.mu.RUnlock()

	return j.status
}
func (j *PluginJob[T]) Error() error {
	j.mu.RLock()
	defer j.mu.RUnlock()

	return j.err
}

type PluginNarrationJob = PluginJob[string]
type PluginSourceItemJob = PluginJob[model.SourceItem]
type PluginSourceItemsJob = PluginJob[[]*model.SourceSummary]
type PluginSourceCollectionJob = PluginJob[model.SourceCollection]

// Ensure PluginNarrationJob implements NarrationJob
var _ NarrationJob = new(PluginNarrationJob)

// Ensure PluginSourceItemJob implements SourceItemJob
var _ SourceItemJob = new(PluginSourceItemJob)

// Ensure PluginSourceItemsJob implements SourceItemsJob
var _ SourceItemsJob = new(PluginSourceItemsJob)

// Ensure PluginCollectionJob implements CollectionJob
var _ SourceCollectionJob = new(PluginSourceCollectionJob)

// helpers
func NewPluginJob[T any]() *PluginJob[T] {
	return &PluginJob[T]{
		status: pb.JobStatus_JOB_STATUS_PROGRESSING,
		mu:     new(sync.RWMutex),
	}
}
