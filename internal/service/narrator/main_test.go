package narrator_test

import (
	"context"
	"errors"
	"os"
	"testing"

	pbmock "github.com/heptaliane/katarive-go-sdk/gen/mock/plugin/v1"
	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
)

var gnsmr *pb.GetNarratorServiceMetadataResponse
var nr *pb.NarrateResponse
var ne error

const VALID_TEXT string = "valid"

func TestMain(m *testing.M) {
	setupGetNarratorServiceMetadataResponse()
	setupNarrateResponse()
	setupError()

	code := m.Run()

	os.Exit(code)
}

func setupGetNarratorServiceMetadataResponse() {
	gnsmr = &pb.GetNarratorServiceMetadataResponse{
		Name:    "narrator",
		Version: "v1",
		SupportedEncoding: []pb.AudioEncoding{
			pb.AudioEncoding_AUDIO_ENCODING_WAV,
			pb.AudioEncoding_AUDIO_ENCODING_MP3,
			pb.AudioEncoding_AUDIO_ENCODING_M4A,
		},
		Speakers: []*pb.SpeakerInfo{
			{Id: 1, Name: "speaker1"},
			{Id: 2, Name: "speaker2"},
		},
	}
}
func setupNarrateResponse() {
	nr = &pb.NarrateResponse{}
}
func setupError() {
	ne = errors.New("Narrate() failed")
}
func setupNarratorServiceClient(t *testing.T) pb.NarratorServiceClient {
	t.Helper()

	nsc := pbmock.NewMockNarratorServiceClient(gomock.NewController(t))

	nsc.EXPECT().GetNarratorServiceMetadata(gomock.Any(), gomock.Any()).
		Return(gnsmr, nil).AnyTimes()
	nsc.EXPECT().Narrate(gomock.Any(), gomock.Any()).
		DoAndReturn(func(
			ctx context.Context,
			req *pb.NarrateRequest,
			opt ...grpc.CallOption,
		) (*pb.NarrateResponse, error) {
			if req.GetText() == VALID_TEXT {
				return nr, nil
			}
			return nil, ne
		})

	return nsc
}
