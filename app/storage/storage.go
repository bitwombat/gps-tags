package storage

import (
	"context"
)

type Storage interface {
	WriteCommit(context.Context, string) error
	GetLastPositions() ([]PositionRecord, error)
}

// A field present in the document but not in the struct here doesn't break anything (is ignored).

// What Mongo returns is odd: sub-document, and single numbers (e.g lat/long) as
// arrays. Only [0] is ever populated.
type MongoPositionRecord struct {
	Document struct {
		SerNo     float64
		SeqNo     float64
		Reason    int64
		Latitude  []float64
		Longitude []float64
		Altitude  []float64
		Speed     []float64
		// DateUTC   float64
		GpsUTC    []string
		Battery   []float64
		SpeedAcc  []float64
		Heading   []float64
		PDOP      []float64
		PosAcc    []float64
		GpsStatus []float64
	}
}

type PositionRecord struct {
	SerNo     float64
	SeqNo     float64
	Reason    int64
	Latitude  float64
	Longitude float64
	Altitude  float64
	Speed     float64
	// DateUTC   float64
	GpsUTC    string
	Battery   float64
	SpeedAcc  float64
	Heading   float64
	PDOP      float64
	PosAcc    float64
	GpsStatus float64
}

func MarshalPositionRecord(m MongoPositionRecord) *PositionRecord {
	pr := &PositionRecord{
		SerNo:  m.Document.SerNo,
		SeqNo:  m.Document.SeqNo,
		Reason: m.Document.Reason,
		// DateUTC:   m.Document.DateUTC,
		// GpsUTC:    m.Document.GpsUTC,
func TimeAgo(timeStr string, Now func() time.Time) string {
	// Parse the given time string
	t, err := time.Parse(time.DateTime, timeStr)
	if err != nil {
		return "<time parsing error>"
	}

	// These are probably only ever absent because of tests which intentionally
	// have incomplete records...
	if len(m.Document.Latitude) > 0 {
		pr.Latitude = m.Document.Latitude[0]
	// Calculate the difference
	diff := Now().Sub(t)

	// Format
	if diff < time.Hour {
		return fmt.Sprintf("%d minutes", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		minutes := int(diff.Minutes()) % 60
		return fmt.Sprintf("%d hours, %d minutes", hours, minutes)
	} else {
		days := int(diff.Hours()) / 24
		hours := int(diff.Hours()) % 24
		minutes := int(diff.Minutes()) % 60
		return fmt.Sprintf("%d days, %d hours, %d minutes", days, hours, minutes)
	}

	if len(m.Document.Longitude) > 0 {
		pr.Longitude = m.Document.Longitude[0]
	}

	if len(m.Document.Altitude) > 0 {
		pr.Altitude = m.Document.Altitude[0]
	}

	if len(m.Document.Speed) > 0 {
		pr.Speed = m.Document.Speed[0]
	}

	if len(m.Document.Battery) > 0 {
		pr.Battery = m.Document.Battery[0]
	}

	if len(m.Document.SpeedAcc) > 0 {
		pr.SpeedAcc = m.Document.SpeedAcc[0]
	}

	if len(m.Document.Heading) > 0 {
		pr.Heading = m.Document.Heading[0]
	}

	if len(m.Document.PDOP) > 0 {
		pr.PDOP = m.Document.PDOP[0]
	}

	if len(m.Document.PosAcc) > 0 {
		pr.PosAcc = m.Document.PosAcc[0]
	}

	if len(m.Document.GpsStatus) > 0 {
		pr.GpsStatus = m.Document.GpsStatus[0]
	}

	return pr
}
