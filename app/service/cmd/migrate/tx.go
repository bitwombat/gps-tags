package main

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
	"maragu.dev/errors"
)

// These types are the internal representation of GPS tag data. It's based on
// the Yabby3 device (the first/only device supported so far).
//
// It only exists in this dir so that it can be a package and thus pulled in by
// the cmd/migrate script. Since that script is basically a one-off, perhaps
// delete it in the future and move this to live with main.go.
//
// It doesn't live in device/ because it's device independent, and its type
// names would collide with the ones in yabby3.go.
//
// From the device, a Record's Fields may be one of several types. This isn't
// great for downstream Go code, so these types here serve to provide a sane
// internal type for use in code.
//
// These types are also used for the one-off migration from MongoDB. Records
// read from MongoDB have floats where integers should be. These types here have
// ints in the correct places.
//
// Most field names are kept the same as the device to reduce mental effort if
// mapping these back to the raw device data.
//
// Reminder on the structure: A Tx has Records which have Fields

type TagTxs []TagTx

type TagTx struct {
	ID      string
	ProdID  int
	Fw      string
	SerNo   int
	Imei    string
	Iccid   string
	Records []Record
}

type Record struct {
	DateUTC Time
	SeqNo   int
	Reason  int
	Fields  []Field
}

type Field interface {
	toSQL(string) string
}

type GPSReading struct { // FType0
	Spd     int
	SpdAcc  int
	Head    int
	GpsStat int
	GpsUTC  Time
	Lat     float64
	Long    float64
	Alt     int
	PosAcc  int
	Pdop    int
}

type GPIOReading struct { // FType2
	DIn     int
	DOut    int
	DevStat int
}

type AnalogueReading struct { // FType6
	InternalBatteryVoltage int
	Temperature            int
	LastGSMCQ              int
	LoadedVoltage          int
}

type TripTypeReading struct { // FType15
	Tt   int
	Trim int
}

// ToSQL creates a string of SQL commands for the given tag transmission.
func (t TagTx) ToSQL() string {
	s := fmt.Sprintf("INSERT INTO tx (ID, ProdID, Fw, SerNo, Imei, Iccid) VALUES ('%s', %v, '%s', %v, '%s', '%s');\n", t.ID, t.ProdID, t.Fw, t.SerNo, t.Imei, t.Iccid)

	for _, r := range t.Records {
		s += r.toSQL(t.ID)
	}

	return s
}

// toSQL turns a Record into SQL commands, iterating through its Fields.
func (r Record) toSQL(txID string) string {
	// TODO: If the r.DateUTC is using model.Time. Should it be using plain
	// time.Time? I thought I made that change already. Is there more than one
	// model that I'm not remembering?

	rID := uuid.NewString()
	dateStr, _ := r.DateUTC.Value() //nolint:errcheck // our implementation of Value() always returns nil error
	s := fmt.Sprintf("INSERT INTO record (ID, TXID, DeviceUTC, SeqNo, Reason) VALUES ('%s', '%s', '%s', %v, %v);\n", rID, txID, dateStr, r.SeqNo, r.Reason)
	for _, f := range r.Fields {
		s += f.toSQL(rID)
	}

	return s
}

// toSQL turns a GPSReading field into SQL commands.
func (g GPSReading) toSQL(recordID string) string {
	dateStr, _ := g.GpsUTC.Value() //nolint:errcheck // our implementation of Value() always returns nil error
	return fmt.Sprintf("INSERT INTO gpsReading (RecordID, Spd, SpdAcc, Head, GpsStat, GpsUTC, Lat, Lng, Alt, PosAcc, Pdop) VALUES ('%s', %v, %v, %v, %v, '%s', %.9f, %.9f, %v, %v, %v);\n", recordID, g.Spd, g.SpdAcc, g.Head, g.GpsStat, dateStr, g.Lat, g.Long, g.Alt, g.PosAcc, g.Pdop)
}

// toSQL turns a GPIOReading field into SQL commands.
func (g GPIOReading) toSQL(recordID string) string {
	return fmt.Sprintf("INSERT INTO gpioReading (RecordID, DIn, DOut, DevStat) VALUES ('%s', %v, %v, %v);\n", recordID, g.DIn, g.DOut, g.DevStat)
}

// toSQL turns an AnalogueReading field into SQL commands.
func (a AnalogueReading) toSQL(recordID string) string {
	return fmt.Sprintf("INSERT INTO analogueReading (RecordID, InternalBatteryVoltage, Temperature, LastGSMCQ, LoadedVoltage) VALUES ('%s', %v, %v, %v, %v);\n", recordID, a.InternalBatteryVoltage, a.Temperature, a.LastGSMCQ, a.LoadedVoltage)
}

// toSQL turns an TripTypeReading field into SQL commands.
func (t TripTypeReading) toSQL(recordID string) string {
	return fmt.Sprintf("INSERT INTO tripTypeReading (RecordID, Tt, Trim) VALUES ('%s', %v, %v);\n", recordID, t.Tt, t.Trim)
}

type Time struct {
	T time.Time
}

// rfc3339Milli is like time.RFC3339Nano, but with millisecond precision, and
// fractional seconds do not have trailing zeros removed.
// Hat tip to https://www.golang.dk/articles/go-and-sqlite-in-the-cloud
const rfc3339Milli = "2006-01-02T15:04:05.000Z07:00"

// Value satisfies driver.Valuer interface.
func (t *Time) Value() (driver.Value, error) {
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
