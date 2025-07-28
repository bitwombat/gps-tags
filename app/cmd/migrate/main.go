package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type TagJSON struct {
	ID     ID       `json:"_id"`
	ProdID float64  `json:"ProdId"`
	Fw     string   `json:"FW"`
	Record []Record `json:"Records"`
	SerNo  float64  `json:"SerNo"`
	Imei   string   `json:"IMEI"`
	Iccid  string   `json:"ICCID"`
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
	AnalogueData AnalogueData `json:"AnalogueData,omitempty"`
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
			return fmt.Errorf("unmarshalling the raw field %v: %w", rawField, err)
		}

		switch obj.FType {
		case 0.0:
			var ft0 FType0
			if err := json.Unmarshal(rawField, &ft0); err != nil {
				return fmt.Errorf("unmarshalling the field %v: %w", rawField, err)
			}
			r.Fields = append(r.Fields, ft0)
		case 2.0:
			var ft2 FType2
			if err := json.Unmarshal(rawField, &ft2); err != nil {
				return fmt.Errorf("unmarshalling the field %v: %w", rawField, err)
			}
			r.Fields = append(r.Fields, ft2)
		case 6.0:
			var ft6 FType6
			if err := json.Unmarshal(rawField, &ft6); err != nil {
				return fmt.Errorf("unmarshalling the field %v: %w", rawField, err)
			}
			r.Fields = append(r.Fields, ft6)
		case 15.0:
			var ft15 FType15
			if err := json.Unmarshal(rawField, &ft15); err != nil {
				return fmt.Errorf("unmarshalling the field %v: %w", rawField, err)
			}
			r.Fields = append(r.Fields, ft15)
		default:
			return fmt.Errorf("unreconised FType: %f", obj.FType)
		}
	}

	return nil
}

func pretty(o any) {
	var b []byte
	b, _ = json.MarshalIndent(o, "", "  ")
	os.Stdout.Write(b)
	fmt.Println()
}

func main() {
	// jsonData, err := os.ReadFile("dogs_stripped_down_all_reasons.json")
	jsonData, err := os.ReadFile("dogs.json")
	if err != nil {
		fmt.Printf("reading input json file: %v\n", err)
		os.Exit(1)
	}

	var data []TagJSON

	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		fmt.Printf("unmarshalling JSON: %v\n", err)
		panic(err)
	}
	//fmt.Printf("%v\n", data)
	pretty(data)
}
