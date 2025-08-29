package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"testing"

	"github.com/stretchr/testify/require"
)

func randomTestFileName() string {
	bytes := make([]byte, 8)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}

	return "test-" + hex.EncodeToString(bytes)
}

func TestGetLatestPosition(t *testing.T) {
	// GIVEN two commits with multiple records and for multiple tags.
	collection, err := NewMongoConnection(mongoURL, randomTestCollectionName())
	require.Nil(t, err)

	storer := NewMongoStorer(collection)

	_, err = storer.WriteTx(context.Background(), strippedDownMultiRecordSample1)
	require.Nil(t, err)

	_, err = storer.WriteTx(context.Background(), strippedDownMultiRecordSample2)
	require.Nil(t, err)

	// WHEN we get the latest position for all tags.
	result, err := storer.GetLastPositions()
	require.Nil(t, err)

	// THEN we get the latest position's values for both known tags.
	require.Len(t, result, 2, "length of result array")
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

}

func TestGetLastNPositions(t *testing.T) {

	// GIVEN commits with multiple records and for multiple tags.
	collection, err := NewMongoConnection(mongoURL, randomTestCollectionName())
	require.Nil(t, err)

	storer := NewMongoStorer(collection)

	for _, r := range nSamples {
		if s, ok := r.(string); ok {
			_, err = storer.WriteTx(context.Background(), s)
			require.Nil(t, err)
		} else {
			t.Fatal("type assertion failed")
		}
	}

	// WHEN we get the last 3 position for all tags.
	result, err := storer.GetLastNPositions(3)
	require.Nil(t, err)

	// THEN we get the latest position's values for both known tags.
	for _, r := range result {
		switch r.SerNo {
		case 810095:
			require.Equal(t, 108.0, r.PathPoints[0].Latitude)
			require.Equal(t, 109.0, r.PathPoints[0].Longitude)
			require.Equal(t, 106.0, r.PathPoints[1].Latitude)
			require.Equal(t, 107.0, r.PathPoints[1].Longitude)
			require.Equal(t, 104.0, r.PathPoints[2].Latitude)
			require.Equal(t, 105.0, r.PathPoints[2].Longitude)
		case 810243:
			require.Equal(t, 118.0, r.PathPoints[0].Latitude)
			require.Equal(t, 119.0, r.PathPoints[0].Longitude)
			require.Equal(t, 116.0, r.PathPoints[1].Latitude)
			require.Equal(t, 117.0, r.PathPoints[1].Longitude)
			require.Equal(t, 114.0, r.PathPoints[2].Latitude)
			require.Equal(t, 115.0, r.PathPoints[2].Longitude)
		default:
			t.Fatalf("Unmatched serNo: %v", r.SerNo)
		}
	}

}
