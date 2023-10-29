package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"encoding/json"
	"testing"

	"github.com/bitwombat/tag/device"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

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

const strippedDownMultiRecordSample = `{
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
          "Long": 152.64215
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
	collection, err := NewMongoConnection(randomTestCollectionName())
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
	collection, err := NewMongoConnection(randomTestCollectionName())
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

func Test_GetLatestPosition(t *testing.T) {
	collection, err := NewMongoConnection(randomTestCollectionName())
	require.Nil(t, err)

	// Unmarshal JSON to map
	var data map[string]interface{}
	err = json.Unmarshal([]byte(strippedDownMultiRecordSample), &data)
	require.Nil(t, err)

	// Insert into MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	insertResult, err := collection.InsertOne(ctx, data)
	require.Nil(t, err)

	fmt.Println("Inserted document:", insertResult.InsertedID)

	pipeline := []bson.M{
		{
			"$unwind": "$Records",
		},
		{
			"$project": bson.M{
				"serNo":     "$SerNo",
				"latitude":  "$Records.Fields.Lat",
				"longitude": "$Records.Fields.Long",
			},
		},
	}
	cursor, err := collection.Aggregate(ctx, pipeline)
	require.Nil(t, err)
	defer cursor.Close(ctx)

	var result []bson.M
	err = cursor.All(ctx, &result)
	if err != nil {
		log.Fatal(err)
	}

	type Location struct {
		SerNo     float64
		Latitude  []float64
		Longitude []float64
	}

	x := result[0]
	y := result[0]["latitude"][0]
	doc, err := bson.Marshal(x)
	if err != nil {
		log.Fatal(err)
	}

	var loc Location
	err = bson.Unmarshal(doc, &loc)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(loc)
	require.Equal(t, 2, len(result))
	require.Equal(t, -31.4577084, result[0]["latitude"])

}