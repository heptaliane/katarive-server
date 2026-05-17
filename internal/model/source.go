package model

import (
	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
)

type SourceItem = pb.SourceItem
type SourceCollection = pb.SourceCollection
type SourceSummary = pb.SourceSummary
type SourceCollectionPackage struct {
	Collection *SourceCollection
	Sources    []*SourceSummary
}
