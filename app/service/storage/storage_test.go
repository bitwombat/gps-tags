package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"encoding/json"
	"testing"

	"github.com/bitwombat/gps-tags/device"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const mongoURL = "mongodb://172.17.0.2:27017"

const basicCompleteSample = `{
  "SerNo": 810095,
  "IMEI": "353785725680796",
  "ICCID": "89610180004127201829",
  "ProdId": 97,
  "FW": "97.2.1.11",
  "Records": [
    {
      "SeqNo": 7494,
      "Reason": 11,
      "DateUTC": "2023-10-21 23:21:42",
      "Fields": [
        {
          "GpsUTC": "2023-10-21 23:17:40",
          "Lat": -31.4577084,
          "Long": 152.64215,
          "Alt": 35,
          "Spd": 0,
          "SpdAcc": 2,
          "Head": 0,
          "PDOP": 17,
          "PosAcc": 10,
          "GpsStat": 7,
          "FType": 0
        },
        {
          "DIn": 1,
          "DOut": 0,
          "DevStat": 1,
          "FType": 2
        },
        {
          "AnalogueData": {
            "1": 4641,
            "3": 3500,
            "4": 8,
            "5": 4500
          },
          "FType": 6
        }
      ]
    },
    {
      "SeqNo": 7495,
      "Reason": 2,
      "DateUTC": "2023-10-21 23:23:36",
      "Fields": [
        {
          "GpsUTC": "2023-10-21 23:17:40",
          "Lat": -31.4577084,
          "Long": 152.64215,
          "Alt": 35,
          "Spd": 0,
          "SpdAcc": 2,
          "Head": 0,
          "PDOP": 17,
          "PosAcc": 10,
          "GpsStat": 7,
          "FType": 0
        },
        {
          "TT": 2,
          "Trim": 300,
          "FType": 15
        },
        {
          "DIn": 0,
          "DOut": 0,
          "DevStat": 0,
          "FType": 2
        },
        {
          "AnalogueData": {
            "1": 4641,
            "3": 3400,
            "4": 8,
            "5": 4504
          },
          "FType": 6
        }
      ]
    }
  ]
}`

const strippedDownMultiRecordSample1 = `{
  "SerNo": 810095,
  "Records": [
    {
      "SeqNo": 7494,
      "Reason": 11,
      "DateUTC": "2023-10-21 23:21:42",
      "Fields": [
        {
          "GpsUTC": "2023-10-21 23:17:40",
          "Lat": -31.4577084,
          "Long": 152.64215
        }
      ]
    },
    {
      "SeqNo": 7495,
      "Reason": 2,
      "DateUTC": "2023-10-21 23:23:36",
      "Fields": [
        {
          "GpsUTC": "2023-10-21 23:17:40",
          "Lat": -31.4577084,
          "Long": 152.64215,
          "Alt": 35,
          "Spd": 1,
          "SpdAcc": 2,
          "Head": 3,
          "PDOP": 17,
          "PosAcc": 10,
          "GpsStat": 7,
          "FType": 0

        }
      ]
    }
  ]
}`

const strippedDownMultiRecordSample2 = `{
  "SerNo": 810243,
  "Records": [
    {
      "SeqNo": 7497,
      "Reason": 2,
      "DateUTC": "2023-10-21 23:23:37",
      "Fields": [
        {
          "GpsUTC": "2023-10-21 23:17:40",
          "Lat": -32.1,
          "Long": 153.1
        }
      ]
    },
    {
      "SeqNo": 7496,
      "Reason": 11,
      "DateUTC": "2023-10-21 23:21:42",
      "Fields": [
        {
          "GpsUTC": "2023-10-21 23:17:40",
          "Lat": -32.0,
          "Long": 153.0
        }
      ]
    }
  ]
}
`

func randomTestCollectionName() string {
	bytes := make([]byte, 8)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}

	return "test-" + hex.EncodeToString(bytes)
}

func Test_WriteCommit(t *testing.T) {
	collection, err := NewMongoConnection(mongoURL, randomTestCollectionName())
	require.Nil(t, err)

	// Unmarshal JSON to map
	var data map[string]interface{}
	err = json.Unmarshal([]byte(basicCompleteSample), &data)
	require.Nil(t, err)

	// Insert into MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	insertResult, err := collection.InsertOne(ctx, data)
	require.Nil(t, err)

	fmt.Println("Inserted document:", insertResult.InsertedID)

}

func Test_ReadCommit(t *testing.T) {
	collection, err := NewMongoConnection(mongoURL, randomTestCollectionName())
	require.Nil(t, err)

	// Unmarshal JSON to map
	var data map[string]interface{}
	err = json.Unmarshal([]byte(basicCompleteSample), &data)
	require.Nil(t, err)

	// Insert into MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	insertResult, err := collection.InsertOne(ctx, data)
	require.Nil(t, err)

	fmt.Println("Inserted document:", insertResult.InsertedID)

	cur, err := collection.Find(context.Background(), bson.D{})
	require.Nil(t, err)
	defer cur.Close(context.Background())

	var results []device.Yabby3Commit
	err = cur.All(context.Background(), &results)
	require.Nil(t, err)
	require.Equal(t, 810095, results[0].SerNo)
	require.Equal(t, 1, len(results))
	require.Equal(t, 2, len(results[0].Records))
}

func insert(collection *mongo.Collection, jsonstr string) (*mongo.InsertOneResult, error) {
	// Unmarshal JSON to map
	var data map[string]interface{}
	err := json.Unmarshal([]byte(jsonstr), &data)
	if err != nil {
		return nil, err
	}

	// Insert into MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	insertResult, err := collection.InsertOne(ctx, data)

	return insertResult, err
}

func Test_GetLatestPosition(t *testing.T) {

	// GIVEN two commits with multiple records and for multiple tags.
	collection, err := NewMongoConnection(mongoURL, randomTestCollectionName())
	require.Nil(t, err)

	storer := NewMongoStorer(collection)

	_, err = storer.WriteCommit(context.Background(), strippedDownMultiRecordSample1)
	require.Nil(t, err)

	_, err = storer.WriteCommit(context.Background(), strippedDownMultiRecordSample2)
	require.Nil(t, err)

	// WHEN we get the latest position for all tags.
	result, err := storer.GetLastPositions()
	require.Nil(t, err)

	// THEN we get the latest position's values for both known tags.
	for _, r := range result {
		switch r.SerNo {
		case 810095:
			require.Equal(t, 7495.0, r.SeqNo)
			require.Equal(t, -31.4577084, r.Latitude)
			require.Equal(t, 152.64215, r.Longitude)
			require.Equal(t, 35.0, r.Altitude)
			require.Equal(t, 1.0, r.Speed)
			require.Equal(t, 2.0, r.SpeedAcc)
			require.Equal(t, 3.0, r.Heading)
			require.Equal(t, 17.0, r.PDOP)
			require.Equal(t, 10.0, r.PosAcc)
			require.Equal(t, 7.0, r.GpsStatus)
			require.Equal(t, "2023-10-21 23:17:40", r.GpsUTC)
		case 810243:
			require.Equal(t, 7497.0, r.SeqNo)
			require.Equal(t, -32.1, r.Latitude)
			require.Equal(t, 153.1, r.Longitude)
		default:
			t.Fatalf("Unmatched serNo: %v", r.SerNo)
		}
	}

	// for cursor.Next(context.Background()) {
	// 	// result := struct{
	// 	//   Foo string
	// 	//   Bar int32
	// 	// }{}
	// 	var result bson.D

	// 	err = cursor.Decode(&result)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	x := result[1].Value
	// 	log.Printf("Type of x: %T\n", x)
	// 	log.Println(x)
	// 	doc, ok := x.(bson.D) // Assert 'x' to 'primitive.D'
	// 	if !ok {
	// 		log.Fatal("Couldn't type assert")
	// 	}
	// 	y := doc[0].Value
	// 	log.Print(y)
	// 	for _, elem := range result {
	// 		key := elem.Key
	// 		value := elem.Value

	// 		fmt.Println(key, value)

	// do something with result...
	// }
	// To get the raw bson bytes use cursor.Current
	//raw := cur.Current
	// do something with raw...
	// }
	// err = cursor.Err()
	// require.Nil(t, err)

}

func TestAgeFrom(t *testing.T) {
	now := func() time.Time {
		t, err := time.Parse(time.DateTime, "2023-11-19 23:21:42")
		if err != nil {
			panic(true) // TODO: what's supposd to be passed to panic?
		}
		return t
	}

	tests := []struct {
		future   string
		expected string
	}{
		{"2023-11-18 23:21:42", "1 days, 0 hours, 0 minutes"},
		{"2023-11-19 22:21:42", "1 hours, 0 minutes"},
		{"2023-11-19 23:20:42", "1 minutes"},
		{"2023-11-17 03:02:02", "2 days, 20 hours, 19 minutes"},
	}

	for _, tt := range tests {
		t.Run(tt.future, func(t *testing.T) {
			age := TimeAgoAsText(tt.future, now)
			require.Equal(t, tt.expected, age)
		})
	}

}
