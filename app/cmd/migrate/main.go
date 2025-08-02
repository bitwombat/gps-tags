package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/bitwombat/gps-tags-cmd/target"
)

type TxsJSON []TxJSON

type TxJSON struct {
	ID      ID       `json:"_id"`
	ProdID  float64  `json:"ProdId"`
	Fw      string   `json:"FW"`
	Records []Record `json:"Records"`
	SerNo   float64  `json:"SerNo"`
	Imei    string   `json:"IMEI"`
	Iccid   string   `json:"ICCID"`
}

type ID struct {
	Oid string `json:"$oid"`
}

type Record struct {
	DateUTC string  `json:"DateUTC"`
	Fields  []Field `json:"Fields"`
	SeqNo   float64 `json:"SeqNo"`
	Reason  float64 `json:"Reason"`
}

type Field any

type FType0 struct {
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

type FType2 struct {
	DIn     float64 `json:"DIn,omitempty"`
	DOut    float64 `json:"DOut,omitempty"`
	DevStat float64 `json:"DevStat,omitempty"`
	FType   float64 `json:"FType"`
}

type FType6 struct {
	AnalogueData AnalogueData `json:"AnalogueData"`
	FType        float64      `json:"FType"`
}

type AnalogueData struct {
	Num1 float64 `json:"1"`
	Num3 float64 `json:"3"`
	Num4 float64 `json:"4"`
	Num5 float64 `json:"5"`
}

type FType15 struct {
	FType float64 `json:"FType"`
	Tt    float64 `json:"TT"`
	Trim  float64 `json:"Trim"`
}

func (r *Record) UnmarshalJSON(p []byte) error {
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
			var ft0 FType0
			if err := json.Unmarshal(rawField, &ft0); err != nil {
				return fmt.Errorf("unmarshaling the field %v: %w", rawField, err)
			}
			r.Fields = append(r.Fields, ft0)
		case 2.0:
			var ft2 FType2
			if err := json.Unmarshal(rawField, &ft2); err != nil {
				return fmt.Errorf("unmarshaling the field %v: %w", rawField, err)
			}
			r.Fields = append(r.Fields, ft2)
		case 6.0:
			var ft6 FType6
			if err := json.Unmarshal(rawField, &ft6); err != nil {
				return fmt.Errorf("unmarshaling the field %v: %w", rawField, err)
			}
			r.Fields = append(r.Fields, ft6)
		case 15.0:
			var ft15 FType15
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

func (is TxsJSON) marshal() (target.Txs, error) {
	var tts target.Txs
	for _, i := range is {
		tt, err := i.marshal()
		if err != nil {
			return nil, fmt.Errorf("marshaling tag %v: %w", i, err)
		}
		tts = append(tts, tt)
	}

	return tts, nil
}

func (i TxJSON) marshal() (target.Tx, error) {
	var o target.Tx

	o.ID = i.ID.Oid
	o.ProdID = int(i.ProdID)
	o.Fw = i.Fw
	o.SerNo = int(i.SerNo)
	o.Imei = i.Imei
	o.Iccid = i.Iccid
	var err error
	o.Records, err = marshalRecords(i.Records)
	if err != nil {
		return target.Tx{}, fmt.Errorf("marshaling records: %w", err)
	}

	return o, nil
}

func marshalRecords(i []Record) ([]target.Record, error) {
	var o []target.Record = make([]target.Record, len(i))
	for k, r := range i {
		o[k].DateUTC = r.DateUTC
		o[k].SeqNo = int(r.SeqNo)
		o[k].Reason = int(r.Reason)
		var err error
		o[k].Fields, err = marshalFields(r.Fields)
		if err != nil {
			return nil, fmt.Errorf("marshaling fields: %w", err)
		}
	}

	return o, nil
}

func marshalFields(i []Field) ([]target.Field, error) {
	var o []target.Field
	for _, f := range i {
		switch ft := f.(type) {
		case FType0:
			var nf target.GPSReading
			nf.Spd = int(ft.Spd)
			nf.SpdAcc = int(ft.SpdAcc)
			nf.Head = int(ft.Head)
			nf.GpsStat = int(ft.GpsStat)
			nf.GpsUTC = ft.GpsUTC
			nf.Lat = ft.Lat
			nf.Long = ft.Long
			nf.Alt = int(ft.Alt)
			nf.PosAcc = int(ft.PosAcc)
			nf.Pdop = int(ft.Pdop)

		case FType2:
			var nf target.GPIOReading
			nf.DIn = int(ft.DIn)
			nf.DOut = int(ft.DOut)
			nf.DevStat = int(ft.DevStat)
			o = append(o, nf)

		case FType6:
			var nf target.AnalogueReading
			nf.InternalBatteryVoltage = int(ft.AnalogueData.Num1)
			nf.Temperature = int(ft.AnalogueData.Num3)
			nf.LastGSMCQ = int(ft.AnalogueData.Num4)
			nf.LoadedVoltage = int(ft.AnalogueData.Num5)
			o = append(o, nf)

		case FType15:
			var nf target.TripTypeReading
			nf.Tt = int(ft.Tt)
			nf.Trim = int(ft.Trim)
			o = append(o, nf)
		default:
			return nil, errors.New("unknown field type")
		}

	}
	return o, nil
}
func pretty(o any) {
	var b []byte
	b, _ = json.MarshalIndent(o, "", "  ")
	os.Stdout.Write(b)
	fmt.Println()
}

func main() {
	text, err := os.ReadFile("dogs.json")
	if err != nil {
		fmt.Printf("reading input json file: %v\n", err)
		os.Exit(1)
	}

	var hardwareTxs TxsJSON

	err = json.Unmarshal(text, &hardwareTxs)
	if err != nil {
		fmt.Printf("unmarshaling JSON: %v\n", err)
		panic(err)
	}

	pretty(hardwareTxs)

	txs, err := hardwareTxs.marshal()
	if err != nil {
		fmt.Printf("marshaling data to target types: %v\n", err)
		panic(err)
	}

	pretty(txs)
}
