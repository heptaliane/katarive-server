package source_test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/testing/protocmp"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/heptaliane/katarive-server/internal/model"
	"github.com/heptaliane/katarive-server/internal/service/source"
	"github.com/heptaliane/katarive-server/internal/service/source/mock"
)

func TestDatabaseSourceRegistry(t *testing.T) {
	t.Parallel()

	si1 := gsir1.GetItem()
	si2 := gsir2.GetItem()
	sc1 := gscr1.GetCollection()
	sc2 := gscr2.GetCollection()
	sis1 := gscr1.GetSources()
	sis2 := gscr2.GetSources()
	ifs1 := []source.GetSourceOption{source.WithItemUrl(SM1_ITEM_URL)}
	ifs2 := []source.GetSourceOption{source.WithItemUrl(SM2_ITEM_URL)}
	cfs1 := []source.GetSourceOption{source.WithCollectionUrl(SM1_COLLECTION_URL)}
	cfs2 := []source.GetSourceOption{source.WithCollectionUrl(SM2_COLLECTION_URL)}

	cases := map[string]struct {
		addItemUrls           []string
		addCollectionUrls     []string
		getItemUrls           []string
		getCollectionUrls     []string
		getItemsFilters       [][]source.GetSourceOption
		getCollectionsFilters [][]source.GetSourceOption
		expectedItem          []*model.SourceItem
		expectedCollection    []*model.SourceCollection
		expectedItems         [][]*model.SourceSummary
		expectedCollections   [][]*model.SourceCollection
	}{
		"Add: [1], Get: [1]": {
			addItemUrls:           []string{SM1_ITEM_URL},
			addCollectionUrls:     []string{SM1_COLLECTION_URL},
			getItemUrls:           []string{SM1_ITEM_URL},
			getCollectionUrls:     []string{SM1_COLLECTION_URL},
			getItemsFilters:       [][]source.GetSourceOption{cfs1},
			getCollectionsFilters: [][]source.GetSourceOption{{}, ifs1, ifs2},
			expectedItem:          []*model.SourceItem{si1},
			expectedCollection:    []*model.SourceCollection{sc1},
			expectedItems:         [][]*model.SourceSummary{sis1},
			expectedCollections:   [][]*model.SourceCollection{{sc1}, {sc1}, nil},
		},
		"Add: [1, 2], Get: [1, 2]": {
			addItemUrls:           []string{SM1_ITEM_URL, SM2_ITEM_URL},
			addCollectionUrls:     []string{SM1_COLLECTION_URL, SM2_COLLECTION_URL},
			getItemUrls:           []string{SM1_ITEM_URL, SM2_ITEM_URL},
			getCollectionUrls:     []string{SM1_COLLECTION_URL, SM2_COLLECTION_URL},
			getItemsFilters:       [][]source.GetSourceOption{cfs1, cfs2},
			getCollectionsFilters: [][]source.GetSourceOption{{}, ifs1, ifs2},
			expectedItem:          []*model.SourceItem{si1, si2},
			expectedCollection:    []*model.SourceCollection{sc1, sc2},
			expectedItems:         [][]*model.SourceSummary{sis1, sis2},
			expectedCollections:   [][]*model.SourceCollection{{sc1, sc2}, {sc1}, {sc2}},
		},
		"Add: [1], Get: [1, 2]": {
			addItemUrls:           []string{SM1_ITEM_URL},
			addCollectionUrls:     []string{SM1_COLLECTION_URL},
			getItemUrls:           []string{SM1_ITEM_URL, SM2_ITEM_URL},
			getCollectionUrls:     []string{SM1_COLLECTION_URL, SM2_COLLECTION_URL},
			getItemsFilters:       [][]source.GetSourceOption{cfs1, cfs2},
			getCollectionsFilters: [][]source.GetSourceOption{{}},
			expectedItem:          []*model.SourceItem{si1, nil},
			expectedCollection:    []*model.SourceCollection{sc1, nil},
			expectedItems:         [][]*model.SourceSummary{sis1, nil},
			expectedCollections:   [][]*model.SourceCollection{{sc1}},
		},
		"Add: [1, 2, 1], Get: [1, 2]": {
			addItemUrls:           []string{SM1_ITEM_URL, SM2_ITEM_URL, SM1_ITEM_URL},
			addCollectionUrls:     []string{SM1_COLLECTION_URL, SM2_COLLECTION_URL, SM1_COLLECTION_URL},
			getItemUrls:           []string{SM1_ITEM_URL, SM2_ITEM_URL},
			getCollectionUrls:     []string{SM1_COLLECTION_URL, SM2_COLLECTION_URL},
			getItemsFilters:       [][]source.GetSourceOption{cfs1, cfs2},
			getCollectionsFilters: [][]source.GetSourceOption{{}},
			expectedItem:          []*model.SourceItem{si1, si2},
			expectedCollection:    []*model.SourceCollection{sc1, sc2},
			expectedItems:         [][]*model.SourceSummary{sis1, sis2},
			expectedCollections:   [][]*model.SourceCollection{{sc1, sc2}},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			sr := setupDatabaseSourceRegistry(t)

			ctx := context.Background()
			for _, url := range tc.addItemUrls {
				err := sr.AddItem(ctx, url)
				if err != nil {
					t.Errorf("AddItem failed: %v", err)
					return
				}
			}
			for _, url := range tc.addCollectionUrls {
				err := sr.AddCollection(ctx, url)
				if err != nil {
					t.Errorf("AddCollection failed: %v", err)
				}
			}

			for i, url := range tc.getItemUrls {
				item, err := sr.GetItem(url)
				if err != nil {
					t.Errorf("GetItem failed (%d): %v", i, err)
					return
				}
				diff := cmp.Diff(tc.expectedItem[i], item, protocmp.Transform())
				if diff != "" {
					t.Errorf("GetItem mismatch (%d) (-want +got):\n%s", i, diff)
					return
				}
			}
			for i, url := range tc.getCollectionUrls {
				collection, err := sr.GetCollection(url)
				if err != nil {
					t.Errorf("GetCollection failed (%d): %v", i, err)
					return
				}
				diff := cmp.Diff(tc.expectedCollection[i], collection, protocmp.Transform())
				if diff != "" {
					t.Errorf("GetCollection mismatch (%d) (-want +got):\n%s", i, diff)
					return
				}
			}

			for i, filters := range tc.getItemsFilters {
				items, err := sr.GetItems(filters...)
				if err != nil {
					t.Errorf("GetItems failed (%d): %v", i, err)
					return
				}
				diff := cmp.Diff(tc.expectedItems[i], items, protocmp.Transform())
				if diff != "" {
					t.Errorf("GetItems mismatch (%d) (-want +got):\n%s", i, diff)
					return
				}
			}
			for i, filters := range tc.getCollectionsFilters {
				collections, err := sr.GetCollections(filters...)
				if err != nil {
					t.Errorf("GetCollections failed (%d): %v", i, err)
					return
				}
				diff := cmp.Diff(tc.expectedCollections[i], collections, protocmp.Transform())
				if diff != "" {
					t.Errorf("GetCollections mismatch (%d) (-want +got):\n%s", i, diff)
					return
				}
			}
		})
	}
}

// Helper functions
func setupSourceManagers(t *testing.T) []source.SourceManager {
	sm1 := mock.NewMockSourceManager(gomock.NewController(t))
	sm2 := mock.NewMockSourceManager(gomock.NewController(t))
	sm1.EXPECT().Name().Return(SM1_NAME).AnyTimes()
	sm2.EXPECT().Name().Return(SM2_NAME).AnyTimes()
	sm1.EXPECT().IsSupportedItem(SM1_ITEM_URL).Return(true).AnyTimes()
	sm2.EXPECT().IsSupportedItem(SM2_ITEM_URL).Return(true).AnyTimes()
	sm1.EXPECT().IsSupportedItem(gomock.Not(SM1_ITEM_URL)).Return(false).AnyTimes()
	sm2.EXPECT().IsSupportedItem(gomock.Not(SM2_ITEM_URL)).Return(false).AnyTimes()
	sm1.EXPECT().IsSupportedCollection(SM1_COLLECTION_URL).Return(true).AnyTimes()
	sm2.EXPECT().IsSupportedCollection(SM2_COLLECTION_URL).Return(true).AnyTimes()
	sm1.EXPECT().IsSupportedCollection(gomock.Not(SM1_COLLECTION_URL)).Return(false).AnyTimes()
	sm2.EXPECT().IsSupportedCollection(gomock.Not(SM2_COLLECTION_URL)).Return(false).AnyTimes()
	sm1.EXPECT().GetSourceItem(gomock.Any(), gomock.Any()).Return(&gsir1, nil).AnyTimes()
	sm2.EXPECT().GetSourceItem(gomock.Any(), gomock.Any()).Return(&gsir2, nil).AnyTimes()
	sm1.EXPECT().GetSourceCollection(gomock.Any(), gomock.Any()).
		Return(&gscr1, nil).AnyTimes()
	sm2.EXPECT().GetSourceCollection(gomock.Any(), gomock.Any()).
		Return(&gscr2, nil).AnyTimes()
	return []source.SourceManager{sm1, sm2}
}
func setupDatabaseSourceRegistry(
	t *testing.T,
) *source.DatabaseSourceRegistry {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "", log.LstdFlags),
			logger.Config{
				LogLevel: logger.Info,
				Colorful: true,
			},
		),
	})
	if err != nil {
		t.Fatalf("Failed to connect database: %v", err)
	}

	return source.NewDatabaseSourceRegistry(db, setupSourceManagers(t))
}
