//go:build playground

package storage

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestSqliteInsertOne(t *testing.T) {
	const dsnURI string = ":memory:"
	db, err := sql.Open("sqlite", dsnURI)
	require.Nil(t, err)
	defer db.Close()

	ctx := context.Background()

	err = db.PingContext(ctx)
	require.Nil(t, err)

	result, err := db.ExecContext(ctx, "DROP TABLE IF EXISTS tx;")
	require.Nil(t, err)

	result, err = db.ExecContext(ctx, `
  CREATE TABLE tx (
      ID INTEGER PRIMARY KEY AUTOINCREMENT,
      ProdID INTEGER,
      Fw TEXT,
      created_at TEXT
  ) STRICT;`) // STRICT makes types have to match exactly. So INTEGER not INT.

	require.Nil(t, err)
	require.NotNil(t, result)

	deviceTime, err := time.Parse(time.DateTime, "2025-09-04 23:21:42")
	if err != nil {
		panic("parsing time")
	}

	result, err = db.ExecContext(ctx,
		"INSERT INTO tx (ProdID, Fw, created_at) VALUES ($1, $2, $3)",
		27,
		"gopher",
		deviceTime,
	)

	require.Nil(t, err)
	require.NotNil(t, result)

	rows, err := db.QueryContext(ctx, "SELECT * FROM Tx")
	require.Nil(t, err)
	defer rows.Close()

	for rows.Next() {
		var id sql.NullString
		var prodID int
		var fw sql.NullString
		var createdAt sql.NullString
		err := rows.Scan(&id, &prodID, &fw, &createdAt)
		require.Nil(t, err)
		fmt.Printf("%v %v %v %v", id, prodID, fw, createdAt)
	}

	err = rows.Err()
	require.Nil(t, err)

	/*
		// Unmarshal JSON to map
		var data map[string]interface{}
		err = json.Unmarshal([]byte(basicCompleteSample), &data)
		require.Nil(t, err)

		// Insert into SQLite
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		insertResult, err := collection.InsertOne(ctx, data)
		require.Nil(t, err)

		fmt.Println("Inserted document:", insertResult.InsertedID)
	*/
}

/*
func TestSqliteFind(t *testing.T) {
	collection, err := NewMongoConnection(mongoURL, randomTestCollectionName())
	require.Nil(t, err)

	// Unmarshal JSON to map
	var data map[string]interface{}
	err = json.Unmarshal([]byte(basicCompleteSample), &data)
	require.Nil(t, err)

	// Insert into SQLite
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
*/
