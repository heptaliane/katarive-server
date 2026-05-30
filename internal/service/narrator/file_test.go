package narrator_test

import (
	"context"
	"fmt"
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
func TestFileNarratorRegistryHas(t *testing.T) {
	t.Parallel()

	basedir := os.TempDir()
	source := &model.SourceItem{
		Id:    "id",
		Title: "title",
	}
	narratorName := "narrator"
	var speakerId int32 = 1
	parentDir := filepath.Join(basedir, narratorName, fmt.Sprintf("%03d", speakerId))
	filename := filepath.Join(parentDir, "id_title.mp3")

	os.MkdirAll(parentDir, 0755)
	if err := os.WriteFile(filename, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create dummy file: %v", err)
	}

	cases := map[string]struct {
		source   *model.SourceItem
		options  []narrator.NarrateOption
		expected bool
	}{
		"cache hit": {
			source: source,
			options: []narrator.NarrateOption{
				narrator.WithEncoding(pb.AudioEncoding_AUDIO_ENCODING_MP3),
				narrator.WithNarrator(narratorName),
				narrator.WithSpeaker(speakerId),
			},
			expected: true,
		},
		"new source": {
			source: &model.SourceItem{
				Id:    "new-id",
				Title: "new-title",
			},
			options: []narrator.NarrateOption{
				narrator.WithEncoding(pb.AudioEncoding_AUDIO_ENCODING_MP3),
				narrator.WithNarrator(narratorName),
				narrator.WithSpeaker(speakerId),
			},
			expected: false,
		},
		"new options": {
			source: source,
			options: []narrator.NarrateOption{
				narrator.WithEncoding(pb.AudioEncoding_AUDIO_ENCODING_M4A),
				narrator.WithNarrator(narratorName),
				narrator.WithSpeaker(speakerId),
			},
			expected: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			nms := []narrator.NarratorManager{setupNarratorManager(t)}
			nr := narrator.NewFileNarratorRegistry(ctx, basedir, nms)

			actual := nr.Has(tc.source, tc.options...)
			if actual != tc.expected {
				t.Errorf("Has unmatch: expected %t but got %t", tc.expected, actual)
				return
			}
		})
	}
}
