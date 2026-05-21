package job

import (
	"context"

	ppb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
	apb "github.com/heptaliane/katarive-server/gen/pb/api/v1"

	"github.com/heptaliane/katarive-server/internal/model"
)

//go:generate mockgen -source=$GOFILE -destination=mock/mock_$GOFILE -package=mock
type Job[T any] interface {
	Result() *T
	Error() error
	Status() apb.JobStatus
}

//go:generate mockgen -source=$GOFILE -destination=mock/mock_$GOFILE -package=mock
type JobQueue[S, T any] interface {
	Queue(ctx context.Context, opts ...JobOption[S]) (string, error)
	Get(id string) (Job[T], error)
}

type NarrationJob = Job[string]
type SourceItemJob = Job[model.SourceItem]
type SourceCollectionJob = Job[model.SourceCollectionPackage]
type NarrationJobQueue = JobQueue[narrationJobOption, string]
type SourceItemJobQueue = JobQueue[sourceItemJobOption, model.SourceItem]
type SourceCollectionJobQueue = JobQueue[sourceCollectionJobOption, model.SourceCollectionPackage]

// helpers
type JobOption[T any] = func(opt *T)

type narrationJobOption struct {
	url          string
	narrator     string
	speakerId    int32
	encoding     ppb.AudioEncoding
	disableCache bool
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
func WithNarrationEncoding(encoding ppb.AudioEncoding) JobOption[narrationJobOption] {
	return func(opt *narrationJobOption) { opt.encoding = encoding }
}
func WithoutNarrationCache(disableCache bool) JobOption[narrationJobOption] {
	return func(opt *narrationJobOption) { opt.disableCache = disableCache }
}

type sourceItemJobOption struct {
	url          string
	disableCache bool
}

func WithSourceItemUrl(url string) JobOption[sourceItemJobOption] {
	return func(opt *sourceItemJobOption) { opt.url = url }
}
func WithoutSourceItemCache(disableCache bool) JobOption[sourceItemJobOption] {
	return func(opt *sourceItemJobOption) { opt.disableCache = disableCache }
}

type sourceCollectionJobOption struct {
	url          string
	disableCache bool
}

func WithSourceCollectionUrl(url string) JobOption[sourceCollectionJobOption] {
	return func(opt *sourceCollectionJobOption) { opt.url = url }
}
func WithoutSourceCollectionCache(disableCache bool) JobOption[sourceCollectionJobOption] {
	return func(opt *sourceCollectionJobOption) { opt.disableCache = disableCache }
}
