package main

import (
	"context"
	"errors"
	"io"
	"net/http/httptest"
	"os"
	"testing"
	"time"

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

	now := func() time.Time {
		t, err := time.Parse(time.DateTime, "2025-09-04 23:21:42")
		if err != nil {
			panic(true) // TODO: what's supposd to be passed to panic?
		}
		return t
	}

	handler := newCurrentMapPageHandler(storer, now)
	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	require.Equal(t, 200, resp.StatusCode, "HTTP status")
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	require.Nil(t, err)

	// early, clearer fail than later checks
	require.Contains(t, string(body), "Rueger")

	assertGolden(t, "current_page", string(body))
}

func assertGolden(t *testing.T, fileBasename string, got string) {
	// We always want a current file, so write it out first, before we have a
	// possibility of bailing if the golden file isn't found.
	currentFilename := "service/test-output/" + fileBasename + ".html"
	err := os.WriteFile(currentFilename, []byte(got), 0644)
	if err != nil {
		t.Fatalf("Couldn't write html file %s: %v", currentFilename, err)
	}

	goldenFilename := "service/test-output/" + fileBasename + ".golden.html"
	golden, err := os.ReadFile(goldenFilename)
	if errors.Is(err, os.ErrNotExist) {
		t.Fatal(goldenFilename + " does not exist. Copy it from " + currentFilename)
	}
	if err != nil {
		t.Fatalf("error reading %s: %v", goldenFilename, err)
	}

	require.Equal(t, string(golden), got)
}
