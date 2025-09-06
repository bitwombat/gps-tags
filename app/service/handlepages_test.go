package main

import (
	"context"
	"io"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/bitwombat/gps-tags/storage"
	"github.com/stretchr/testify/require"
)

func TestCurrentMapPageHandler(t *testing.T) {
	err := os.Chdir("..")
	require.Nil(t, err, "changing directory to where public_html is")

	storer := FakeStorer{
		fnGetLastPositions: func(_ context.Context) ([]storage.PositionRecord, error) {
			return []storage.PositionRecord{
				storage.PositionRecord{
					SerNo:     810095,
					SeqNo:     1,
					Reason:    3,
					Latitude:  5.0,
					Longitude: 7.0,
					Altitude:  11,
					Speed:     13,
					DateUTC:   "2025-09-02 10:07:00",
					GpsUTC:    "2025-09-03 11:08:01",
					PosAcc:    17,
					GpsStatus: 23,
					Battery:   27,
				},
				storage.PositionRecord{
					SerNo:     810243,
					SeqNo:     11,
					Reason:    13,
					Latitude:  15.0,
					Longitude: 17.0,
					Altitude:  111,
					Speed:     113,
					DateUTC:   "2024-09-02 10:07:00",
					GpsUTC:    "2024-09-03 11:08:01",
					PosAcc:    117,
					GpsStatus: 123,
					Battery:   127,
				},
			}, nil
		},
	}

	handler := newCurrentMapPageHandler(storer)
	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	require.Equal(t, 200, resp.StatusCode, "HTTP status")
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	require.Nil(t, err)

	require.Equal(t, "Hello, client\n", string(body))
}
