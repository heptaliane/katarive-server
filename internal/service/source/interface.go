package source

import (
	"context"

	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
	"github.com/heptaliane/katarive-server/internal/model"
)

//go:generate mockgen -source=$GOFILE -destination=mock/mock_$GOFILE -package=mock
type SourceManager interface {
	Name() string
	IsSupported(url string) bool

	GetSourceItem(
		ctx context.Context,
		req *pb.GetSourceItemRequest,
	) (*pb.GetSourceItemResponse, error)
	GetSourceCollection(
		ctx context.Context,
		req *pb.GetSourceCollectionRequest,
	) (*pb.GetSourceCollectionResponse, error)
}

//go:generate mockgen -source=$GOFILE -destination=mock/mock_$GOFILE -package=mock
type SourceRegistry interface {
	SourceItem(ctx context.Context, url string, opts ...SourceOption) (*model.SourceItem, error)
	SourceItems(ctx context.Context, url string, opts ...SourceOption) ([]*model.SourceSummary, error)
	SourceCollection(
		ctx context.Context,
		url string,
		opts ...SourceOption,
	) (*model.SourceCollection, error)
	SourceCollections() ([]*model.SourceCollection, error)
}

type sourceOptions struct {
	disableCache bool
}
type SourceOption = func(opt *sourceOptions)

func WithoutCache(disableCache bool) SourceOption {
	return func(opt *sourceOptions) { opt.disableCache = disableCache }
}
