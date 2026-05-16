package job_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/heptaliane/katarive-server/internal/model"
	"github.com/heptaliane/katarive-server/internal/service/narrator"
	nmock "github.com/heptaliane/katarive-server/internal/service/narrator/mock"
	"github.com/heptaliane/katarive-server/internal/service/source"
	smock "github.com/heptaliane/katarive-server/internal/service/source/mock"
)

var sc *model.SourceCollection
var si *model.SourceItem
var sis []*model.SourceSummary
var sce error
var sie error
var sise error
var ne error
var interval time.Duration

const VALID_URL string = "http://valid.com"
const VALID_PATH string = "item-id_item-title"

func TestMain(m *testing.M) {
	setupSourceItem()
	setupSourceCollection()
	setupSourceItems()
	setupError()
	setupInterval()

	code := m.Run()

	os.Exit(code)
}

func setupSourceItem() {
	si = &model.SourceItem{
		Id:    "item-id",
		Title: "item-title",
		Url:   VALID_URL,
	}
}
func setupSourceCollection() {
	sc = &model.SourceCollection{
		Id:    "colelction-id",
		Title: "collection-title",
	}
}
func setupSourceItems() {
	sis = []*model.SourceSummary{
		{Id: "item-1", Title: "title-1"},
		{Id: "item-2", Title: "title-2"},
	}
}
func setupError() {
	sce = errors.New("SourceRegistry.SourceCollection failed")
	sie = errors.New("SourceRegistry.SourceItem failed")
	sise = errors.New("SourceRegistry.SourceItems failed")
	ne = errors.New("NarrateRegistry.Do failed")
}
func setupInterval() {
	interval, _ = time.ParseDuration("10ms")
}
func setupSourceRegistry(t *testing.T) source.SourceRegistry {
	sr := smock.NewMockSourceRegistry(gomock.NewController(t))
	sr.EXPECT().SourceItem(gomock.Any(), VALID_URL).Return(si, nil).AnyTimes()
	sr.EXPECT().SourceItem(gomock.Any(), gomock.Not(VALID_URL)).Return(nil, sie).AnyTimes()
	sr.EXPECT().SourceCollection(gomock.Any(), VALID_URL).Return(sc, nil).AnyTimes()
	sr.EXPECT().SourceCollection(gomock.Any(), gomock.Not(VALID_URL)).Return(nil, sce).AnyTimes()
	sr.EXPECT().SourceItems(gomock.Any(), VALID_URL).Return(sis, nil).AnyTimes()
	sr.EXPECT().SourceItems(gomock.Any(), gomock.Not(VALID_URL)).Return(nil, sise).AnyTimes()
	return sr
}
func setupNarrateRegistry(t *testing.T) narrator.NarrateRegistry {
	nr := nmock.NewMockNarrateRegistry(gomock.NewController(t))
	nr.EXPECT().Do(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(
			ctx context.Context,
			item *model.SourceItem,
			opts ...narrator.NarrateOption,
		) (string, error) {
			if item.GetUrl() == VALID_URL {
				return VALID_PATH, nil
			}
			return "", ne
		}).AnyTimes()
	return nr
}
