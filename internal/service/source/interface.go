package source

import (
	"context"

	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
	"github.com/heptaliane/katarive-server/internal/model"
)

//go:generate mockgen -source=$GOFILE -destination=mock/mock_$GOFILE -package=mock
type SourceManager interface {
	Name() string
	IsSupportedItem(url string) bool
	IsSupportedCollection(url string) bool

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
	AddItem(ctx context.Context, itemUrl string) error
	AddCollection(ctx context.Context, collectionUrl string) error

	GetItem(itemUrl string) (*model.SourceItem, error)
	GetItems(opts ...GetSourceOption) ([]*model.SourceSummary, error)
	GetCollection(collectionUrl string) (*model.SourceCollection, error)
	GetCollections(opts ...GetSourceOption) ([]*model.SourceCollection, error)
}

type getSourceOptions struct {
	itemUrl       string
	collectionUrl string
}
type GetSourceOption = func(opt *getSourceOptions)

func WithItemUrl(itemUrl string) GetSourceOption {
	return func(opt *getSourceOptions) { opt.itemUrl = itemUrl }
}
func WithCollectionUrl(collectionUrl string) GetSourceOption {
	return func(opt *getSourceOptions) { opt.collectionUrl = collectionUrl }
}
