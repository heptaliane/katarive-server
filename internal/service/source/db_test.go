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
		urls               []string
		expectedItem       []*model.SourceItem
		expectedCollection []*model.SourceCollection
		expectedItems      [][]*model.SourceSummary
	}{
		"[1]": {
			urls:               []string{SM1_URL},
			expectedItem:       []*model.SourceItem{si1},
			expectedCollection: []*model.SourceCollection{sc1},
			expectedItems:      [][]*model.SourceSummary{sis1},
		},
		"[1, 2]": {
			urls:               []string{SM1_URL, SM2_URL},
			expectedItem:       []*model.SourceItem{si1, si2},
			expectedCollection: []*model.SourceCollection{sc1, sc2},
			expectedItems:      [][]*model.SourceSummary{sis1, sis2},
		},
		"[1, 1]": {
			urls:               []string{SM1_URL, SM1_URL},
			expectedItem:       []*model.SourceItem{si1, si1},
			expectedCollection: []*model.SourceCollection{sc1, sc1},
			expectedItems:      [][]*model.SourceSummary{sis1, sis1},
		},
		"[1, 2, 1]": {
			urls:               []string{SM1_URL, SM2_URL, SM1_URL},
			expectedItem:       []*model.SourceItem{si1, si2, si1},
			expectedCollection: []*model.SourceCollection{sc1, sc2, sc1},
			expectedItems:      [][]*model.SourceSummary{sis1, sis2, sis1},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			sr := setupDatabaseSourceRegistry(t)

			ctx := context.Background()
			for i, url := range tc.urls {
				si, err := sr.SourceItem(ctx, url)
				if err != nil {
					t.Errorf("GetItem failed: %v", err)
					return
				}
				diff := cmp.Diff(tc.expectedItem[i], si, protocmp.Transform())
				if diff != "" {
					t.Errorf("SourceItem mismatch (%d) (-want +got):\n%s", i, diff)
					return
				}
				sc, err := sr.SourceCollection(ctx, url)
				if err != nil {
					t.Errorf("GetCollection failed: %v", err)
					return
				}
				diff = cmp.Diff(tc.expectedCollection[i], sc, protocmp.Transform())
				if diff != "" {
					t.Errorf("SourceCollection mismatch (%d) (-want +got):\n%s", i, diff)
					return
				}
				sis, err := sr.SourceItems(ctx, url)
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
		})
	}
}

// Helper functions
func setupSourceManagers(t *testing.T) []source.SourceManager {
	sm1 := mock.NewMockSourceManager(gomock.NewController(t))
	sm2 := mock.NewMockSourceManager(gomock.NewController(t))
	sm1.EXPECT().Name().Return(SM1_NAME).AnyTimes()
	sm2.EXPECT().Name().Return(SM2_NAME).AnyTimes()
	sm1.EXPECT().IsSupported(SM1_URL).Return(true).AnyTimes()
	sm2.EXPECT().IsSupported(SM2_URL).Return(true).AnyTimes()
	sm1.EXPECT().IsSupported(gomock.Not(SM1_URL)).Return(false).AnyTimes()
	sm2.EXPECT().IsSupported(gomock.Not(SM2_URL)).Return(false).AnyTimes()
	sm1.EXPECT().GetSourceItem(gomock.Any(), gomock.Any()).Return(&gsir1, nil).AnyTimes()
	sm2.EXPECT().GetSourceItem(gomock.Any(), gomock.Any()).Return(&gsir2, nil).AnyTimes()
	sm1.EXPECT().GetSourceCollection(gomock.Any(), gomock.Any()).
		Return(&gscr1, nil).AnyTimes()
	sm2.EXPECT().GetSourceCollection(gomock.Any(), gomock.Any()).
		Return(&gscr2, nil).AnyTimes()
	return []source.SourceManager{sm1, sm2}
}
func setupDatabaseSourceRegistry(t *testing.T) *source.DatabaseSourceRegistry {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect database: %v", err)
	}

	return source.NewDatabaseSourceRegistry(db, setupSourceManagers(t))
}
