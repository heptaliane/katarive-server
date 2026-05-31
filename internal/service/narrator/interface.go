package narrator

import (
	"context"

	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"

	"github.com/heptaliane/katarive-server/internal/model"
)

//go:generate mockgen -source=$GOFILE -destination=mock/mock_$GOFILE -package=mock
type NarratorManager interface {
	Metadata() *model.NarratorManagerMetadata

	Narrate(ctx context.Context, req *pb.NarrateRequest) (*pb.NarrateResponse, error)
}

//go:generate mockgen -source=$GOFILE -destination=mock/mock_$GOFILE -package=mock
type NarrateRegistry interface {
	Metadata() []*model.NarratorManagerMetadata

	Do(ctx context.Context, source *model.SourceItem, opts ...NarrateOption) error
	Get(source *model.SourceItem, opts ...NarrateOption) *model.NarrateResult
}

// -----------------
// Narrate options
// -----------------

type narrateOption struct {
	encoding  pb.AudioEncoding
	speakerId int32
	narrator  string
}
type NarrateOption = func(*narrateOption)

func WithEncoding(encoding pb.AudioEncoding) NarrateOption {
	return func(opt *narrateOption) {
		opt.encoding = encoding
	}
}
func WithSpeaker(speakerId int32) NarrateOption {
	return func(opt *narrateOption) {
		opt.speakerId = speakerId
	}
}
func WithNarrator(narrator string) NarrateOption {
	return func(opt *narrateOption) {
		opt.narrator = narrator
	}
}
