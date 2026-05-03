package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"

	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
)

// ==============================
// Interfaces for Narrator handlers
// ==============================

//go:generate mockgen -source=$GOFILE -destination=mock/mock_$GOFILE -package=mock
type NarratorManager interface {
	Do(ctx context.Context, basePath string, text string, opts ...NarrateOption) (string, error)
	GetName() string
	SupportedOptions() []*pb.NarratorOption
}

//go:generate mockgen -source=$GOFILE -destination=mock/mock_$GOFILE -package=mock
type NarratorRegistry interface {
	Use(name string)
	Narrators() []string
	Do(ctx context.Context, url string, text string, opts ...NarrateOption) (string, error)
}

// -----------------
// Helper components
// -----------------

type narrateOptions struct {
	opts     map[string]string
	language pb.Language
	encoding pb.AudioEncoding
}
type NarrateOption func(*narrateOptions)

func WithNarrateOption(key string, value string) NarrateOption {
	return func(opt *narrateOptions) {
		opt.opts[key] = value
	}
}
func WithNarrateLanguage(language pb.Language) NarrateOption {
	return func(opt *narrateOptions) {
		opt.language = language
	}
}
func WithNarrateEncoding(encoding pb.AudioEncoding) NarrateOption {
	return func(opt *narrateOptions) {
		opt.encoding = encoding
	}
}

// ============================
// NarratorManager Implementation
// ============================

// ----------------------
// SemaphoreNarratorManager
// ----------------------
type SemaphoreNarratorManager struct {
	narrator pb.NarratorServiceClient

	name      string
	version   string
	encodings []pb.AudioEncoding
	options   []*pb.NarratorOption

	mu *sync.RWMutex
}

func (n *SemaphoreNarratorManager) Do(
	ctx context.Context,
	basePath string,
	text string,
	opts ...NarrateOption,
) (string, error) {
	var options narrateOptions
	for _, opt := range opts {
		opt(&options)
	}

	if !slices.Contains(n.encodings, options.encoding) {
		return "", &UnsupportedEncodingError{
			Target:   n.GetName(),
			Encoding: options.encoding.String(),
		}
	}
	path := fmt.Sprintf("%s.%s", basePath, getAudioExtension(options.encoding))
	req := &pb.NarrateRequest{
		Path:     path,
		Text:     text,
		Language: options.language,
		Encoding: options.encoding,
		Options:  options.opts,
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	res, err := n.narrator.Narrate(ctx, req)
	if err != nil {
		return "", err
	}

	if res.GetError() {
		return "", &NarrateError{Reason: res.GetReason()}
	}

	return path, nil
}
func (n *SemaphoreNarratorManager) GetName() string {
	return fmt.Sprintf("%s:%s", n.name, n.version)
}
func (n *SemaphoreNarratorManager) SupportedOptions() []*pb.NarratorOption {
	return n.options
}

// Ensure SemaphoreNarratorManager implements NarratorManager
var _ NarratorManager = new(SemaphoreNarratorManager)

// -----------------
// Helper components
// -----------------

func NewSemaphoreNarratorManager(
	ctx context.Context,
	narrator pb.NarratorServiceClient,
) (*SemaphoreNarratorManager, error) {
	req := &pb.GetNarratorServiceMetadataRequest{}
	res, err := narrator.GetNarratorServiceMetadata(ctx, req)
	if err != nil {
		return nil, err
	}

	return &SemaphoreNarratorManager{
		narrator:  narrator,
		name:      res.GetName(),
		version:   res.GetVersion(),
		options:   res.GetOptions(),
		encodings: res.GetSupportedEncoding(),
		mu:        new(sync.RWMutex),
	}, nil

}

// =============================
// NarratorRegistry Implementation
// =============================

// ------------------
// FileNarratorRegistry
// ------------------

type FileNarratorRegistry struct {
	basedir   string
	narrators map[string]NarratorManager
	cursor    NarratorManager
}

func (n *FileNarratorRegistry) Use(name string) {
	n.cursor = n.narrators[name]
}
func (n *FileNarratorRegistry) Narrators() []string {
	keys := make([]string, 0)
	for name := range n.narrators {
		keys = append(keys, name)
	}
	return keys
}
func (n *FileNarratorRegistry) Do(
	ctx context.Context,
	url string,
	text string,
	opts ...NarrateOption,
) (string, error) {
	if n.cursor == nil {
		return "", UnspecifiedNarratorError
	}

	basedir := filepath.Join(n.basedir, n.cursor.GetName())
	os.MkdirAll(basedir, 0755)

	path := filepath.Join(basedir, url2filename(url))
	if Exists(path) {
		return path, nil
	}

	return n.cursor.Do(ctx, path, text, opts...)
}

// Ensure NarratorRegistry implements NarratorRegistry
var _ NarratorRegistry = new(FileNarratorRegistry)

// -----------------
// Helper components
// -----------------

func NewFileNarratorRegistry(
	basedir string,
	narrators []NarratorManager,
) *FileNarratorRegistry {
	var nms = make(map[string]NarratorManager)
	var cursor NarratorManager
	for _, narrator := range narrators {
		nms[narrator.GetName()] = narrator
		if cursor == nil {
			cursor = narrator
		}
	}

	return &FileNarratorRegistry{
		basedir:   basedir,
		narrators: nms,
		cursor:    cursor,
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
