package narrator

import (
	"context"
	"fmt"
	"sync"

	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"

	"github.com/heptaliane/katarive-server/internal/model"
)

type SemaphoreNarratorManager struct {
	client pb.NarratorServiceClient

	name      string
	version   string
	encodings []*pb.AudioEncoding
	speakers  []*pb.SpeakerInfo

	mu *sync.RWMutex
}

func (n *SemaphoreNarratorManager) Metadata() *model.NarratorManagerMetadata {
	return &model.NarratorManagerMetadata{
		Name:      fmt.Sprintf("%s.%s", n.name, n.version),
		Encodings: n.encodings,
		Speakers:  n.speakers,
	}
}
func (n *SemaphoreNarratorManager) Narrate(
	ctx context.Context,
	req *pb.NarrateRequest,
) (*pb.NarrateResponse, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.client.Narrate(ctx, req)
}

// Ensure NarratorManager implementation
var _ NarratorManager = new(SemaphoreNarratorManager)
