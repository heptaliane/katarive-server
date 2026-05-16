package handler_test

import (
	"errors"
	"os"
	"testing"
	"time"

	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
	"go.uber.org/mock/gomock"

	"github.com/heptaliane/katarive-server/internal/handler"
	"github.com/heptaliane/katarive-server/internal/model"
	"github.com/heptaliane/katarive-server/internal/service/narrator"
	nmock "github.com/heptaliane/katarive-server/internal/service/narrator/mock"
	"github.com/heptaliane/katarive-server/internal/service/source"
	smock "github.com/heptaliane/katarive-server/internal/service/source/mock"
)

var si *model.SourceItem
var sc *model.SourceCollection
var sis []*model.SourceSummary
var nmms []*model.NarratorManagerMetadata
var sie error
var sce error
var sise error
var ne error
var interval time.Duration

const VALID_URL string = "http://valid.com"
const VALID_PATH string = "/path/to/valid"

func TestMain(m *testing.M) {
	setupSourceItem()
	setupSourceCollection()
	setupSourceItems()
	setupNarrateManagerMetadata()
	setupError()
	setupInterval()

	code := m.Run()
	os.Exit(code)
}

func setupError() {
	sce = errors.New("SourceRegistry.SourceCollection failed")
	sie = errors.New("SourceRegistry.SourceItem failed")
	sise = errors.New("SourceRegistry.SourceItems failed")
	ne = errors.New("NarrateRegistry.Do failed")
}
func setupSourceItem() {
	si = &model.SourceItem{
		Id:      "item-id",
		Url:     VALID_URL,
		Title:   "item-title",
		Content: "item-content",
	}
}
func setupSourceCollection() {
	sc = &model.SourceCollection{
		Id:          "collection-id",
		Url:         VALID_URL,
		Title:       "collection-title",
		Description: "collection-description",
		Author:      "collection-author",
		Tags:        []string{"tag1", "tag2"},
	}
}
func setupSourceItems() {
	sis = []*model.SourceSummary{
		{Id: "item1", Title: "title1", Url: "http://valid.com/1"},
		{Id: "item2", Title: "title2", Url: "http://valid.com/2"},
	}
}
func setupNarrateManagerMetadata() {
	nmms = []*model.NarratorManagerMetadata{
		{
			Name: "narrator1",
			Encodings: []pb.AudioEncoding{
				pb.AudioEncoding_AUDIO_ENCODING_WAV,
			},
			Speakers: []*pb.SpeakerInfo{
				{Id: 1, Name: "narrator1-name1"},
			},
		},
		{
			Name: "narrator2",
			Encodings: []pb.AudioEncoding{
				pb.AudioEncoding_AUDIO_ENCODING_MP3,
				pb.AudioEncoding_AUDIO_ENCODING_M4A,
			},
			Speakers: []*pb.SpeakerInfo{
				{Id: 1, Name: "narrator2-name1"},
				{Id: 2, Name: "narrator2-name2"},
			},
		},
	}
}
func setupInterval() {
	interval, _ = time.ParseDuration("10ms")
}
func newSourceRegistry(t *testing.T) source.SourceRegistry {
	sr := smock.NewMockSourceRegistry(gomock.NewController(t))
	sr.EXPECT().SourceItem(gomock.Any(), VALID_URL).Return(si, nil).AnyTimes()
	sr.EXPECT().SourceItem(gomock.Any(), gomock.Not(VALID_URL)).Return(nil, sie).AnyTimes()
	sr.EXPECT().SourceCollection(gomock.Any(), VALID_URL).Return(sc, nil).AnyTimes()
	sr.EXPECT().SourceCollection(gomock.Any(), gomock.Not(VALID_URL)).Return(nil, sce).AnyTimes()
	sr.EXPECT().SourceItems(gomock.Any(), VALID_URL).Return(sis, nil).AnyTimes()
	sr.EXPECT().SourceItems(gomock.Any(), gomock.Not(VALID_URL)).Return(nil, sise).AnyTimes()
	return sr
}
func newNarrateRegistry(t *testing.T) narrator.NarrateRegistry {
	nr := nmock.NewMockNarrateRegistry(gomock.NewController(t))
	nr.EXPECT().Metadata().Return(nmms).AnyTimes()
	nr.EXPECT().Do(gomock.Any(), gomock.Any(), gomock.Any()).Return(VALID_PATH, nil).AnyTimes()
	return nr
}
func newKatariveHandlerV1(t *testing.T) *handler.KatariveHandlerV1 {
	sr := newSourceRegistry(t)
	nr := newNarrateRegistry(t)
	pm := handler.NewBasePathModifier()
	return handler.NewKatariveHandlerV1(sr, nr, pm)
}
