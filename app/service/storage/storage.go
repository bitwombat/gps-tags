package storage

import (
	"context"
	"time"

	"github.com/bitwombat/gps-tags/model"
)

type Storage interface {
	WriteTx(context.Context, model.TagTx) (string, error)
	GetLastStatuses(context.Context) (Statuses, error)
	GetLastNCoords(context.Context, int) (Coords, error)
}

type Status struct {
	SeqNo     int32
	Reason    model.ReasonCode
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

type Statuses map[int32]Status

type Coord struct {
	Latitude  float64
	Longitude float64
}

type Coords map[int32][]Coord
