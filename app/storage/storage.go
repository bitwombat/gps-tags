package storage

import (
	"context"
)

type Storage interface {
	WriteCommit(context.Context, string) error
	GetLastPositions() ([]PositionRecord, error)
}

// A field present in the document but not in the struct here doesn't break anything (is ignored).

// This is based on what Mongo returns, not what it should be.  For example,
// sub-document, and lat/long as arrays (only [0] is ever populated)
type PositionRecord struct {
	Document struct {
		SerNo     float64
		SeqNo     float64
		Reason    int64
		Latitude  []float64
		Longitude []float64
		Altitude  []float64
		Speed     []float64
		// DateUTC   float64
		// GpsUTC    string
		Battery   []float64
		SpeedAcc  []float64
		Heading   []float64
		PDOP      []float64
		PosAcc    []float64
		GpsStatus []float64
	}
}
