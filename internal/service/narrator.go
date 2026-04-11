package service

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
)

type narrateOptions struct {
	opts     map[string]string
	language pb.Language
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

type NarratorManager interface {
	Do(ctx context.Context, path string, text string, opts ...NarrateOption) error
	GetName() string
	SupportedOptions() []*pb.NarratorOption
}

type SemaphoreNarratorManager struct {
	narrator pb.NarratorServiceServer

	name    string
	version string
	options []*pb.NarratorOption

	mu *sync.RWMutex
}

func (n *SemaphoreNarratorManager) Do(
	ctx context.Context,
	path string,
	text string,
	opts ...NarrateOption,
) error {
	var options narrateOptions
	for _, opt := range opts {
		opt(&options)
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	req := &pb.NarrateRequest{
		Path:     path,
		Text:     text,
		Language: options.language,
		Options:  options.opts,
	}
	res, err := n.narrator.Narrate(ctx, req)
	if err != nil {
		return err
	}

	if res.GetError() {
		return &NarrateError{Reason: res.GetReason()}
	}

	return nil
}
func (n *SemaphoreNarratorManager) GetName() string {
	return fmt.Sprintf("%s:%s", n.name, n.version)
}
func (n *SemaphoreNarratorManager) SupportedOptions() []*pb.NarratorOption {
	return n.options
}

// Ensure SemaphoreNarratorManager implements NarratorManager
var _ NarratorManager = new(SemaphoreNarratorManager)

func NewSemaphoreNarratorManager(
	ctx context.Context,
	narrator pb.NarratorServiceServer,
) (*SemaphoreNarratorManager, error) {
	req := &pb.GetNarratorServiceMetadataRequest{}
	res, err := narrator.GetNarratorServiceMetadata(ctx, req)
	if err != nil {
		return nil, err
	}

	return &SemaphoreNarratorManager{
		narrator: narrator,
		name:     res.GetName(),
		version:  res.GetVersion(),
		options:  res.GetOptions(),
		mu:       new(sync.RWMutex),
	}, nil

}

type MockNarratorManager struct {
	NarrateResult error
	Name          string
	Options       []*pb.NarratorOption
}

func (n *MockNarratorManager) Do(
	ctx context.Context,
	path string,
	text string,
	opts ...NarrateOption,
) error {
	return n.NarrateResult
}
func (n *MockNarratorManager) GetName() string {
	return n.Name
}
func (n *MockNarratorManager) SupportedOptions() []*pb.NarratorOption {
	return n.Options
}

// Ensure MockNarratorManager implements NarratorManager
var _ NarratorManager = new(MockNarratorManager)

type NarratorRegistry struct {
	basedir   string
	narrators map[string]NarratorManager
	cursor    NarratorManager
}

func (n *NarratorRegistry) Use(name string) {
	n.cursor = n.narrators[name]
}
func (n *NarratorRegistry) Narrators() []string {
	keys := make([]string, 0)
	for name, _ := range n.narrators {
		keys = append(keys, name)
	}
	return keys
}
func (n *NarratorRegistry) GetNarration(
	ctx context.Context,
	url string,
	text string,
	opts ...NarrateOption,
) (string, error) {
	if n.cursor == nil {
		return "", UnspecifiedNarratorError
	}

	filename := fmt.Sprintf("%s.json", url2filename(url))
	path := filepath.Join(n.basedir, n.cursor.GetName(), filename)
	if Exists(path) {
		return path, nil
	}

	err := n.cursor.Do(ctx, path, text, opts...)
	if err != nil {
		return "", err
	}
	return path, err
}
func NewNarratorRegistry(
	basedir string,
	narrators []NarratorManager,
) *NarratorRegistry {
	var nms = make(map[string]NarratorManager)
	for _, narrator := range narrators {
		nms[narrator.GetName()] = narrator
	}

	return &NarratorRegistry{
		basedir:   basedir,
		narrators: nms,
	}
}
