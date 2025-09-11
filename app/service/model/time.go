package model

import (
	"database/sql/driver"
	"time"

	"maragu.dev/errors"
)

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
