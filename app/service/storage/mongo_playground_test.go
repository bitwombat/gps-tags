//go:build playground

package storage

import (
	"context"
	"fmt"
	"time"

	"encoding/json"
	"testing"

	"github.com/bitwombat/gps-tags/device"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMongoInsertOne(t *testing.T) {
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

func TestMongoFind(t *testing.T) {
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

	var results []device.TagTx
	err = cur.All(context.Background(), &results)
	require.Nil(t, err)
	require.Equal(t, 810095, results[0].SerNo)
	require.Equal(t, 1, len(results))
	require.Equal(t, 2, len(results[0].Records))
}
