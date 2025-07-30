import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type storer struct {
	db *sql.DB
}

func NewSQLConnection(connectionString string) (*sql.DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, err
	}

	// Check the connection
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func NewSQLStorer(db *sql.DB) storer {
	return storer{db: db}
}

func (s storer) WriteCommit(ctx context.Context, jsonStr string) (string, error) {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return "", err
	}

	columns := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))

	for col, val := range data {
		columns = append(columns, fmt.Sprintf("%s", col))
		values = append(values, val)
	}

	placeholders := strings.Repeat(",?", len(columns)-1) + "?"
	sqlQuery := fmt.Sprintf("INSERT INTO commits (%s) VALUES (%s)", strings.Join(columns, ","), placeholders)

	_, err = s.db.ExecContext(ctx, sqlQuery, values...)
	if err != nil {
		return "", err
	}

	return "Success", nil
}

func (s storer) GetLastPositions() ([]PositionRecord, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	query := `
        SELECT ser_no, seq_no, reason, date_utc, latitude, longitude, altitude, speed, speed_acc, heading, pdop, pos_acc, gps_status, battery
        FROM positions
        WHERE (ser_no, seq_no) IN (
            SELECT ser_no, MAX(seq_no)
            FROM positions
            GROUP BY ser_no
        )
    `

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("executing query: %w", err)
	}
	defer rows.Close()

	var records []PositionRecord
	for rows.Next() {
		var pr PositionRecord
		err := rows.Scan(
			&pr.SerNo,
			&pr.SeqNo,
			&pr.Reason,
			&pr.DateUTC,
			&pr.Latitude,
			&pr.Longitude,
			&pr.Altitude,
			&pr.Speed,
			&pr.SpeedAcc,
			&pr.Heading,
			&pr.PDOP,
			&pr.PosAcc,
			&pr.GpsStatus,
			&pr.Battery,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}
		records = append(records, pr)
	}

	return records, nil
}

func (s storer) GetLastNPositions(n int) ([]PathPointRecord, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	query := `
        SELECT ser_no, latitude, longitude
        FROM (
            SELECT ser_no, latitude, longitude,
                   ROW_NUMBER() OVER (PARTITION BY ser_no ORDER BY seq_no DESC) AS rn
            FROM positions
        ) t
        WHERE rn <= $1
    `

	rows, err := s.db.QueryContext(ctx, query, n)
	if err != nil {
		return nil, fmt.Errorf("executing query: %w", err)
	}
	defer rows.Close()

	var pathPoints []PathPoint
	currentSerNo := ""
	for rows.Next() {
		var pp PathPoint
		var serNo string
		err := rows.Scan(&serNo, &pp.Latitude, &pp.Longitude)
		if err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}

		if currentSerNo == "" || currentSerNo != serNo {
			pathPoints = append(pathPoints, pp)
			currentSerNo = serNo
		}
	}

	return []PathPointRecord{{SerNo: currentSerNo, PathPoints: pathPoints}}, nil
}

