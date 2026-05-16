package job_test

import (
	"errors"
	"os"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/heptaliane/katarive-server/internal/model"
	"github.com/heptaliane/katarive-server/internal/service/source"
	"github.com/heptaliane/katarive-server/internal/service/source/mock"
)

var sc *model.SourceCollection
var si *model.SourceItem
var sie error
var interval time.Duration

const VALID_URL string = "http://valid.com"

func TestMain(m *testing.M) {
	setupSourceItem()
	setupError()
	setupInterval()

	code := m.Run()

	os.Exit(code)
}

func setupSourceItem() {
	si = &model.SourceItem{
		Id:    "item-id",
		Title: "item-title",
	}
}
func setupSourceCollection() {
	sc = &model.SourceCollection{
		Id:    "colelction-id",
		Title: "collection-title",
	}
}
func setupError() {
	sie = errors.New("SourceRegistry.SourceItem failed")
}
func setupInterval() {
	interval, _ = time.ParseDuration("10ms")
}
func setupSourceRegistry(t *testing.T) source.SourceRegistry {
	sr := mock.NewMockSourceRegistry(gomock.NewController(t))
	sr.EXPECT().SourceItem(gomock.Any(), VALID_URL).Return(si, nil).AnyTimes()
	sr.EXPECT().SourceItem(gomock.Any(), gomock.Not(VALID_URL)).Return(nil, sie).AnyTimes()
	sr.EXPECT().SourceCollection(gomock.Any(), VALID_URL).Return(sc, nil).AnyTimes()
	sr.EXPECT().SourceCollection(gomock.Any(), gomock.Not(VALID_URL)).Return(nil, sie).AnyTimes()
	return sr
}
