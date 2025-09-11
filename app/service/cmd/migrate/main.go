package main

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"maragu.dev/migrate"
	_ "modernc.org/sqlite"
)

const dataSourceName = "dogtags.db"

// TxsMongo and friends are for reading the MongoDB export file, which is JSON.
type TxsMongo []TxMongo

type TxMongo struct {
	ID      IDMongo       `json:"_id"`
	ProdID  float64       `json:"ProdId"`
	Fw      string        `json:"FW"`
	Records []RecordMongo `json:"Records"`
	SerNo   float64       `json:"SerNo"`
	Imei    string        `json:"IMEI"`
	Iccid   string        `json:"ICCID"`
}

type IDMongo struct {
	Oid string `json:"$oid"`
}

type RecordMongo struct {
	DateUTC string       `json:"DateUTC"`
	Fields  []FieldMongo `json:"Fields"`
	SeqNo   float64      `json:"SeqNo"`
	Reason  float64      `json:"Reason"`
}

type FieldMongo any

type FType0Mongo struct {
	Spd     float64 `json:"Spd,omitempty"`
	SpdAcc  float64 `json:"SpdAcc,omitempty"`
	Head    float64 `json:"Head,omitempty"`
	GpsStat float64 `json:"GpsStat,omitempty"`
	GpsUTC  string  `json:"GpsUTC,omitempty"`
	Lat     float64 `json:"Lat,omitempty"`
	Long    float64 `json:"Long,omitempty"`
	Alt     float64 `json:"Alt,omitempty"`
	FType   float64 `json:"FType"`
	PosAcc  float64 `json:"PosAcc,omitempty"`
	Pdop    float64 `json:"PDOP,omitempty"`
}

type FType2Mongo struct {
	DIn     float64 `json:"DIn,omitempty"`
	DOut    float64 `json:"DOut,omitempty"`
	DevStat float64 `json:"DevStat,omitempty"`
	FType   float64 `json:"FType"`
}

type FType6Mongo struct {
	AnalogueData AnalogueDataMongo `json:"AnalogueData"`
	FType        float64           `json:"FType"`
}

type AnalogueDataMongo struct {
	Num1 float64 `json:"1"`
	Num3 float64 `json:"3"`
	Num4 float64 `json:"4"`
	Num5 float64 `json:"5"`
}

type FType15Mongo struct {
	FType float64 `json:"FType"`
	Tt    float64 `json:"TT"`
	Trim  float64 `json:"Trim"`
}

func (r *RecordMongo) UnmarshalJSON(p []byte) error {
	// This code is necessary because the type of the Fields varies.
	// Unmarshal the regular parts of the JSON value
	var rawRecord struct {
		DateUTC string            `json:"DateUTC"`
		Fields  []json.RawMessage `json:"Fields"`
		SeqNo   float64           `json:"SeqNo"`
		Reason  float64           `json:"Reason"`
	}

	if err := json.Unmarshal(p, &rawRecord); err != nil {
		return fmt.Errorf("custom unmarshaling Record: %w", err)
	}

	r.DateUTC = rawRecord.DateUTC
	r.SeqNo = rawRecord.SeqNo
	r.Reason = rawRecord.Reason

	for _, rawField := range rawRecord.Fields {
		obj := struct {
			FType float64 `json:"FType"`
		}{}
		if err := json.Unmarshal(rawField, &obj); err != nil {
			return fmt.Errorf("unmarshaling the raw field %v: %w", rawField, err)
		}

		switch obj.FType {
		case 0.0:
			var ft0 FType0Mongo
			if err := json.Unmarshal(rawField, &ft0); err != nil {
				return fmt.Errorf("unmarshaling the field %v: %w", rawField, err)
			}
			r.Fields = append(r.Fields, ft0)
		case 2.0:
			var ft2 FType2Mongo
			if err := json.Unmarshal(rawField, &ft2); err != nil {
				return fmt.Errorf("unmarshaling the field %v: %w", rawField, err)
			}
			r.Fields = append(r.Fields, ft2)
		case 6.0:
			var ft6 FType6Mongo
			if err := json.Unmarshal(rawField, &ft6); err != nil {
				return fmt.Errorf("unmarshaling the field %v: %w", rawField, err)
			}
			r.Fields = append(r.Fields, ft6)
		case 15.0:
			var ft15 FType15Mongo
			if err := json.Unmarshal(rawField, &ft15); err != nil {
				return fmt.Errorf("unmarshaling the field %v: %w", rawField, err)
			}
			r.Fields = append(r.Fields, ft15)
		default:
			return fmt.Errorf("unrecognised FType: %f", obj.FType)
		}
	}

	return nil
}

func (txs TxsMongo) convert() (TagTxs, error) {
	var tts TagTxs
	for _, tx := range txs {
		tt, err := tx.convert()
		if err != nil {
			return nil, fmt.Errorf("converting tag %v: %w", tx, err)
		}
		tts = append(tts, tt)
	}

	return tts, nil
}

func (i TxMongo) convert() (TagTx, error) {
	var o TagTx

	o.ID = i.ID.Oid
	o.ProdID = int(i.ProdID)
	o.Fw = i.Fw
	o.SerNo = int(i.SerNo)
	o.Imei = i.Imei
	o.Iccid = i.Iccid
	var err error
	o.Records, err = convertRecords(i.Records)
	if err != nil {
		return TagTx{}, fmt.Errorf("converting records: %w", err)
	}

	return o, nil
}

// Most of these convert functions are just typecasting.
func convertRecords(i []RecordMongo) ([]Record, error) {
	o := make([]Record, len(i))
	for k, r := range i {

		parsedT, err := time.Parse(time.DateTime, r.DateUTC)
		if err != nil {
			return nil, fmt.Errorf("parsing time %s: %w", r.DateUTC, err)
		}

		o[k].DateUTC = Time{T: parsedT}
		o[k].SeqNo = int(r.SeqNo)
		o[k].Reason = int(r.Reason)
		o[k].Fields, err = convertFields(r.Fields)
		if err != nil {
			return nil, fmt.Errorf("converting fields: %w", err)
		}
	}

	return o, nil
}

func convertFields(i []FieldMongo) ([]Field, error) {
	var o []Field
	for _, f := range i {
		switch ft := f.(type) {
		case FType0Mongo:
			var nf GPSReading
			parsedT, err := time.Parse(time.DateTime, ft.GpsUTC)
			if err != nil {
				return nil, fmt.Errorf("parsing time %s: %w", ft.GpsUTC, err)
			}

			nf.Spd = int(ft.Spd)
			nf.SpdAcc = int(ft.SpdAcc)
			nf.Head = int(ft.Head)
			nf.GpsStat = int(ft.GpsStat)
			nf.GpsUTC = Time{T: parsedT}
			nf.Lat = ft.Lat
			nf.Long = ft.Long
			nf.Alt = int(ft.Alt)
			nf.PosAcc = int(ft.PosAcc)
			nf.Pdop = int(ft.Pdop)
			o = append(o, nf)

		case FType2Mongo:
			var nf GPIOReading
			nf.DIn = int(ft.DIn)
			nf.DOut = int(ft.DOut)
			nf.DevStat = int(ft.DevStat)
			o = append(o, nf)

		case FType6Mongo:
			var nf AnalogueReading
			nf.InternalBatteryVoltage = int(ft.AnalogueData.Num1)
			nf.Temperature = int(ft.AnalogueData.Num3)
			nf.LastGSMCQ = int(ft.AnalogueData.Num4)
			nf.LoadedVoltage = int(ft.AnalogueData.Num5)
			o = append(o, nf)

		case FType15Mongo:
			var nf TripTypeReading
			nf.Tt = int(ft.Tt)
			nf.Trim = int(ft.Trim)
			o = append(o, nf)

		default:
			return nil, errors.New("unknown field type")
		}
	}
	return o, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Need filename")
		os.Exit(1)
	}

	filename := os.Args[1]

	if filepath.Ext(filename) != ".json" {
		fmt.Println("Expecting a file ending in .json")
		os.Exit(1)
	}

	fmt.Println("Reading JSON")
	text, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("reading input json file: %v\n", err)
		os.Exit(1)
	}

	var hardwareTxs TxsMongo

	err = json.Unmarshal(text, &hardwareTxs)
	if err != nil {
		fmt.Printf("unmarshaling JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Read %v transmissions from JSON.\n", len(hardwareTxs))

	fmt.Println("Converting types")
	txs, err := hardwareTxs.convert()
	if err != nil {
		fmt.Printf("converting data to target types: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Converted %v JSON transmissions to own structure.\n", len(txs))

	outfilename := strings.TrimSuffix(filepath.Base(filename), ".json") + ".sql"

	f, err := os.Create(outfilename)
	if err != nil {
		fmt.Printf("opening file %s: %v\n", outfilename, err)
		os.Exit(1)
	}

	defer f.Close()

	w := bufio.NewWriterSize(f, 1<<16) // 64k buffer

	fmt.Println("Writing to", outfilename)

	for _, tx := range txs {
		sqlString := tx.ToSQL()

		_, err = w.WriteString(sqlString)
		if err != nil {
			fmt.Printf("error writing to %s: %v\n", outfilename, err)
			os.Exit(1) //nolint:gocritic // one-off CLI util, don't care
		}
	}

	w.Flush()

	fmt.Println("Creating the database and setting up schema")
	// Make the database while we're here.
	db, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		fmt.Println("error opening database: ", err)
		os.Exit(1)
	}
	migrations := os.DirFS("../../migrations")
	err = migrate.Up(context.Background(), db, migrations)
	if err != nil {
		fmt.Println("error migrating: ", err)
		os.Exit(1)
	}

	fmt.Println("Done.")
}
