package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/bitwombat/gps-tags/device"
	"github.com/bitwombat/gps-tags/model"
	"github.com/google/uuid"
	_ "modernc.org/sqlite" // library code isn't used directly
)

// Concurrency support in sqlite is complicated.
// I'm not even sure of the conclusion yet.
// Do concurrent writes block? If so, set the db timeout to 5s so they'll
// go through and not error out.
// Do concurrent reads and writes make weird stuff? This link says yes
// (last section).
// https://www.sqlite.org/isolation.html

// This link says writes queue up. So, set the db timeout to a few seconds so
// they don't error out.
// https://www.sqlite.org/whentouse.html
//	High Concurrency
//
//	SQLite supports an unlimited number of simultaneous readers, but it will
//	only allow one writer at any instant in time. For many situations, this is
//	not a problem. Writers queue up. Each application does its database work
//	quickly and moves on, and no lock lasts for more than a few dozen
//	milliseconds. But there are some applications that require more concurrency,
//	and those applications may need to seek a different solution.
//
// Other links:
// https://sqlite.org/wal.html#concurrency
// https://www.reddit.com/r/golang/comments/16xswxd/concurrency_when_writing_data_into_sqlite/

type SqliteStorer struct {
	db *sql.DB
}

type PointRecordDAO struct {
	SerNo     sql.NullInt32
	SeqNo     sql.NullInt32
	Latitude  sql.NullFloat64
	Longitude sql.NullFloat64
}

func NewSQLiteStorer(dataSourceName string) (SqliteStorer, error) {
	var dsn string
	if dataSourceName == ":memory:" {
		dsn = dataSourceName
	} else {
		dsn = "file://" + dataSourceName + "?_pragma=foreign_keys(1)"
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return SqliteStorer{}, fmt.Errorf("opening database: %w", err)
	}

	// NOTE: We really should check that the pragma has been set, because the
	// driver has to be compiled with support for it. The below code returns
	// <nil> for all three values, so that may indicate this driver doesn't have
	// support compiled in (though the docs indicate all pragmas are supported,
	// and use this one as an example:
	// https://pkg.go.dev/modernc.org/sqlite#Driver.Open
	// Perhaps try one of the other drivers someday, eg. "github.com/mattn/go-sqlite3"
	// as used by https://www.golang.dk/articles/go-and-sqlite-in-the-cloud
	// result, err := db.ExecContext(context.Background(), "PRAGMA foreign_keys;")
	// fmt.Println(err)
	// fmt.Println(result.LastInsertId())
	// fmt.Println(result.RowsAffected())

	var ss SqliteStorer
	ss.db = db

	return ss, nil
}

func (s SqliteStorer) WriteTx(ctx context.Context, tx device.TagTx) (string, error) {
	txID := uuid.NewString() // Not unmarshalling $oid for no real reason. Zero trust that it's unique this way.

	_, err := s.db.ExecContext(ctx, "INSERT INTO tx (ID, ProdID, Fw, SerNo, IMEI, ICCID) VALUES (?, ?, ?, ?, ?, ?);", txID, tx.ProdID, tx.Fw, tx.SerNo, tx.IMEI, tx.ICCID)
	if err != nil {
		return "", err
	}

	for _, r := range tx.Records {
		rID, err := insertRecord(ctx, r, s.db, txID)
		if err != nil {
			return "", err
		}
		if r.GPSReading != nil {
			err := insertGPSReading(ctx, *r.GPSReading, s.db, rID)
			if err != nil {
				return "", err
			}
		}
		if r.GPIOReading != nil {
			err := insertGPIOReading(ctx, *r.GPIOReading, s.db, rID)
			if err != nil {
				return "", err
			}
		}
		if r.AnalogueReading != nil {
			err := insertAnalogueReading(ctx, *r.AnalogueReading, s.db, rID)
			if err != nil {
				return "", err
			}
		}
		if r.TripTypeReading != nil {
			err := insertTripTypeReading(ctx, *r.TripTypeReading, s.db, rID)
			if err != nil {
				return "", err
			}
		}
	}
	return txID, nil
}

func insertRecord(ctx context.Context, r device.Record, db *sql.DB, txID string) (string, error) {
	rID := uuid.NewString()
	_, err := db.ExecContext(ctx, `INSERT INTO record (ID, TxID, DeviceUTC, SeqNo, Reason) VALUES (?, ?, ?, ?, ?);`, rID, txID, r.DateUTC, r.SeqNo, r.Reason)

	return rID, err
}

func insertGPSReading(ctx context.Context, g device.GPSReading, db *sql.DB, recordID string) error {
	_, err := db.ExecContext(ctx, `INSERT INTO gpsReading (RecordID, Spd, SpdAcc, Head, GpsStat, GpsUTC, Lat, Lng, Alt, PosAcc, PDOP) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`, recordID, g.Spd, g.SpdAcc, g.Head, g.GpsStat, g.GpsUTC, g.Lat, g.Long, g.Alt, g.PosAcc, g.PDOP)
	return err
}

func insertGPIOReading(ctx context.Context, g device.GPIOReading, db *sql.DB, recordID string) error {
	_, err := db.ExecContext(ctx, `INSERT INTO gpioReading (RecordID, DIn, DOut, DevStat) VALUES (?, ?, ?, ?);`, recordID, g.DIn, g.DOut, g.DevStat)
	return err
}

func insertAnalogueReading(ctx context.Context, a device.AnalogueReading, db *sql.DB, recordID string) error {
	_, err := db.ExecContext(ctx, `INSERT INTO analogueReading (RecordID, InternalBatteryVoltage, Temperature, LastGSMCQ, LoadedVoltage) VALUES (?, ?, ?, ?, ?);`, recordID, a.AnalogueData.Num1, a.AnalogueData.Num3, a.AnalogueData.Num4, a.AnalogueData.Num5) // TODO: Leakage from device into business domain. Map these to better names like old tx.go had.
	return err
}

func insertTripTypeReading(ctx context.Context, t device.TripTypeReading, db *sql.DB, recordID string) error {
	_, err := db.ExecContext(ctx, `INSERT INTO tripTypeReading (RecordID, Tt, Trim) VALUES (?, ?, ?);`, recordID, t.Tt, t.Trim)
	return err
}

func (s SqliteStorer) GetLastPositions(ctx context.Context) ([]PositionRecord, error) {
	// This type exists only to take advantage of the sql.Null type, which I'm using
	// mostly out of caution rather than understanding a real threat of error. NOT
	// NULL is on every column in the database. integrity.
	type PositionRecordDAO struct {
		SerNo     sql.NullInt32
		SeqNo     sql.NullInt32
		Reason    sql.NullInt32
		Latitude  sql.NullFloat64
		Longitude sql.NullFloat64
		Altitude  sql.NullInt32
		Speed     sql.NullInt32
		DateUTC   model.Time // TODO: Oops, probably shouldn't be using a model type in a DAO? model.Time should be in this package, as it's only used in this DAO.
		GpsUTC    model.Time //
		PosAcc    sql.NullInt32
		GpsStatus sql.NullInt32
		Battery   sql.NullInt32 // TODO: Probably call this InternalBatteryVoltage, to match the db, or at least BatteryVoltage. Also, figure out if the LoadedVoltage field is more useful.
	}

	isValid := func(pr PositionRecordDAO) bool {
		// seqNo checked separately so error message can use it.
		return (pr.SerNo.Valid &&
			pr.Reason.Valid &&
			pr.Latitude.Valid &&
			pr.Longitude.Valid &&
			pr.Altitude.Valid &&
			pr.Speed.Valid &&
			pr.PosAcc.Valid &&
			pr.GpsStatus.Valid &&
			pr.Battery.Valid)
	}

	query := `
WITH RankedRecords AS (
    SELECT
		tx.SerNo,
		record.SeqNo,
		record.Reason,
		gpsReading.Lat,
		gpsReading.Lng,
		gpsReading.Alt,
		gpsReading.Spd,
		gpsReading.GpsUTC,
		gpsReading.PosAcc,
		gpsReading.GpsStat,
        record.DeviceUTC,
		analogueReading.InternalBatteryVoltage,
        ROW_NUMBER() OVER (PARTITION BY tx.SerNo ORDER BY record.DeviceUTC DESC) as rn
    FROM
        tx
    JOIN
        record ON record.TxID = tx.ID
	JOIN
        gpsReading ON gpsReading.RecordID = record.ID
	JOIN
	    analogueReading on analogueReading.RecordID = record.ID
)
SELECT
    SerNo,
	SeqNo,
	Reason,
    Lat,
	Lng,
	Alt,
	Spd,
	DeviceUTC,
	GpsUTC,
	PosAcc,
	GpsStat,
	InternalBatteryVoltage
FROM
    RankedRecords
WHERE
    rn = 1
ORDER BY
    DeviceUTC DESC
LIMIT 5;
`

	var prs []PositionRecord

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return []PositionRecord{}, fmt.Errorf("error querying database for last positions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var prDAO PositionRecordDAO
		err = rows.Scan(
			&prDAO.SerNo,
			&prDAO.SeqNo,
			&prDAO.Reason,
			&prDAO.Latitude,
			&prDAO.Longitude,
			&prDAO.Altitude,
			&prDAO.Speed,
			&prDAO.DateUTC,
			&prDAO.GpsUTC,
			&prDAO.PosAcc,
			&prDAO.GpsStatus,
			&prDAO.Battery)
		if err != nil {
			return []PositionRecord{}, fmt.Errorf("error scanning row: %w", err)
		}
		if !prDAO.SeqNo.Valid {
			return []PositionRecord{}, errors.New("SeqNo in record is NULL")
		}
		if !isValid(prDAO) {
			return []PositionRecord{}, fmt.Errorf("one of the fields of database row for SeqNo %v is NULL", prDAO.SeqNo.Int32)
		}

		prs = append(prs, PositionRecord{
			SerNo:     prDAO.SerNo.Int32,
			SeqNo:     prDAO.SeqNo.Int32,
			Reason:    prDAO.Reason.Int32,
			Latitude:  prDAO.Latitude.Float64,
			Longitude: prDAO.Longitude.Float64,
			Altitude:  prDAO.Altitude.Int32,
			Speed:     prDAO.Speed.Int32,
			DateUTC:   prDAO.DateUTC.T,
			GpsUTC:    prDAO.GpsUTC.T,
			PosAcc:    prDAO.PosAcc.Int32,
			GpsStatus: prDAO.GpsStatus.Int32,
			Battery:   prDAO.Battery.Int32,
		})
	}

	err = rows.Err()
	if err != nil {
		return []PositionRecord{}, fmt.Errorf("error after scanning rows: %w", err)
	}

	return prs, nil
}

// checkValidity checks validity of everything in a PositionRecordDAO except:
//   - SeqNo, which we check separately so we can include it in other fields' error
//     messages.
//   - DateUTC and GpsUTC which are of type model.Time, and are validated in the
//     Time.Scan() method.
func (s SqliteStorer) GetLastNPositions(ctx context.Context, n int) ([]PathPointRecord, error) {
	query := `
WITH NumberedRecords AS (
    SELECT
		tx.SerNo,
		record.SeqNo,
		gpsReading.Lat,
		gpsReading.Lng,
        record.DeviceUTC,
        ROW_NUMBER() OVER (PARTITION BY tx.SerNo ORDER BY record.SeqNo DESC) AS rn
    FROM
        tx
    JOIN
        record ON record.TxID = tx.ID
	JOIN
        gpsReading ON gpsReading.RecordID = record.ID
)
SELECT
    SerNo,
	SeqNo,
    Lat,
	Lng
FROM
    NumberedRecords
WHERE
    rn <= ?
ORDER BY
    SerNo, DeviceUTC DESC
;
`
	pps := make(map[int32][]PathPoint) // TODO: Make PathPointRecord type this map so the later conversion isn't necessary

	rows, err := s.db.QueryContext(ctx, query, n)
	if err != nil {
		return []PathPointRecord{}, fmt.Errorf("error querying database for last N positions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var prDAO PointRecordDAO
		err := rows.Scan(
			&prDAO.SerNo,
			&prDAO.SeqNo,
			&prDAO.Latitude,
			&prDAO.Longitude,
		)
		if err != nil {
			return []PathPointRecord{}, fmt.Errorf("error scanning row: %w", err)
		}
		if !prDAO.SeqNo.Valid {
			return []PathPointRecord{}, errors.New("SeqNo in record is NULL")
		}
		if !prDAO.Latitude.Valid || !prDAO.Longitude.Valid {
			return []PathPointRecord{}, fmt.Errorf("one of the fields of database row for SeqNo %v is NULL", prDAO.SeqNo.Int32)
		}

		pps[prDAO.SerNo.Int32] = append(pps[prDAO.SerNo.Int32], PathPoint{
			Latitude:  prDAO.Latitude.Float64,
			Longitude: prDAO.Longitude.Float64,
		})
	}

	err = rows.Err()
	if err != nil {
		return []PathPointRecord{}, fmt.Errorf("error after scanning rows: %w", err)
	}

	keys := maps.Keys(pps) // only need them sorted for testing. TODO: Fix test.
	keysSlice := slices.Sorted(keys)

	pprs := make([]PathPointRecord, len(keysSlice))

	for i, k := range keysSlice {
		pprs[i] = PathPointRecord{SerNo: k, PathPoints: pps[k]}
	}

	return pprs, nil
}
