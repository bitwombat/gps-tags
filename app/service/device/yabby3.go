package device

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bitwombat/gps-tags/model"
)

// TODO: do these field names need to be exported?

// These types are close to what comes from the Yabby except:
// - []Fields are flattened up to the Record
// - Fields are pointers to types because they're optional. This is so the
// icky/weird JSON format doesn't affect the model and business logic. But, see
// the TODO below: this probably isn't the right place to do this.
//
// I can only get away with this because of the irregular JSON decoding that
// gives me a place to insert logic and make the Reading* elements as I want
// them (pointers).
//
// TODO: Maybe? It would be better if Record had []Fields, which were of type
// any, like I had in a previous version. Then this file would be straight JSON
// decoding (though with the irregular JSON fun). Then the convert function here
// would flatten the structure and use the pointers. This would have the
// responsibilities more accurately contained/partitioned.

type TagTx struct {
	ID      string
	SerNo   int      `json:"SerNo"`
	IMEI    string   `json:"IMEI"`
	ICCID   string   `json:"ICCID"`
	ProdID  int      `json:"ProdId"`
	Fw      string   `json:"FW"`
	Records []Record `json:"Records"`
}

type Record struct {
	ID              string
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

// TODO: Do we ever get multiple transmissions? See if the raw data from the
// yabby, which I think I'm logging, starts with an JSON array [.
// Make sure my logging (main branch) isn't JSON decoding first!

func Unmarshal(jsonBytes []byte) (model.TagTx, error) {
	var deviceData TagTx

	// Use a Decoder so that we can set the DisallowUnknownFields and be
	// stricter in our JSON decoding.
	// However, during testing, this didn't seem to work. Adding or removing
	// fields from handledata_test.go did not cause errors.
	decoder := json.NewDecoder(strings.NewReader(string(jsonBytes)))
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&deviceData)
	if err != nil {
		return model.TagTx{}, fmt.Errorf("error unmarshalling JSON: %w", err)
	}

	return ModelFrom(deviceData), nil
}

// These convert functions belong here so that the device knows about the model
// (imports the package), but the model (higher up, nearer the business logic)
// doesn't have to know about the device.
func ModelFrom(d TagTx) model.TagTx {
	var m model.TagTx
	m.ID = d.ID
	m.SerNo = d.SerNo
	m.IMEI = d.IMEI
	m.ICCID = d.ICCID
	m.ProdID = d.ProdID // TODO: Fix Id case
	m.Fw = d.Fw

	var mrs []model.Record
	for _, r := range d.Records {
		mr := convertRecord(r)
		mrs = append(mrs, mr)
	}

	m.Records = mrs

	return m
}

func convertRecord(r Record) model.Record {
	var mr model.Record
	mr.ID = r.ID
	mr.SeqNo = r.SeqNo
	mr.Reason = r.Reason

	t, err := model.TimeFromString(r.DateUTC)
	if err != nil {
		panic(fmt.Sprintf("failed scanning time string %s: %v", r.DateUTC, err))
	}
	mr.DateUTC = t

	mr.GPSReading = convertGPSReading(r.GPSReading)
	mr.GPIOReading = convertGPIOReading(r.GPIOReading)
	mr.AnalogueReading = convertAnalogueReading(r.AnalogueReading)
	mr.TripTypeReading = convertTripTypeReading(r.TripTypeReading)

	return mr
}

func convertGPSReading(gr *GPSReading) *model.GPSReading {
	if gr == nil {
		return nil
	}
	var mr model.GPSReading
	mr.Spd = gr.Spd
	mr.SpdAcc = gr.SpdAcc
	mr.Head = gr.Head
	mr.GpsStat = gr.GpsStat

	t, err := model.TimeFromString(gr.GpsUTC)
	if err != nil {
		panic(fmt.Sprintf("failed scanning time string %s: %v", gr.GpsUTC, err))
	}
	mr.GpsUTC = t

	mr.Lat = gr.Lat
	mr.Long = gr.Long
	mr.Alt = gr.Alt
	mr.PosAcc = gr.PosAcc
	mr.PDOP = gr.PDOP

	return &mr
}

func convertGPIOReading(gr *GPIOReading) *model.GPIOReading {
	if gr == nil {
		return nil
	}
	var mr model.GPIOReading
	mr.DIn = gr.DIn
	mr.DOut = gr.DOut
	mr.DevStat = gr.DevStat

	return &mr
}

func convertAnalogueReading(ar *AnalogueReading) *model.AnalogueReading {
	if ar == nil {
		return nil
	}
	var mr model.AnalogueReading
	mr.InternalBatteryVoltage = ar.AnalogueData.Num1
	mr.Temperature = ar.AnalogueData.Num3
	mr.LastGSMCQ = ar.AnalogueData.Num4
	mr.LoadedVoltage = ar.AnalogueData.Num5

	return &mr
}

func convertTripTypeReading(tt *TripTypeReading) *model.TripTypeReading {
	if tt == nil {
		return nil
	}
	var mr model.TripTypeReading
	mr.Trim = tt.Trim
	mr.Tt = tt.Tt

	return &mr
}

// TODO: Is this used? (unused checker doesn't always catch these).
var SerNoToName = map[int]string{
	810095: "rueger",
	810243: "charlie",
}
