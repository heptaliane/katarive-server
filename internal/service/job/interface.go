package job

import (
	"context"

	pb "github.com/heptaliane/katarive-server/gen/pb/api/v1"

	"github.com/heptaliane/katarive-server/internal/model"
)

//go:generate mockgen -source=$GOFILE -destination=mock/mock_$GOFILE -package=mock
type Job[T any] interface {
	Result() *T
	Error() error
	Status() pb.JobStatus
}

//go:generate mockgen -source=$GOFILE -destination=mock/mock_$GOFILE -package=mock
type JobQueue[S, T any] interface {
	Queue(ctx context.Context, opts ...JobOption[S]) (string, error)
	Get(id string) (Job[T], error)
}

type NarrationJob = Job[string]
type SourceItemJob = Job[model.SourceItem]
type SourceItemsJob = Job[[]*model.SourceSummary]
type SourceCollectionJob = Job[model.SourceCollection]
type NarrationJobQueue = JobQueue[narrationJobOption, string]
type SourceItemJobQueue = JobQueue[sourceItemJobOption, model.SourceItem]
type SourceItemsJobQueue = JobQueue[sourceItemsJobOption, []*model.SourceSummary]
type SourceCollectionJobQueue = JobQueue[sourceCollectionJobOption, model.SourceCollection]

// helpers
type JobOption[T any] = func(opt *T)

type narrationJobOption struct {
	url       string
	narrator  string
	speakerId int32
}

func WithNarrationUrl(url string) JobOption[narrationJobOption] {
	return func(opt *narrationJobOption) { opt.url = url }
}
func WithNarrationNarrator(narrator string) JobOption[narrationJobOption] {
	return func(opt *narrationJobOption) { opt.narrator = narrator }
}
func WithNarrationSpeakerId(speakerId int32) JobOption[narrationJobOption] {
	return func(opt *narrationJobOption) { opt.speakerId = speakerId }
}

type sourceItemJobOption struct {
	url string
}

func WithSourceItemUrl(url string) JobOption[sourceItemJobOption] {
	return func(opt *sourceItemJobOption) { opt.url = url }
}

type sourceItemsJobOption struct {
	url string
}

func WithSourceItemsUrl(url string) JobOption[sourceItemsJobOption] {
	return func(opt *sourceItemsJobOption) { opt.url = url }
}

type sourceCollectionJobOption struct {
	url string
}

func WithSourceCollectionUrl(url string) JobOption[sourceCollectionJobOption] {
	return func(opt *sourceCollectionJobOption) { opt.url = url }
}
