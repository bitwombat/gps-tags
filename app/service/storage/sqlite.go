package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/bitwombat/gps-tags/device"
	"github.com/google/uuid"
	_ "modernc.org/sqlite"
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

type sqliteStorer struct {
	db *sql.DB
}

// This exists only to take advantage of the sql.Null type, which I'm using
// mostly out of caution rather than understanding a real threat of error.
// Having NOT NULL everywhere in the database would be the smart way of
// enforcing data integrity. This is mostly an invariant.
// Its at this level because the verification functions need it.
// TODO: Make sure schema has NOT NULL for everything.
type PositionRecordDAO struct {
	SerNo     sql.NullInt32
	SeqNo     sql.NullInt32
	Reason    sql.NullInt32
	Latitude  sql.NullFloat64
	Longitude sql.NullFloat64
	Altitude  sql.NullInt32
	Speed     sql.NullInt32
	DateUTC   sql.NullString
	GpsUTC    sql.NullString
	PosAcc    sql.NullInt32
	GpsStatus sql.NullInt32
	Battery   sql.NullInt32 // TODO: Probably call this InternalBatteryVoltage, to match the db, or at least BatteryVoltage. Also, figure out if the LoadedVoltage field is more useful.
}

type PointRecordDAO struct {
	SerNo     sql.NullInt32
	SeqNo     sql.NullInt32
	Latitude  sql.NullFloat64
	Longitude sql.NullFloat64
}

func NewSQLiteStorer(dataSourceName string) (sqliteStorer, error) {
	db, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return sqliteStorer{}, fmt.Errorf("opening database: %w", err)
	}
	var ss sqliteStorer
	ss.db = db

	return ss, nil
}

// TODO: make sure to set PRAGMA foreign_keys = true for every connection.
func (s sqliteStorer) WriteTx(ctx context.Context, tx device.TagTx) (string, error) {
	// uuid.SetRand(rand.New(rand.NewSource(1)))  // TODO: Make it deterministic for testing (saving this line for later)
	txID := uuid.NewString() // TODO: Maybe create this when unmarshaling? The field's there.

	// TODO: use fmt.Sprintf everywhere, or turn this into ? $1
	_, err := s.db.ExecContext(ctx, fmt.Sprintf("INSERT INTO tx (ID, ProdID, Fw, SerNo, IMEI, ICCID) VALUES ('%s', %v, '%s', %v, '%s', '%s');", txID, tx.ProdID, tx.Fw, tx.SerNo, tx.IMEI, tx.ICCID)) // TODO: By not using query parameters, ? or $1, this code is subject to injection attack by the device.
	if err != nil {
		return "", err
	}

	for _, r := range tx.Records {
		rId, err := insertRecord(r, ctx, s.db, txID)
		if err != nil {
			return "", err
		}
		if r.GPSReading != nil {
			err := insertGPSReading(*r.GPSReading, ctx, s.db, rId)
			if err != nil {
				return "", err
			}
		}
		if r.GPIOReading != nil {
			err := insertGPIOReading(*r.GPIOReading, ctx, s.db, rId)
			if err != nil {
				return "", err
			}
		}
		if r.AnalogueReading != nil {
			err := insertAnalogueReading(*r.AnalogueReading, ctx, s.db, rId)
			if err != nil {
				return "", err
			}
		}
		if r.TripTypeReading != nil {
			err := insertTripTypeReading(*r.TripTypeReading, ctx, s.db, rId)
			if err != nil {
				return "", err
			}
		}
	}
	return txID, nil
}

func insertRecord(r device.Record, ctx context.Context, db *sql.DB, txID string) (string, error) {
	// uuid.SetRand(rand.New(rand.NewSource(1)))  // TODO: Make it deterministic
	// for testing? (saving this line for later)
	rID := uuid.NewString()
	_, err := db.ExecContext(ctx, `INSERT INTO record (ID, TxID, DeviceDateTime, SeqNo, Reason) VALUES (?, ?, ?, ?, ?);`, rID, txID, r.DateUTC, r.SeqNo, r.Reason)

	return rID, err
}

func insertGPSReading(g device.GPSReading, ctx context.Context, db *sql.DB, recordID string) error {
	_, err := db.ExecContext(ctx, `INSERT INTO gpsReading (RecordID, Spd, SpdAcc, Head, GpsStat, GpsUTC, Lat, Lng, Alt, PosAcc, PDOP) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`, recordID, g.Spd, g.SpdAcc, g.Head, g.GpsStat, g.GpsUTC, g.Lat, g.Long, g.Alt, g.PosAcc, g.PDOP)
	return err
}

func insertGPIOReading(g device.GPIOReading, ctx context.Context, db *sql.DB, recordID string) error {
	_, err := db.ExecContext(ctx, `INSERT INTO gpioReading (RecordID, DIn, DOut, DevStat) VALUES (?, ?, ?, ?);`, recordID, g.DIn, g.DOut, g.DevStat)
	return err
}

func insertAnalogueReading(a device.AnalogueReading, ctx context.Context, db *sql.DB, recordID string) error {
	_, err := db.ExecContext(ctx, `INSERT INTO analogueReading (RecordID, InternalBatteryVoltage, Temperature, LastGSMCQ, LoadedVoltage) VALUES ($1, $2, $3, $4, $5);`, recordID, a.AnalogueData.Num1, a.AnalogueData.Num3, a.AnalogueData.Num4, a.AnalogueData.Num5) // TODO: Leakage from device into business domain. Map these to better names like old tx.go had.
	// TODO: Wrong number of parameters fails silently - record not inserted.
	// Sprintf would complain... but then injection attacks?
	return err
}

func insertTripTypeReading(t device.TripTypeReading, ctx context.Context, db *sql.DB, recordID string) error {
	_, err := db.ExecContext(ctx, `INSERT INTO tripTypeReading (RecordID, Tt, Trim) VALUES (?, ?, ?);`, recordID, t.Tt, t.Trim)
	return err
}

func (s sqliteStorer) GetLastPositions(ctx context.Context) ([]PositionRecord, error) {
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
        record.DeviceDateTime,
		analogueReading.InternalBatteryVoltage,
        ROW_NUMBER() OVER (PARTITION BY tx.SerNo ORDER BY record.DeviceDateTime DESC) as rn
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
	DeviceDateTime,
	GpsUTC,
	PosAcc,
	GpsStat,
	InternalBatteryVoltage
FROM
    RankedRecords
WHERE
    rn = 1
ORDER BY
    DeviceDateTime DESC
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
		err := rows.Scan(
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
			DateUTC:   prDAO.DateUTC.String,
			GpsUTC:    prDAO.GpsUTC.String,
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

// checkValidity checks validity of everything in a PositionRecordDAO except
// SeqNo, which we check separately so we can include it to provide better errors
// when another field is NULL
func isValid(pr PositionRecordDAO) bool {
	return (pr.SerNo.Valid &&
		pr.Reason.Valid &&
		pr.Latitude.Valid &&
		pr.Longitude.Valid &&
		pr.Altitude.Valid &&
		pr.Speed.Valid &&
		pr.DateUTC.Valid &&
		pr.GpsUTC.Valid &&
		pr.PosAcc.Valid &&
		pr.GpsStatus.Valid &&
		pr.Battery.Valid)
}

func (s sqliteStorer) GetLastNPositions(ctx context.Context, n int) ([]PathPointRecord, error) {
	query := `
WITH NumberedRecords AS (
    SELECT
		tx.SerNo,
		record.SeqNo,
		gpsReading.Lat,
		gpsReading.Lng,
        record.DeviceDateTime,
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
    rn <= %d
ORDER BY
    SerNo, DeviceDateTime DESC
;
`
	query = fmt.Sprintf(query, n)

	var pps = make(map[int32][]PathPoint) // TODO: Make PathPointRecord type this map so the later conversion isn't necessary

	rows, err := s.db.QueryContext(ctx, query)
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
		if !(prDAO.Latitude.Valid && prDAO.Longitude.Valid) {
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

	var pprs []PathPointRecord

	keys := maps.Keys(pps) // only need them sorted for testing. TODO: Fix test.
	keysSlice := slices.Sorted(keys)

	for _, k := range keysSlice {
		pprs = append(pprs, PathPointRecord{SerNo: int32(k), PathPoints: pps[int32(k)]})
	}

	return pprs, nil
}
