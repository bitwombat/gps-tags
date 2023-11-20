package storage

import (
	"context"
	"fmt"
	"time"
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
		DateUTC   string
		Latitude  []float64
		Longitude []float64
		Altitude  []float64
		Speed     []float64
		GpsUTC    []string
		SpeedAcc  []float64
		Heading   []float64
		PDOP      []float64
		PosAcc    []float64
		GpsStatus []float64
		Battery   []float64
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
	DateUTC   string
	GpsUTC    string
	SpeedAcc  float64
	Heading   float64
	PDOP      float64
	PosAcc    float64
	GpsStatus float64
	Battery   float64
}

func TimeAgoAsText(timeStr string, Now func() time.Time) string {
	// Parse the given time string
	t, err := time.Parse(time.DateTime, timeStr)
	if err != nil {
		return "<time parsing error>"
	}

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
}

func TimeAgoInMinutes(timeStr string, Now func() time.Time) int {
	// Parse the given time string
	t, err := time.Parse(time.DateTime, timeStr)
	if err != nil {
		return 9999
	}

	// Calculate the difference
	diff := Now().Sub(t)

	return int(diff.Minutes())
}

func TimeAgoInColour(timeStr string, Now func() time.Time) string {
	const heartBeatTimeInMinutes = 10

	// Parse the given time string
	t, err := time.Parse(time.DateTime, timeStr)
	if err != nil {
		return "black"
	}

	// Calculate the difference
	diff := Now().Sub(t)

	if diff < heartBeatTimeInMinutes+1 { // if it's reported in properly, recently
		return "red"
	} else if diff < time.Hour { // somewhat recently, probably working
		return "#ff6969"
	} else { // not good
		return "#8d8d8d"
	}
}
