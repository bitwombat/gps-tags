package storage

import (
	"context"
)

type Storage interface {
	WriteCommit(context.Context, string) error
	GetLastPositions() ([]PositionRecord, error)
}

// This is based on what Mongo returns, not what it should be.  For example,
// sub-document, and lat/long as arrays (only [0] is ever populated)
type PositionRecord struct {
	Document struct {
		SerNo     float64
		SeqNo     float64
		Latitude  []float64
		Longitude []float64
	}
}
