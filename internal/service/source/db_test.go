package source_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/testing/protocmp"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

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

	cases := map[string]struct {
		itemUrls            []string
		collectionUrls      []string
		nCalls1             int
		nCalls2             int
		disableCache        bool
		expectedItem        []*model.SourceItem
		expectedCollection  []*model.SourceCollection
		expectedItems       [][]*model.SourceSummary
		expectedCollections []*model.SourceCollection
	}{
		"[1]": {
			itemUrls:            []string{SM1_ITEM_URL},
			collectionUrls:      []string{SM1_COLLECTION_URL},
			nCalls1:             1,
			nCalls2:             0,
			expectedItem:        []*model.SourceItem{si1},
			expectedCollection:  []*model.SourceCollection{sc1},
			expectedItems:       [][]*model.SourceSummary{sis1},
			expectedCollections: []*model.SourceCollection{sc1},
		},
		"[1, 2]": {
			itemUrls:            []string{SM1_ITEM_URL, SM2_ITEM_URL},
			collectionUrls:      []string{SM1_COLLECTION_URL, SM2_COLLECTION_URL},
			nCalls1:             1,
			nCalls2:             1,
			expectedItem:        []*model.SourceItem{si1, si2},
			expectedCollection:  []*model.SourceCollection{sc1, sc2},
			expectedItems:       [][]*model.SourceSummary{sis1, sis2},
			expectedCollections: []*model.SourceCollection{sc1, sc2},
		},
		"[1, 1]": {
			itemUrls:            []string{SM1_ITEM_URL, SM1_ITEM_URL},
			collectionUrls:      []string{SM1_COLLECTION_URL, SM1_COLLECTION_URL},
			nCalls1:             1,
			nCalls2:             0,
			expectedItem:        []*model.SourceItem{si1, si1},
			expectedCollection:  []*model.SourceCollection{sc1, sc1},
			expectedItems:       [][]*model.SourceSummary{sis1, sis1},
			expectedCollections: []*model.SourceCollection{sc1},
		},
		"[1, 2, 1]": {
			itemUrls: []string{SM1_ITEM_URL, SM2_ITEM_URL, SM1_ITEM_URL},
			collectionUrls: []string{
				SM1_COLLECTION_URL,
				SM2_COLLECTION_URL,
				SM1_COLLECTION_URL,
			},
			nCalls1:             1,
			nCalls2:             1,
			expectedItem:        []*model.SourceItem{si1, si2, si1},
			expectedCollection:  []*model.SourceCollection{sc1, sc2, sc1},
			expectedItems:       [][]*model.SourceSummary{sis1, sis2, sis1},
			expectedCollections: []*model.SourceCollection{sc1, sc2},
		},
		"[1, 1]; nocache": {
			itemUrls:            []string{SM1_ITEM_URL, SM1_ITEM_URL},
			collectionUrls:      []string{SM1_COLLECTION_URL, SM1_COLLECTION_URL},
			nCalls1:             2,
			nCalls2:             0,
			disableCache:        true,
			expectedItem:        []*model.SourceItem{si1, si1},
			expectedCollection:  []*model.SourceCollection{sc1, sc1},
			expectedItems:       [][]*model.SourceSummary{sis1, sis1},
			expectedCollections: []*model.SourceCollection{sc1},
		},
		"[1, 2, 1]; nocache": {
			itemUrls: []string{SM1_ITEM_URL, SM2_ITEM_URL, SM1_ITEM_URL},
			collectionUrls: []string{
				SM1_COLLECTION_URL,
				SM2_COLLECTION_URL,
				SM1_COLLECTION_URL,
			},
			nCalls1:             2,
			nCalls2:             1,
			disableCache:        true,
			expectedItem:        []*model.SourceItem{si1, si2, si1},
			expectedCollection:  []*model.SourceCollection{sc1, sc2, sc1},
			expectedItems:       [][]*model.SourceSummary{sis1, sis2, sis1},
			expectedCollections: []*model.SourceCollection{sc1, sc2},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			sr := setupDatabaseSourceRegistry(t, tc.nCalls1, tc.nCalls2)

			ctx := context.Background()
			for i := range tc.itemUrls {
				si, err := sr.SourceItem(
					ctx, tc.itemUrls[i], source.WithoutCache(tc.disableCache),
				)
				if err != nil {
					t.Errorf("GetItem failed: %v", err)
					return
				}
				diff := cmp.Diff(tc.expectedItem[i], si, protocmp.Transform())
				if diff != "" {
					t.Errorf("SourceItem mismatch (%d) (-want +got):\n%s", i, diff)
					return
				}
				sc, err := sr.SourceCollection(
					ctx, tc.collectionUrls[i], source.WithoutCache(tc.disableCache),
				)
				if err != nil {
					t.Errorf("GetCollection failed: %v", err)
					return
				}
				diff = cmp.Diff(tc.expectedCollection[i], sc, protocmp.Transform())
				if diff != "" {
					t.Errorf("SourceCollection mismatch (%d) (-want +got):\n%s", i, diff)
					return
				}
				sis, err := sr.SourceItems(ctx, tc.collectionUrls[i])
				if err != nil {
					t.Errorf("GetItems failed: %v", err)
					return
				}
				diff = cmp.Diff(tc.expectedItems[i], sis, protocmp.Transform())
				if diff != "" {
					t.Errorf("SourceItems mismatch (%d) (-want +got):\n%s", i, diff)
					return
				}
			}

			scs, err := sr.SourceCollections()
			if err != nil {
				t.Errorf("GetSourceCollection failed: %v", err)
				return
			}
			diff := cmp.Diff(tc.expectedCollections, scs, protocmp.Transform())
			if diff != "" {
				t.Errorf("SourceCollections mismatch (-want +got):\n%s", diff)
				return
			}
		})
	}
}

// Helper functions
func setupSourceManagers(t *testing.T, nCalls1, nCalls2 int) []source.SourceManager {
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
	sm1.EXPECT().GetSourceItem(gomock.Any(), gomock.Any()).Return(&gsir1, nil).Times(nCalls1)
	sm2.EXPECT().GetSourceItem(gomock.Any(), gomock.Any()).Return(&gsir2, nil).Times(nCalls2)
	sm1.EXPECT().GetSourceCollection(gomock.Any(), gomock.Any()).
		Return(&gscr1, nil).Times(nCalls1)
	sm2.EXPECT().GetSourceCollection(gomock.Any(), gomock.Any()).
		Return(&gscr2, nil).Times(nCalls2)
	return []source.SourceManager{sm1, sm2}
}
func setupDatabaseSourceRegistry(
	t *testing.T,
	nCalls1, nCalls2 int,
) *source.DatabaseSourceRegistry {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect database: %v", err)
	}

	return source.NewDatabaseSourceRegistry(db, setupSourceManagers(t, nCalls1, nCalls2))
}
