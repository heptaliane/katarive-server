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
	SourceId(ctx context.Context, url string) string
	SourceItem(ctx context.Context, itemId string) *model.SourceItem
	SourceCollection(ctx context.Context, collectionId string) *model.SourceCollection
	SourceItems(ctx context.Context, collectionId string) []*model.SourceSummary
}
