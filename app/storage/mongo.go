package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
)

func Insert(ctx context.Context, collection *mongo.Collection, jsonStr string) error {
	// Unmarshal JSON to map
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return err
	}

	// Insert into MongoDB
	insertResult, err := collection.InsertOne(ctx, data)
	if err != nil {
		return err
	}

	fmt.Println("Inserted document:", insertResult.InsertedID)

	return nil
}
