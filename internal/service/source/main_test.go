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
var gsir1 pb.GetSourceItemResponse
var gsir2 pb.GetSourceItemResponse
var gscr1 pb.GetSourceCollectionResponse
var gscr2 pb.GetSourceCollectionResponse

const SM1_NAME string = "sm1"
const SM2_NAME string = "sm2"
const SM1_URL string = "http://example.com/1"
const SM2_URL string = "http://example.com/2"

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
	collection1 := "collection1"
	collection2 := "collection2"
	gsir1.Item = &pb.SourceItem{
		Id:           "item1",
		CollectionId: &collection1,
		Url:          SM1_URL,
		Title:        "title1",
		Content:      "content1",
		Language:     pb.Language_LANGUAGE_ENGLISH,
	}
	gsir2.Item = &pb.SourceItem{
		Id:           "item2",
		CollectionId: &collection2,
		Url:          SM2_URL,
		Title:        "title2",
		Content:      "content2",
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
	gscr1.Collection = &pb.SourceCollection{
		Id:          "collection1",
		Url:         "http://example.com/1",
		Title:       "collection-title1",
		Description: "collection-description1",
		Author:      "collction-author1",
		Tags:        []string{"tag1", "tag2"},
	}
	gscr1.Sources = []*pb.SourceSummary{
		{
			Id:    "item1",
			Title: "title1",
			Url:   "http://example.com/1",
		},
	}
	gscr2.Collection = &pb.SourceCollection{
		Id:          "collection2",
		Url:         "http://example.com/2",
		Title:       "collection-title2",
		Description: "collection-description2",
		Author:      "collction-author2",
		Tags:        []string{"tag1", "tag3"},
	}
	gscr2.Sources = []*pb.SourceSummary{
		{
			Id:    "item2",
			Title: "title2",
			Url:   "http://example.com/2",
		},
		{
			Id:    "item3",
			Title: "title3",
			Url:   "http://example.com/3",
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
