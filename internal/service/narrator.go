package service

import (
	"context"
	"fmt"
	"sync"

	katarive "github.com/heptaliane/katarive-go-sdk"
	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
)

type narrateOptions struct {
	opts map[string]string
}
type NarrateOption func(*narrateOptions)

func WithNarrateOption(key string, value string) NarrateOption {
	return func(opt *narrateOptions) {
		opt.opts[key] = value
	}
}

type NarratorManager interface {
	Do(ctx context.Context, path string, text string, opts ...NarrateOption) error
	GetName() string
	SupportedOptions() []*pb.NarratorOption
}

type SemaphoreNarratorManager struct {
	narrator katarive.Narrator

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

	res, err := n.narrator.Narrate(ctx, path, text, options.opts)
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
	narrator katarive.Narrator,
) (*SemaphoreNarratorManager, error) {
	res, err := narrator.GetNarratorServiceMetadata(ctx)
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
