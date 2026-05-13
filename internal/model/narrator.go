package model

import (
	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
)

type NarratorManagerMetadata struct {
	Name      string
	Encodings []*pb.AudioEncoding
	Speakers  []*pb.SpeakerInfo
}
