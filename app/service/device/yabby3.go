package device

import (
	"encoding/json"
	"fmt"
)

type TagTx struct {
	SerNo   int      `json:"SerNo"`
	Imei    string   `json:"IMEI"`
	Iccid   string   `json:"ICCID"`
	ProdID  int      `json:"ProdId"`
	Fw      string   `json:"FW"`
	Records []Record `json:"Records"`
}

type Record struct {
	SeqNo   int     `json:"SeqNo"`
	Reason  int     `json:"Reason"`
	DateUTC string  `json:"DateUTC"`
	Fields  []Field `json:"Fields"`
}

type Field any

type FType0 struct {
	Spd     int     `json:"Spd"`
	SpdAcc  int     `json:"SpdAcc"`
	Head    int     `json:"Head"`
	GpsStat int     `json:"GpsStat"`
	GpsUTC  string  `json:"GpsUTC"`
	Lat     float64 `json:"Lat"`
	Long    float64 `json:"Long"`
	Alt     int     `json:"Alt"`
	FType   int     `json:"FType"`
	PosAcc  int     `json:"PosAcc"`
	Pdop    int     `json:"PDOP"`
}

type FType2 struct {
	DIn     int `json:"DIn"`
	DOut    int `json:"DOut"`
	DevStat int `json:"DevStat"`
	FType   int `json:"FType"`
}

type FType6 struct {
	AnalogueData AnalogueData `json:"AnalogueData"`
	FType        int          `json:"FType"`
}

type AnalogueData struct {
	Num1 int `json:"1"`
	Num3 int `json:"3"`
	Num4 int `json:"4"`
	Num5 int `json:"5"`
}

type FType15 struct {
	FType int `json:"FType"`
	Tt    int `json:"TT"`
	Trim  int `json:"Trim"`
}

func (r *Record) UnmarshalJSON(p []byte) error {
	// This code is necessary because the type of the Fields varies.
	// Unmarshal the regular parts of the JSON value
	var rawRecord struct {
		DateUTC string            `json:"DateUTC"`
		Fields  []json.RawMessage `json:"Fields"`
		SeqNo   int               `json:"SeqNo"`
		Reason  int               `json:"Reason"`
	}

	if err := json.Unmarshal(p, &rawRecord); err != nil {
		return fmt.Errorf("custom unmarshaling Record: %w", err)
	}

	r.DateUTC = rawRecord.DateUTC
	r.SeqNo = rawRecord.SeqNo
	r.Reason = rawRecord.Reason

	for _, rawField := range rawRecord.Fields {
		fType := struct {
			FType int `json:"FType"`
		}{}
		if err := json.Unmarshal(rawField, &fType); err != nil {
			return fmt.Errorf("unmarshaling the raw field %v: %w", rawField, err)
		}

		switch fType.FType {
		case 0:
			var ft0 FType0
			if err := json.Unmarshal(rawField, &ft0); err != nil {
				return fmt.Errorf("unmarshaling the field %v: %w", rawField, err)
			}
			r.Fields = append(r.Fields, ft0)
		case 2:
			var ft2 FType2
			if err := json.Unmarshal(rawField, &ft2); err != nil {
				return fmt.Errorf("unmarshaling the field %v: %w", rawField, err)
			}
			r.Fields = append(r.Fields, ft2)
		case 6:
			var ft6 FType6
			if err := json.Unmarshal(rawField, &ft6); err != nil {
				return fmt.Errorf("unmarshaling the field %v: %w", rawField, err)
			}
			r.Fields = append(r.Fields, ft6)
		case 15:
			var ft15 FType15
			if err := json.Unmarshal(rawField, &ft15); err != nil {
				return fmt.Errorf("unmarshaling the field %v: %w", rawField, err)
			}
			r.Fields = append(r.Fields, ft15)
		default:
			return fmt.Errorf("unrecognised FType: %v", fType.FType)
		}
	}

	return nil
}

const AnalogueDataFType = 6

var IdToName = map[float64]string{
	810095: "rueger",
	810243: "charlie",
}
