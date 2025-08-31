package storage

import (
	"context"
	"os"

	"testing"

	"github.com/bitwombat/gps-tags/device"
	"github.com/stretchr/testify/require"
	"maragu.dev/migrate"
)

var sampleTx1 device.TagTx = device.TagTx{
	SerNo:  810095,
	IMEI:   "353785725680796",
	ICCID:  "89610180004127201829",
	ProdID: 97,
	Fw:     "97.2.1.11",
	Records: []device.Record{
		{
			SeqNo:   7494,
			Reason:  11,
			DateUTC: "2023-10-21 23:21:42",
			GPSReading: &device.GPSReading{
				GpsUTC:  "2023-10-21 23:17:40",
				Lat:     -31.4577084,
				Long:    152.64215,
				Alt:     35,
				Spd:     1,
				SpdAcc:  2,
				Head:    0,
				PDOP:    17,
				PosAcc:  10,
				GpsStat: 7,
				FType:   0,
			},
			GPIOReading: &device.GPIOReading{
				DIn:     1,
				DOut:    0,
				DevStat: 1,
				FType:   2,
			},
			AnalogueReading: &device.AnalogueReading{
				AnalogueData: device.AnalogueData{
					Num1: 4641,
					Num3: 3500,
					Num4: 8,
					Num5: 4500,
				},
				FType: 6,
			},
			TripTypeReading: &device.TripTypeReading{
				Tt:    2,
				Trim:  300,
				FType: 15,
			},
		},
		{
			SeqNo:   7495,
			Reason:  2,
			DateUTC: "2023-10-21 23:23:36",
			GPSReading: &device.GPSReading{
				GpsUTC:  "2023-10-21 23:17:40",
				Lat:     -31.4577084,
				Long:    152.64215,
				Alt:     35,
				Spd:     5,
				SpdAcc:  2,
				Head:    0,
				PDOP:    17,
				PosAcc:  10,
				GpsStat: 7,
				FType:   0,
			},
			TripTypeReading: &device.TripTypeReading{
				Tt:    2,
				Trim:  300,
				FType: 15,
			},
			GPIOReading: &device.GPIOReading{
				DIn:     0,
				DOut:    0,
				DevStat: 0,
				FType:   2,
			},
			AnalogueReading: &device.AnalogueReading{
				AnalogueData: device.AnalogueData{
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

var sampleTx2 device.TagTx = device.TagTx{
	SerNo:  810243,
	IMEI:   "353785725680796",
	ICCID:  "89610180004127201829",
	ProdID: 97,
	Fw:     "97.2.1.11",
	Records: []device.Record{
		{
			SeqNo:   7496,
			Reason:  11,
			DateUTC: "2023-10-21 23:21:42",
			GPSReading: &device.GPSReading{
				GpsUTC:  "2023-10-21 23:17:40",
				Lat:     -31.4577084,
				Long:    152.64215,
				Alt:     35,
				Spd:     11,
				SpdAcc:  2,
				Head:    0,
				PDOP:    17,
				PosAcc:  10,
				GpsStat: 7,
				FType:   0,
			},
			GPIOReading: &device.GPIOReading{
				DIn:     1,
				DOut:    0,
				DevStat: 1,
				FType:   2,
			},
			AnalogueReading: &device.AnalogueReading{
				AnalogueData: device.AnalogueData{
					Num1: 4641,
					Num3: 3500,
					Num4: 8,
					Num5: 4500,
				},
				FType: 6,
			},
			TripTypeReading: &device.TripTypeReading{
				Tt:    2,
				Trim:  300,
				FType: 15,
			},
		},
		{
			SeqNo:   7497,
			Reason:  2,
			DateUTC: "2023-10-21 23:23:36",
			GPSReading: &device.GPSReading{
				GpsUTC:  "2023-10-21 23:17:40",
				Lat:     -31.99,
				Long:    152.99,
				Alt:     35,
				Spd:     13,
				SpdAcc:  2,
				Head:    0,
				PDOP:    17,
				PosAcc:  10,
				GpsStat: 7,
				FType:   0,
			},
			TripTypeReading: &device.TripTypeReading{
				Tt:    2,
				Trim:  300,
				FType: 15,
			},
			GPIOReading: &device.GPIOReading{
				DIn:     0,
				DOut:    0,
				DevStat: 0,
				FType:   2,
			},
			AnalogueReading: &device.AnalogueReading{
				AnalogueData: device.AnalogueData{
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

func TestGetLatestPosition(t *testing.T) {
	// GIVEN a two transmissions for two tags with multiple records
	storer, err := NewSQLiteStorer(":memory:")
	require.Nil(t, err)

	var migrations = os.DirFS("../migrations")

	err = migrate.Up(context.Background(), storer.db, migrations)
	require.Nil(t, err)

	_, err = storer.WriteTx(context.Background(), []device.TagTx{sampleTx1, sampleTx2})

	// WHEN we get the latest position for all tags.
	result, err := storer.GetLastPositions(context.Background())
	require.Nil(t, err)

	// THEN we get the latest position's values for both known tags.
	require.Len(t, result, 2, "length of result array")
	for _, r := range result {
		switch r.SerNo {
		case 810095:
			require.Equal(t, int32(7495), r.SeqNo)
			require.Equal(t, -31.4577084, r.Latitude)
			require.Equal(t, 152.64215, r.Longitude)
			require.Equal(t, int32(35), r.Altitude)
			require.Equal(t, int32(5), r.Speed)
			require.Equal(t, int32(10), r.PosAcc)
			require.Equal(t, int32(7), r.GpsStatus)
			require.Equal(t, "2023-10-21 23:17:40", r.GpsUTC)
		case 810243:
			require.Equal(t, int32(7497), r.SeqNo)
			require.Equal(t, -31.99, r.Latitude)
			require.Equal(t, 152.99, r.Longitude)
		default:
			t.Fatalf("Unmatched serNo: %v", r.SerNo)
		}
	}

}

// func TestGetLastNPositions(t *testing.T) {

// 	// GIVEN commits with multiple records and for multiple tags.
// 	collection, err := NewMongoConnection(mongoURL, randomTestCollectionName())
// 	require.Nil(t, err)

// 	storer := NewMongoStorer(collection)

// 	for _, r := range nSamples {
// 		if s, ok := r.(string); ok {
// 			_, err = storer.WriteTx(context.Background(), s)
// 			require.Nil(t, err)
// 		} else {
// 			t.Fatal("type assertion failed")
// 		}
// 	}

// 	// WHEN we get the last 3 position for all tags.
// 	result, err := storer.GetLastNPositions(3)
// 	require.Nil(t, err)

// 	// THEN we get the latest position's values for both known tags.
// 	for _, r := range result {
// 		switch r.SerNo {
// 		case 810095:
// 			require.Equal(t, 108.0, r.PathPoints[0].Latitude)
// 			require.Equal(t, 109.0, r.PathPoints[0].Longitude)
// 			require.Equal(t, 106.0, r.PathPoints[1].Latitude)
// 			require.Equal(t, 107.0, r.PathPoints[1].Longitude)
// 			require.Equal(t, 104.0, r.PathPoints[2].Latitude)
// 			require.Equal(t, 105.0, r.PathPoints[2].Longitude)
// 		case 810243:
// 			require.Equal(t, 118.0, r.PathPoints[0].Latitude)
// 			require.Equal(t, 119.0, r.PathPoints[0].Longitude)
// 			require.Equal(t, 116.0, r.PathPoints[1].Latitude)
// 			require.Equal(t, 117.0, r.PathPoints[1].Longitude)
// 			require.Equal(t, 114.0, r.PathPoints[2].Latitude)
// 			require.Equal(t, 115.0, r.PathPoints[2].Longitude)
// 		default:
// 			t.Fatalf("Unmatched serNo: %v", r.SerNo)
// 		}
// 	}

// }
