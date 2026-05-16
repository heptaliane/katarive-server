package narrator_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"

	"github.com/heptaliane/katarive-server/internal/model"
	"github.com/heptaliane/katarive-server/internal/service/narrator"
)

func TestFileNarratorRegistryDo(t *testing.T) {
	t.Parallel()

	basedir := os.TempDir()

	cases := map[string]struct {
		source   *model.SourceItem
		options  []narrator.NarrateOption
		expected string
	}{
		"normal": {
			source: &model.SourceItem{
				Id:      "id",
				Title:   "title",
				Content: "content",
			},
			options: []narrator.NarrateOption{
				narrator.WithEncoding(pb.AudioEncoding_AUDIO_ENCODING_MP3),
				narrator.WithNarrator(metadata.Name),
				narrator.WithSpeaker(1),
			},
			expected: filepath.Join(basedir, "narrator.v1", "001", "id_title.mp3"),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			nms := []narrator.NarratorManager{setupNarratorManager(t)}
			nr := narrator.NewFileNarratorRegistry(ctx, basedir, nms)

			actual, err := nr.Do(ctx, tc.source, tc.options...)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tc.expected != actual {
				t.Errorf("Do unmatch: expected '%s' but got '%s'", tc.expected, actual)
				return
			}
		})
	}
}
