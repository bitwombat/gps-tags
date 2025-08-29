package device

import (
	"encoding/json"
	"fmt"
)

// TODO: do these field names need to be exported?

// These types are a mix of JSON decoding types (DAOs?) and models. The
// difference between the two (besides the JSON tags) is that I flatten the
// []Fields from the device, with their optional members, into pointers to types
// (thus optional). This works better with Go than having to have switch.(type)
// statements in multiple places. Don't let the device's design leak into the
// business logic.
// I can only get away with this because of the irregular JSON decoding that
// gives me a place to insert logic and make the Reading* elements as I want
// them (pointers).
type TagTx struct {
	SerNo   int      `json:"SerNo"`
	IMEI    string   `json:"IMEI"`
	ICCID   string   `json:"ICCID"`
	ProdID  int      `json:"ProdId"`
	Fw      string   `json:"FW"`
	Records []Record `json:"Records"`
}

type Record struct {
	SeqNo           int    `json:"SeqNo"`
	Reason          int    `json:"Reason"`
	DateUTC         string `json:"DateUTC"`
	GPSReading      *GPSReading
	GPIOReading     *GPIOReading
	AnalogueReading *AnalogueReading
	TripTypeReading *TripTypeReading
}

type GPSReading struct { // FType0
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
	PDOP    int     `json:"PDOP"`
}

type GPIOReading struct { // FType2
	DIn     int `json:"DIn"`
	DOut    int `json:"DOut"`
	DevStat int `json:"DevStat"`
	FType   int `json:"FType"`
}

type AnalogueReading struct { // FType6
	AnalogueData AnalogueData `json:"AnalogueData"`
	FType        int          `json:"FType"`
}

type AnalogueData struct {
	Num1 int `json:"1"`
	Num3 int `json:"3"`
	Num4 int `json:"4"`
	Num5 int `json:"5"`
}

type TripTypeReading struct { // FType15
	FType int `json:"FType"`
	Tt    int `json:"TT"`
	Trim  int `json:"Trim"`
}

func (r *Record) UnmarshalJSON(p []byte) error {
	// The type of the Fields varies.
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
			var f *GPSReading
			if err := json.Unmarshal(rawField, &f); err != nil {
				return fmt.Errorf("unmarshaling the field %v: %w", rawField, err)
			}
			r.GPSReading = f
		case 2:
			var f *GPIOReading
			if err := json.Unmarshal(rawField, &f); err != nil {
				return fmt.Errorf("unmarshaling the field %v: %w", rawField, err)
			}
			r.GPIOReading = f
		case 6:
			var f *AnalogueReading
			if err := json.Unmarshal(rawField, &f); err != nil {
				return fmt.Errorf("unmarshaling the field %v: %w", rawField, err)
			}
			r.AnalogueReading = f
		case 15:
			var f *TripTypeReading
			if err := json.Unmarshal(rawField, &f); err != nil {
				return fmt.Errorf("unmarshaling the field %v: %w", rawField, err)
			}
			r.TripTypeReading = f
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
