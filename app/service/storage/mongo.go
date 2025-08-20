package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type storer struct {
	collection *mongo.Collection
}

// A field present in the document but not in the struct here doesn't break anything (is ignored).

// What Mongo returns is odd: sub-document, and single numbers (e.g lat/long) as
// arrays. Only [0] is ever populated.
type MongoPositionRecord struct {
	Document struct {
		SerNo     float64
		SeqNo     float64
		Reason    int64
		DateUTC   string
		Latitude  []float64
		Longitude []float64
		Altitude  []float64
		Speed     []float64
		GpsUTC    []string
		SpeedAcc  []float64
		Heading   []float64
		PDOP      []float64
		PosAcc    []float64
		GpsStatus []float64
		Battery   []float64
	}
}

type MongoPathPoint struct {
	Document []struct {
		SerNo     float64
		SeqNo     float64
		Latitude  []float64
		Longitude []float64
	}
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

func (s storer) WriteTx(ctx context.Context, jsonStr string) (string, error) {
	// Unmarshal JSON to map
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", err
	}

	// Insert into MongoDB
	insertResult, err := s.collection.InsertOne(ctx, data)
	if err != nil {
		return "", err
	}

	return fmt.Sprint(insertResult.InsertedID), nil
}

func (s storer) GetLastPositions() ([]PositionRecord, error) {
	// (A field present in the projection but not in the struct decoded to does not break anything.)
	pipeline := []bson.M{
		{
			"$unwind": "$Records",
		},
		{
			"$project": bson.M{
				"serNo":     "$SerNo",
				"seqNo":     "$Records.SeqNo",
				"reason":    "$Records.Reason",
				"dateUTC":   "$Records.DateUTC",
				"gpsUTC":    "$Records.Fields.GpsUTC",
				"latitude":  "$Records.Fields.Lat",
				"longitude": "$Records.Fields.Long",
				"altitude":  "$Records.Fields.Alt",
				"speed":     "$Records.Fields.Spd",
				"speedAcc":  "$Records.Fields.SpdAcc",
				"heading":   "$Records.Fields.Head",
				"PDOP":      "$Records.Fields.PDOP",
				"posAcc":    "$Records.Fields.PosAcc",
				"gpsStatus": "$Records.Fields.GpsStat",
				"battery":   "$Records.Fields.AnalogueData.1",
			},
		},
		{
			"$group": bson.M{
				"_id": "$serNo",
				"document": bson.M{
					"$top": bson.M{
						"sortBy": bson.M{"seqNo": -1},
						"output": "$$ROOT",
					},
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cursor, err := s.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("calling collection.Aggregate: %w", err)
	}
	defer cursor.Close(ctx)

	var result []MongoPositionRecord
	//var result []bson.M

	err = cursor.All(ctx, &result)
	if err != nil {
		return nil, fmt.Errorf("calling cursor.All: %w", err)
	}

	var records []PositionRecord
	for _, r := range result {
		records = append(records, *MarshalPositionRecord(r))
	}

	return records, nil
}

func (s storer) GetLastNPositions(n int) ([]PathPointRecord, error) {
	// (A field present in the projection but not in the struct decoded to does not break anything.)
	pipeline := []bson.M{
		{
			"$unwind": "$Records",
		},
		{
			"$project": bson.M{
				"serNo":     "$SerNo",
				"seqNo":     "$Records.SeqNo",
				"latitude":  "$Records.Fields.Lat",
				"longitude": "$Records.Fields.Long",
			},
		},
		{
			"$group": bson.M{
				"_id": "$serNo",
				"document": bson.M{
					"$topN": bson.M{
						"n":      n,
						"sortBy": bson.M{"seqNo": -1},
						"output": "$$ROOT",
					},
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cursor, err := s.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("calling collection.Aggregate: %w", err)
	}
	defer cursor.Close(ctx)

	var result []MongoPathPoint

	err = cursor.All(ctx, &result)
	if err != nil {
		return nil, fmt.Errorf("calling cursor.All: %w", err)
	}

	var pathPointRecords []PathPointRecord
	for _, pr := range result {
		pathPointRecords = append(pathPointRecords, MarshalPathPointRecord(pr))
	}

	return pathPointRecords, nil
}

func MarshalPositionRecord(m MongoPositionRecord) *PositionRecord {
	pr := &PositionRecord{
		SerNo:   m.Document.SerNo,
		SeqNo:   m.Document.SeqNo,
		Reason:  m.Document.Reason,
		DateUTC: m.Document.DateUTC,
	}

	// These are probably only ever absent because of tests which intentionally
	// have incomplete records...
	if len(m.Document.Latitude) > 0 {
		pr.Latitude = m.Document.Latitude[0]
	}

	if len(m.Document.Longitude) > 0 {
		pr.Longitude = m.Document.Longitude[0]
	}

	if len(m.Document.Altitude) > 0 {
		pr.Altitude = m.Document.Altitude[0]
	}

	if len(m.Document.Speed) > 0 {
		pr.Speed = m.Document.Speed[0]
	}

	if len(m.Document.GpsUTC) > 0 {
		pr.GpsUTC = m.Document.GpsUTC[0]
	}

	if len(m.Document.SpeedAcc) > 0 {
		pr.SpeedAcc = m.Document.SpeedAcc[0]
	}

	if len(m.Document.Heading) > 0 {
		pr.Heading = m.Document.Heading[0]
	}

	if len(m.Document.PDOP) > 0 {
		pr.PDOP = m.Document.PDOP[0]
	}

	if len(m.Document.PosAcc) > 0 {
		pr.PosAcc = m.Document.PosAcc[0]
	}

	if len(m.Document.GpsStatus) > 0 {
		pr.GpsStatus = m.Document.GpsStatus[0]
	}

	if len(m.Document.Battery) > 0 {
		pr.Battery = m.Document.Battery[0]
	}

	return pr
}

func MarshalPathPointRecord(m MongoPathPoint) PathPointRecord {
	pr := PathPointRecord{}

	var pathPoints []PathPoint
	for _, doc := range m.Document {
		pr.SerNo = doc.SerNo
		pathPoints = append(pathPoints, PathPoint{
			Latitude:  doc.Latitude[0],
			Longitude: doc.Longitude[0],
		})
	}
	pr.PathPoints = pathPoints

	return pr
}
