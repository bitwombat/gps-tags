package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type storer struct {
	collection *mongo.Collection
}

func NewMongoConnection(mongoURL, collectionName string) (*mongo.Collection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		return nil, err
	}

	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	collection := client.Database("tags").Collection(collectionName)

	return collection, nil

}

func NewMongoStorer(collection *mongo.Collection) storer {
	return storer{collection: collection}
}

func (s storer) WriteCommit(ctx context.Context, jsonStr string) error {
	// Unmarshal JSON to map
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return err
	}

	// Insert into MongoDB
	insertResult, err := s.collection.InsertOne(ctx, data)
	if err != nil {
		return err
	}

	fmt.Println("Inserted document:", insertResult.InsertedID)

	return nil
}

func (s storer) GetLastPositions() ([]PositionRecord, error) {
	return nil, nil // TODO
}
