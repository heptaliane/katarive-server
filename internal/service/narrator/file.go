package narrator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"

	"github.com/heptaliane/katarive-server/internal/model"
)

type FileNarratorRegistry struct {
	basedir string

	narrators map[string]NarratorManager
	metadata  []*model.NarratorManagerMetadata
}

func (r *FileNarratorRegistry) Metadata() []*model.NarratorManagerMetadata {
	return r.metadata
}

func (r *FileNarratorRegistry) Do(
	ctx context.Context,
	source *model.SourceItem,
	opts ...NarrateOption,
) error {
	var options narrateOption
	for _, opt := range opts {
		opt(&options)
	}

	path := r.path(source, &options)
	basedir := filepath.Dir(path)
	os.MkdirAll(basedir, 0755)

	nm := r.narrators[options.narrator]
	req := &pb.NarrateRequest{
		Path:      path,
		Text:      source.GetContent(),
		Language:  source.GetLanguage(),
		SpeakerId: options.speakerId,
		Encoding:  options.encoding,
	}

	_, err := nm.Narrate(ctx, req)

	return err
}
func (r *FileNarratorRegistry) Get(
	source *model.SourceItem,
	opts ...NarrateOption,
) *model.NarrateResult {
	var options narrateOption
	for _, opt := range opts {
		opt(&options)
	}

	path := r.path(source, &options)
	if !Exists(path) {
		return nil
	}
	return &model.NarrateResult{Path: path}
}

// Ensure FileNarratorRegistry implements NarrateRegistry
var _ NarrateRegistry = new(FileNarratorRegistry)

// helpers
func (r *FileNarratorRegistry) path(
	source *model.SourceItem,
	options *narrateOption,
) string {
	basedir := filepath.Join(r.basedir, options.narrator, fmt.Sprintf("%03d", options.speakerId))
	extension := getAudioExtension(options.encoding)
	return filepath.Join(
		basedir,
		fmt.Sprintf("%s_%s.%s", source.GetId(), source.GetTitle(), extension),
	)
}
func NewFileNarratorRegistry(
	ctx context.Context,
	basedir string,
	nms []NarratorManager,
) *FileNarratorRegistry {
	var metadata []*model.NarratorManagerMetadata
	nmap := make(map[string]NarratorManager)
	for _, nm := range nms {
		d := nm.Metadata()
		metadata = append(metadata, d)
		nmap[d.Name] = nm
	}
	return &FileNarratorRegistry{
		basedir:   basedir,
		narrators: nmap,
		metadata:  metadata,
	}
}
func getAudioExtension(encoding pb.AudioEncoding) string {
	switch encoding {
	case pb.AudioEncoding_AUDIO_ENCODING_WAV:
		return "wav"
	case pb.AudioEncoding_AUDIO_ENCODING_MP3:
		return "mp3"
	case pb.AudioEncoding_AUDIO_ENCODING_M4A:
		return "m4a"
	}
	return ""
}
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
