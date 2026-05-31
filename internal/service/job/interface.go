package job

import (
	"context"

	ppb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
	apb "github.com/heptaliane/katarive-server/gen/pb/api/v1"
)

//go:generate mockgen -source=$GOFILE -destination=mock/mock_$GOFILE -package=mock
type Job interface {
	Error() error
	Status() apb.JobStatus
	set(status apb.JobStatus, err error)
}

//go:generate mockgen -source=$GOFILE -destination=mock/mock_$GOFILE -package=mock
type JobQueue[T any] interface {
	Queue(ctx context.Context, opts ...JobOption[T]) (Job, error)
}

type NarrateJobQueue = JobQueue[narrationJobOption]
type SourceCollectionJobQueue = JobQueue[sourceCollectionJobOption]
type SourceItemJobQueue = JobQueue[sourceItemJobOption]

// helpers
type JobOption[T any] = func(opt *T)

type narrationJobOption struct {
	url       string
	narrator  string
	speakerId int32
	encoding  ppb.AudioEncoding
}
type NarrationJobOption = JobOption[narrationJobOption]

func WithNarrationUrl(url string) NarrationJobOption {
	return func(opt *narrationJobOption) { opt.url = url }
}
func WithNarrationNarrator(narrator string) NarrationJobOption {
	return func(opt *narrationJobOption) { opt.narrator = narrator }
}
func WithNarrationSpeakerId(speakerId int32) NarrationJobOption {
	return func(opt *narrationJobOption) { opt.speakerId = speakerId }
}
func WithNarrationEncoding(encoding ppb.AudioEncoding) NarrationJobOption {
	return func(opt *narrationJobOption) { opt.encoding = encoding }
}

type sourceItemJobOption struct {
	url string
}
type SourceItemJobOption = JobOption[sourceItemJobOption]

func WithSourceItemUrl(url string) SourceItemJobOption {
	return func(opt *sourceItemJobOption) { opt.url = url }
}

type sourceCollectionJobOption struct {
	url string
}
type SourceCollectionJobOption = JobOption[sourceCollectionJobOption]

func WithSourceCollectionUrl(url string) SourceCollectionJobOption {
	return func(opt *sourceCollectionJobOption) { opt.url = url }
}
