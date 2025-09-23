package model

import (
	"database/sql/driver"
	"time"

	"maragu.dev/errors"
)

// Time implements the driver.Valuer and sql.Scanner interfaces so that we can
// read and write proper time.Times to/from sqlite3 database.
// It lives here in model, but DAOs will need to use it, since that's the point
// of it. Yet, it shouldn't live in storage, since any business value that is
// intended to go into the db should use it. The alternative would be to copy
// any business value into a DAO that was identical except for dates (ie.
// turning time.Time into Time). This is too tedious for the value it would
// provide, so keep Time here.
type Time struct {
	T time.Time
}

// rfc3339Milli is like time.RFC3339Nano, but with millisecond precision, and
// fractional seconds do not have trailing zeros removed.
// Hat tip to https://www.golang.dk/articles/go-and-sqlite-in-the-cloud
const rfc3339Milli = "2006-01-02T15:04:05.000Z07:00"

// Value satisfies driver.Valuer interface.
func (t Time) Value() (driver.Value, error) {
	return t.T.UTC().Format(rfc3339Milli), nil
}

// Scan satisfies sql.Scanner interface.
func (t *Time) Scan(src any) error {
	if src == nil {
		return nil
	}

	s, ok := src.(string)
	if !ok {
		return errors.Newf("error scanning time, got %+v", src)
	}

	parsedT, err := time.Parse(rfc3339Milli, s)
	if err != nil {
		return err
	}

	t.T = parsedT.UTC()

	return nil
}

func TimeFromString(s string) (Time, error) {
	parsedT, err := time.Parse(time.DateTime, s)
	if err != nil {
		return Time{}, err
	}
	return Time{T: parsedT}, err
}
