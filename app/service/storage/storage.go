package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/bitwombat/gps-tags/device"
)

type Storage interface {
	WriteTx(context.Context, device.TagTx) (string, error)
	GetLastPositions() ([]PositionRecord, error)
	GetLastNPositions(int) ([]PathPointRecord, error)
}

type PositionRecord struct {
	SerNo     int32
	SeqNo     int32
	Reason    int32
	Latitude  float64
	Longitude float64
	Altitude  int32
	Speed     int32
	DateUTC   string
	GpsUTC    string
	PosAcc    int32
	GpsStatus int32
	Battery   int32
}

type PathPoint struct {
	Latitude  float64
	Longitude float64
}

type PathPointRecord struct {
	SerNo      float64
	PathPoints []PathPoint
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
	const heartBeatTimeInMinutes = 10 * time.Minute

	// Parse the given time string
	t, err := time.Parse(time.DateTime, timeStr)
	if err != nil {
		return "black"
	}

	// Calculate the difference
	diff := Now().Sub(t)

	if diff < heartBeatTimeInMinutes+1*time.Minute { // if it's reported in properly, recently
		return "red"
	} else if diff < time.Hour { // somewhat recently, probably working
		return "#a23535"
	} else { // not good
		return "#8d8d8d"
	}
}
