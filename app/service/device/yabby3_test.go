package device

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed sample.json
var sampleJSON []byte

func TestUnmarshals(t *testing.T) {
	require.NotEmpty(t, sampleJSON)
	var tx TagTx
	err := json.Unmarshal(sampleJSON, &tx)
	require.Nil(t, err)

	// Not strictly necessary because of require.Equal below, but fail early.
	require.Len(t, tx.Records, 2, "number of records")
	require.NotNil(t, tx.Records[0].GPSReading, "GPSReading defined in record [0]")
	require.NotNil(t, tx.Records[0].GPIOReading, "GPIOReading defined in record [0]")
	require.NotNil(t, tx.Records[0].AnalogueReading, "AnalogueReading defined in record [0]")
	require.Nil(t, tx.Records[0].TripTypeReading, "TripTypeReading not defined in record [0]")
	require.NotNil(t, tx.Records[1].GPSReading, "GPSReading defined in record [1]")
	require.NotNil(t, tx.Records[1].GPIOReading, "GPIOReading defined in record [1]")
	require.NotNil(t, tx.Records[1].AnalogueReading, "AnalogueReading defined in record [1]")
	require.NotNil(t, tx.Records[1].TripTypeReading, "TripTypeReading defined in record [1]")

	expected := TagTx{
		SerNo:  810095,
		IMEI:   "353785725680796",
		ICCID:  "89610180004127201829",
		ProdID: 97,
		Fw:     "97.2.1.11",
		Records: []Record{
			{
				SeqNo:   7494,
				Reason:  11,
				DateUTC: "2023-10-21 23:21:42",
				GPSReading: &GPSReading{
					GpsUTC:  "2023-10-21 23:17:40",
					Lat:     -31.4577084,
					Long:    152.64215,
					Alt:     35,
					Spd:     0,
					SpdAcc:  2,
					Head:    0,
					PDOP:    17,
					PosAcc:  10,
					GpsStat: 7,
					FType:   0,
				},
				GPIOReading: &GPIOReading{
					DIn:     1,
					DOut:    0,
					DevStat: 1,
					FType:   2,
				},
				AnalogueReading: &AnalogueReading{
					AnalogueData: AnalogueData{
						Num1: 4641,
						Num3: 3500,
						Num4: 8,
						Num5: 4500,
					},
					FType: 6,
				},
			},
			{
				SeqNo:   7495,
				Reason:  2,
				DateUTC: "2023-10-21 23:23:36",
				GPSReading: &GPSReading{
					GpsUTC:  "2023-10-21 23:17:40",
					Lat:     -31.4577084,
					Long:    152.64215,
					Alt:     35,
					Spd:     0,
					SpdAcc:  2,
					Head:    0,
					PDOP:    17,
					PosAcc:  10,
					GpsStat: 7,
					FType:   0,
				},
				TripTypeReading: &TripTypeReading{
					Tt:    2,
					Trim:  300,
					FType: 15,
				},
				GPIOReading: &GPIOReading{
					DIn:     0,
					DOut:    0,
					DevStat: 0,
					FType:   2,
				},
				AnalogueReading: &AnalogueReading{
					AnalogueData: AnalogueData{
						Num1: 4641,
						Num3: 3400,
						Num4: 8,
						Num5: 4504,
					},
					FType: 6,
				},
			},
		},
	}

	require.Equal(t, expected, tx, "fully unmarshalled value")
}
