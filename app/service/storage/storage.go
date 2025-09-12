package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/bitwombat/gps-tags/model"
)

type Storage interface {
	WriteTx(context.Context, model.TagTx) (string, error)
	GetLastStatuses(context.Context) ([]Status, error)
	GetLastNCoords(context.Context, int) (Coords, error)
}

type Status struct { // TODO: Change this to a map like the Coords type. For consistency, better looking function signatures.
	SerNo     int32
	SeqNo     int32
	Reason    int32
	Latitude  float64
	Longitude float64
	Altitude  int32
	Speed     int32
	DateUTC   time.Time
	GpsUTC    time.Time
	PosAcc    int32
	GpsStatus int32
	Battery   int32
}

type Coord struct {
	Latitude  float64
	Longitude float64
}

type Coords map[int32][]Coord

func StrTimeAgoAsText(ts string, now func() time.Time) string {
	// Parse the given time string
	t, err := time.Parse(time.DateTime, ts)
	if err != nil {
		return "<time parsing error>"
	}
	return TimeAgoAsText(t, now)
}

func TimeAgoAsText(t time.Time, now func() time.Time) string {
	// Calculate the difference
	diff := now().Sub(t)

	// Format
	if diff < time.Hour {
		return fmt.Sprintf("%d minutes", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		minutes := int(diff.Minutes()) % 60

		return fmt.Sprintf("%d hours, %d minutes", hours, minutes)
	}

	days := int(diff.Hours()) / 24
	hours := int(diff.Hours()) % 24
	minutes := int(diff.Minutes()) % 60

	return fmt.Sprintf("%d days, %d hours, %d minutes", days, hours, minutes)
}

func TimeAgoInMinutes(t time.Time, now func() time.Time) int {
	// Calculate the difference
	diff := now().Sub(t)

	return int(diff.Minutes())
}

func TimeAgoInColour(t time.Time, now func() time.Time) string {
	const heartBeatTime = 10 * time.Minute

	// Calculate the difference
	diff := now().Sub(t)

	if diff < heartBeatTime+1*time.Minute { // if it's reported in properly, recently
		return "red"
	} else if diff < time.Hour { // somewhat recently, probably working
		return "#a23535"
	}
	// not good
	return "#8d8d8d"
}
