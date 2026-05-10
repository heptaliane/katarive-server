package source_test

import (
	"os"
	"testing"

	pbmock "github.com/heptaliane/katarive-go-sdk/gen/mock/plugin/v1"
	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
	"go.uber.org/mock/gomock"
)

var gssmr pb.GetSourceServiceMetadataResponse
var gsir pb.GetSourceItemResponse
var gscr pb.GetSourceCollectionResponse

func TestMain(m *testing.M) {
	setupGetSourceServiceMetadataResponse()
	setupGetSourceItemResponse()
	setupGetSourceCollectionResponse()

	code := m.Run()

	os.Exit(code)
}

// Setup functions
func setupGetSourceServiceMetadataResponse() {
	gssmr.Name = "example-name"
	gssmr.Version = "v1"
	gssmr.SupportedPattern = `^http://example\.com/.*$`
}
func setupGetSourceItemResponse() {
	collectionId := "collection-id"
	gsir.Item = &pb.SourceItem{
		Id:           "item-id",
		CollectionId: &collectionId,
		Url:          "http://example.com/001",
		Title:        "title",
		Content:      "content",
		Language:     pb.Language_LANGUAGE_ENGLISH,
	}
}
func setupGetSourceCollectionResponse() {
	gscr.Collection = &pb.SourceCollection{
		Id:          "collection-id",
		Url:         "http://example.com",
		Title:       "collection-title",
		Description: "collection-description",
		Author:      "collction-author",
		Tags:        []string{"tag1", "tag2"},
	}
	gscr.Sources = []*pb.SourceSummary{
		{
			Id:    "item-id-1",
			Title: "item-title-1",
			Url:   "http://example.com/1",
		},
		{
			Id:    "item-id-2",
			Title: "item-title-2",
			Url:   "http://example.com/2",
		},
	}
}
func setupSourceServiceClient(t *testing.T) pb.SourceServiceClient {
	t.Helper()

	client := pbmock.NewMockSourceServiceClient(gomock.NewController(t))

	client.EXPECT().GetSourceServiceMetadata(gomock.Any(), gomock.Any()).
		Return(&gssmr, nil).AnyTimes()
	client.EXPECT().GetSourceItem(gomock.Any(), gomock.Any()).
		Return(&gsir, nil).AnyTimes()
	client.EXPECT().GetSourceCollection(gomock.Any(), gomock.Any()).
		Return(&gscr, nil).AnyTimes()

	return client
}
