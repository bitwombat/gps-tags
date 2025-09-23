package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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

func NewSQLiteStorer(dataSourceName string) (SqliteStorer, error) {
	var dsn string
	if dataSourceName == ":memory:" {
		dsn = dataSourceName
	} else {
		dsn = "file:" + dataSourceName + "?_pragma=foreign_keys(1)"
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

func (s SqliteStorer) WriteTx(ctx context.Context, tx model.TagTx) (string, error) {
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

func insertRecord(ctx context.Context, r model.Record, db *sql.DB, txID string) (string, error) {
	rID := uuid.NewString()
	_, err := db.ExecContext(ctx, `INSERT INTO record (ID, TxID, DeviceUTC, SeqNo, Reason) VALUES (?, ?, ?, ?, ?);`, rID, txID, r.DateUTC, r.SeqNo, r.Reason)

	return rID, err
}

func insertGPSReading(ctx context.Context, g model.GPSReading, db *sql.DB, recordID string) error {
	_, err := db.ExecContext(ctx, `INSERT INTO gpsReading (RecordID, Spd, SpdAcc, Head, GpsStat, GpsUTC, Lat, Lng, Alt, PosAcc, PDOP) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`, recordID, g.Spd, g.SpdAcc, g.Head, g.GpsStat, g.GpsUTC, g.Lat, g.Long, g.Alt, g.PosAcc, g.PDOP)
	return err
}

func insertGPIOReading(ctx context.Context, g model.GPIOReading, db *sql.DB, recordID string) error {
	_, err := db.ExecContext(ctx, `INSERT INTO gpioReading (RecordID, DIn, DOut, DevStat) VALUES (?, ?, ?, ?);`, recordID, g.DIn, g.DOut, g.DevStat)
	return err
}

func insertAnalogueReading(ctx context.Context, a model.AnalogueReading, db *sql.DB, recordID string) error {
	_, err := db.ExecContext(ctx, `INSERT INTO analogueReading (RecordID, InternalBatteryVoltage, Temperature, LastGSMCQ, LoadedVoltage) VALUES (?, ?, ?, ?, ?);`, recordID, a.InternalBatteryVoltage, a.Temperature, a.LastGSMCQ, a.LoadedVoltage)
	return err
}

func insertTripTypeReading(ctx context.Context, t model.TripTypeReading, db *sql.DB, recordID string) error {
	_, err := db.ExecContext(ctx, `INSERT INTO tripTypeReading (RecordID, Tt, Trim) VALUES (?, ?, ?);`, recordID, t.Tt, t.Trim)
	return err
}

func (s SqliteStorer) GetLastStatuses(ctx context.Context) (Statuses, error) {
	// This type exists only to take advantage of the sql.Null type, which I'm using
	// mostly out of caution rather than understanding a real threat of error. NOT
	// NULL is on every column in the database. integrity.
	type StatusDAO struct {
		SerNo                  sql.NullInt32
		SeqNo                  sql.NullInt32
		Reason                 sql.NullInt32
		Latitude               sql.NullFloat64
		Longitude              sql.NullFloat64
		Altitude               sql.NullInt32
		Speed                  sql.NullInt32
		DateUTC                model.Time
		GpsUTC                 model.Time
		PosAcc                 sql.NullInt32
		GpsStatus              sql.NullInt32
		InternalBatteryVoltage sql.NullInt32
	}

	isValid := func(pr StatusDAO) bool {
		// seqNo checked elsewhere so error message can use it.
		return (pr.SerNo.Valid &&
			pr.Reason.Valid &&
			pr.Latitude.Valid &&
			pr.Longitude.Valid &&
			pr.Altitude.Valid &&
			pr.Speed.Valid &&
			pr.PosAcc.Valid &&
			pr.GpsStatus.Valid &&
			pr.InternalBatteryVoltage.Valid)
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

	ss := make(Statuses)

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return Statuses{}, fmt.Errorf("error querying database for last statuses: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var sDAO StatusDAO
		// TODO: try https://github.com/jmoiron/sqlx for direct scanning into
		// struct.
		err = rows.Scan(
			&sDAO.SerNo,
			&sDAO.SeqNo,
			&sDAO.Reason,
			&sDAO.Latitude,
			&sDAO.Longitude,
			&sDAO.Altitude,
			&sDAO.Speed,
			&sDAO.DateUTC,
			&sDAO.GpsUTC,
			&sDAO.PosAcc,
			&sDAO.GpsStatus,
			&sDAO.InternalBatteryVoltage)
		if err != nil {
			return Statuses{}, fmt.Errorf("error scanning row: %w", err)
		}
		if !sDAO.SeqNo.Valid {
			return Statuses{}, errors.New("SeqNo in record is NULL")
		}
		if !isValid(sDAO) {
			return Statuses{}, fmt.Errorf("one of the fields of database row for SeqNo %v is NULL", sDAO.SeqNo.Int32)
		}

		ss[sDAO.SerNo.Int32] = Status{
			SeqNo:     sDAO.SeqNo.Int32,
			Reason:    model.ReasonCode(sDAO.Reason.Int32),
			Latitude:  sDAO.Latitude.Float64,
			Longitude: sDAO.Longitude.Float64,
			Altitude:  sDAO.Altitude.Int32,
			Speed:     sDAO.Speed.Int32,
			DateUTC:   sDAO.DateUTC.T,
			GpsUTC:    sDAO.GpsUTC.T,
			PosAcc:    sDAO.PosAcc.Int32,
			GpsStatus: sDAO.GpsStatus.Int32,
			Battery:   sDAO.InternalBatteryVoltage.Int32,
		}
	}

	err = rows.Err()
	if err != nil {
		return Statuses{}, fmt.Errorf("error after scanning rows: %w", err)
	}

	return ss, nil
}

func (s SqliteStorer) GetLastNCoords(ctx context.Context, n int) (Coords, error) {
	type pointRecordDAO struct {
		SerNo     sql.NullInt32
		SeqNo     sql.NullInt32
		Latitude  sql.NullFloat64
		Longitude sql.NullFloat64
	}

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

	rows, err := s.db.QueryContext(ctx, query, n)
	if err != nil {
		return Coords{}, fmt.Errorf("error querying database for last N coordinates: %w", err)
	}
	defer rows.Close()

	ppr := make(Coords)

	for rows.Next() {
		var prDAO pointRecordDAO
		err := rows.Scan(
			&prDAO.SerNo,
			&prDAO.SeqNo,
			&prDAO.Latitude,
			&prDAO.Longitude,
		)
		if err != nil {
			return Coords{}, fmt.Errorf("error scanning row: %w", err)
		}
		if !prDAO.SeqNo.Valid {
			return Coords{}, errors.New("SeqNo in record is NULL")
		}
		if !prDAO.Latitude.Valid || !prDAO.Longitude.Valid {
			return Coords{}, fmt.Errorf("one of the fields of database row for SeqNo %v is NULL", prDAO.SeqNo.Int32)
		}

		ppr[prDAO.SerNo.Int32] = append(ppr[prDAO.SerNo.Int32], Coord{
			Latitude:  prDAO.Latitude.Float64,
			Longitude: prDAO.Longitude.Float64,
		})
	}

	err = rows.Err()
	if err != nil {
		return Coords{}, fmt.Errorf("error after scanning rows: %w", err)
	}

	return ppr, nil
}
