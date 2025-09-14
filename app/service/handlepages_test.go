package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/bitwombat/gps-tags/storage"
	"github.com/stretchr/testify/require"
)

func mkTime(ts string) time.Time {
	t, err := time.Parse(time.DateTime, ts)
	if err != nil {
		panic("parsing time")
	}
	return t
}

func TestCurrentMapPageHandler(t *testing.T) {
	err := os.Chdir("..")
	require.Nil(t, err, "changing directory to where public_html is")

	storer := &FakeStorer{
		fnGetLastStatuses: func(_ context.Context) (storage.Statuses, error) {
			return storage.Statuses{
				810095: {
					SeqNo:     1,
					Reason:    3,
					Latitude:  5.0,
					Longitude: 7.0,
					Altitude:  11,
					Speed:     13,
					DateUTC:   mkTime("2025-09-02 10:07:00"),
					GpsUTC:    mkTime("2025-09-03 11:08:01"),
					PosAcc:    17,
					GpsStatus: 23,
					Battery:   27,
				},

				810243: {
					SeqNo:     11,
					Reason:    13,
					Latitude:  15.0,
					Longitude: 17.0,
					Altitude:  111,
					Speed:     113,
					DateUTC:   mkTime("2024-09-02 10:07:00"),
					GpsUTC:    mkTime("2024-09-03 11:08:01"),
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
			panic("parsing time")
		}

		return t
	}

	handler := newCurrentMapPageHandler(storer, now)
	req := httptest.NewRequest(http.MethodGet, "http://example.com/foo", http.NoBody)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	require.Equal(t, 200, resp.StatusCode, "HTTP status")
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	require.Nil(t, err)

	// This is just an early, clearer fail
	require.Contains(t, string(body), "Rueger")

	assertGolden(t, "current_page", string(body))
}

func assertGolden(tb testing.TB, fileBasename, got string) {
	tb.Helper()

	// We always want a current file, so write it out first, before we have a
	// possibility of bailing if the golden file isn't found.
	currentFilename := "service/test-output/" + fileBasename + ".html"
	err := os.WriteFile(currentFilename, []byte(got), 0o644) //nolint:gosec  // Test code, don't care
	if err != nil {
		tb.Fatalf("Couldn't write html file %s: %v", currentFilename, err)
	}

	goldenFilename := "service/test-output/" + fileBasename + ".golden.html"
	golden, err := os.ReadFile(goldenFilename)
	if errors.Is(err, os.ErrNotExist) {
		tb.Fatal(goldenFilename + " does not exist. Copy it from " + currentFilename)
	}
	if err != nil {
		tb.Fatalf("error reading %s: %v", goldenFilename, err)
	}

	require.Equal(tb, string(golden), got)
}
