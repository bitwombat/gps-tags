package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const basicCompleteSample = `{
  "SerNo": 810095,
  "IMEI": "353785725680796",
  "ICCID": "89610180004127201829",
  "ProdId": 97,
  "FW": "97.2.1.11",
  "Records": [
    {
      "SeqNo": 7494,
      "Reason": 11,
      "DateUTC": "2023-10-21 23:21:42",
      "Fields": [
        {
          "GpsUTC": "2023-10-21 23:17:40",
          "Lat": -31.4577084,
          "Long": 152.64215,
          "Alt": 35,
          "Spd": 0,
          "SpdAcc": 2,
          "Head": 0,
          "PDOP": 17,
          "PosAcc": 10,
          "GpsStat": 7,
          "FType": 0
        },
        {
          "DIn": 1,
          "DOut": 0,
          "DevStat": 1,
          "FType": 2
        },
        {
          "AnalogueData": {
            "1": 1641,
            "3": 3500,
            "4": 8,
            "5": 4500
          },
          "FType": 6
        }
      ]
    },
    {
      "SeqNo": 7495,
      "Reason": 2,
      "DateUTC": "2023-10-21 23:23:36",
      "Fields": [
        {
          "GpsUTC": "2023-10-21 23:17:40",
          "Lat": -32.4577084,
          "Long": 152.64215,
          "Alt": 35,
          "Spd": 0,
          "SpdAcc": 2,
          "Head": 0,
          "PDOP": 17,
          "PosAcc": 10,
          "GpsStat": 7,
          "FType": 0
        },
        {
          "TT": 2,
          "Trim": 300,
          "FType": 15
        },
        {
          "DIn": 0,
          "DOut": 0,
          "DevStat": 0,
          "FType": 2
        },
        {
          "AnalogueData": {
            "1": 1641,
            "3": 3400,
            "4": 8,
            "5": 4504
          },
          "FType": 6
        }
      ]
    }
  ]
}`

func TestPostDataHandler(t *testing.T) {
	cwd, _ := os.Getwd()
	err := os.Chdir("..")
	require.Nil(t, err, "changing directory to where zone kml's are")

	storer := &FakeStorer{}

	now := func() time.Time {
		t, err := time.Parse(time.DateTime, "2025-09-04 13:21:42") // midday so battery notification happens
		if err != nil {
			panic("parsing time")
		}

		return t
	}

	notifier := &FakeNotifier{}
	handler := newDataPostHandler(storer, notifier, "xxxx", now)
	body := strings.NewReader(basicCompleteSample)
	req := httptest.NewRequest(http.MethodPost, "http://example.com/foo", body)
	header := http.Header{}
	header.Add("auth", "xxxx")
	req.Header = header
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	require.Equal(t, http.StatusOK, resp.StatusCode, "HTTP status")

	assert.Equal(t, "RUEGER's battery low", notifier.notifications[0].title)
	assert.Equal(t, "Battery voltage: 1.641 V", notifier.notifications[0].message)
	assert.Equal(t, "RUEGER's battery critical", notifier.notifications[1].title)
	assert.Equal(t, "Battery voltage: 1.641 V", notifier.notifications[1].message)
	assert.Equal(t, "RUEGER is off the property", notifier.notifications[2].title)
	assert.Equal(t, "Last seen Not in any known zone.", notifier.notifications[2].message)

	_ = os.Chdir(cwd)
}

func TestPostDataHandlerAuth(t *testing.T) {
	cwd, _ := os.Getwd()
	err := os.Chdir("..")
	require.Nil(t, err, "changing directory to where zone kml's are")

	storer := &FakeStorer{}

	now := func() time.Time {
		return time.Time{}
	}

	notifier := &FakeNotifier{}
	handler := newDataPostHandler(storer, notifier, "xxxx", now)
	req := httptest.NewRequest(http.MethodPost, "http://example.com/foo", http.NoBody)

	t.Run("no auth header", func(t *testing.T) {
		w := httptest.NewRecorder()
		handler(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode, "HTTP status")
	})

	t.Run("empty auth", func(t *testing.T) {
		header := http.Header{}
		header.Add("auth", "")
		req.Header = header
		w := httptest.NewRecorder()
		handler(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode, "HTTP status")
	})

	t.Run("wrong auth", func(t *testing.T) {
		header := http.Header{}
		header.Add("auth", "xxxy")
		req.Header = header
		w := httptest.NewRecorder()
		handler(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode, "HTTP status")
	})

	_ = os.Chdir(cwd)
}
