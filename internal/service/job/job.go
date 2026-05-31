package job

import (
	"sync"

	pb "github.com/heptaliane/katarive-server/gen/pb/api/v1"
)

type MutexJob struct {
	status pb.JobStatus
	err    error

	mu *sync.RWMutex
}

func (j *MutexJob) Status() pb.JobStatus {
	j.mu.RLock()
	defer j.mu.RUnlock()

	return j.status
}
func (j *MutexJob) Error() error {
	j.mu.RLock()
	defer j.mu.RUnlock()

	return j.err
}
func (j *MutexJob) set(status pb.JobStatus, err error) {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.status = status
	j.err = err
}

// Ensure MutexJob implements Job
var _ Job = new(MutexJob)

// helpers
func NewMutexJob() *MutexJob {
	return &MutexJob{
		status: pb.JobStatus_JOB_STATUS_PROGRESSING,
		mu:     new(sync.RWMutex),
	}
}
