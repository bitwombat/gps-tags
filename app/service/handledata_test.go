package main

import (
	"fmt"
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
	cwd, _ := os.Getwd() //nolint:errcheck // don't care
	err := os.Chdir("..")
	require.Nil(t, err, "changing directory to where zone kml's are")

	now := func() time.Time {
		t, err := time.Parse(time.DateTime, "2025-09-04 13:21:42") // midday so battery notification happens
		if err != nil {
			panic("parsing time")
		}

		return t
	}

	storer := &FakeStorer{}
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

	assert.Equal(t, "RUEGER's battery low", string(notifier.notifications[0].title))
	assert.Equal(t, "Battery voltage: 1.641 V", string(notifier.notifications[0].message))
	assert.Equal(t, "RUEGER's battery critical", string(notifier.notifications[1].title))
	assert.Equal(t, "Battery voltage: 1.641 V", string(notifier.notifications[1].message))
	assert.Equal(t, "RUEGER is off the property", string(notifier.notifications[2].title))
	assert.Equal(t, "Last seen Not in any known zone.", string(notifier.notifications[2].message))

	_ = os.Chdir(cwd) //nolint:errcheck // don't care
}

func TestPostDataHandlerBatteryLevels(t *testing.T) {
	cwd, _ := os.Getwd() //nolint:errcheck // don't care
	err := os.Chdir("..")
	require.Nil(t, err, "changing directory to where zone kml's are")

	now := func() time.Time {
		t, err := time.Parse(time.DateTime, "2025-09-04 13:21:42") // midday so battery notification happens
		if err != nil {
			panic("parsing time")
		}
		return t
	}

	tests := []struct {
		name                string
		batteryMillivolts   int
		expectedNotifyCount int
		expectedTitles      []string
		expectedMessages    []string
		description         string
	}{
		{
			name:                "just_below_low_threshold",
			batteryMillivolts:   3999, // 3.999V (just below 4.0V threshold)
			expectedNotifyCount: 1,
			expectedTitles:      []string{"RUEGER's battery low"},
			expectedMessages:    []string{"Battery voltage: 3.999 V"},
			description:         "Battery just below low threshold should trigger low battery notification",
		},
		{
			name:                "at_low_threshold",
			batteryMillivolts:   4000, // 4.0V (exactly at threshold)
			expectedNotifyCount: 0,
			expectedTitles:      []string{},
			expectedMessages:    []string{},
			description:         "Battery exactly at threshold should NOT trigger (< comparison)",
		},
		{
			name:                "just_above_low_threshold",
			batteryMillivolts:   4001, // 4.001V (just above threshold)
			expectedNotifyCount: 0,
			expectedTitles:      []string{},
			expectedMessages:    []string{},
			description:         "Battery just above threshold should NOT trigger",
		},
		{
			name:                "just_below_reset_threshold",
			batteryMillivolts:   4099, // 4.099V (just below 4.1V reset)
			expectedNotifyCount: 0,
			expectedTitles:      []string{},
			expectedMessages:    []string{},
			description:         "Battery just below reset threshold should NOT trigger",
		},
		{
			name:                "at_reset_threshold",
			batteryMillivolts:   4100, // 4.1V (exactly at reset threshold)
			expectedNotifyCount: 0,
			expectedTitles:      []string{},
			expectedMessages:    []string{},
			description:         "Battery at reset threshold should NOT trigger",
		},
		{
			name:                "just_above_reset_threshold",
			batteryMillivolts:   4101, // 4.101V (just above reset)
			expectedNotifyCount: 0,
			expectedTitles:      []string{},
			expectedMessages:    []string{},
			description:         "Battery just above reset threshold should NOT trigger (no previous low state to reset from)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storer := &FakeStorer{}
			notifier := &FakeNotifier{}
			handler := newDataPostHandler(storer, notifier, "xxxx", now)

			batteryTestSample := fmt.Sprintf(`{
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
          "AnalogueData": {
            "1": %d,
            "3": 9999,
            "4": 9999,
            "5": 9999
          },
          "FType": 6
        }
      ]
    }
  ]
}`, tt.batteryMillivolts)

			body := strings.NewReader(batteryTestSample)
			req := httptest.NewRequest(http.MethodPost, "http://example.com/foo", body)
			header := http.Header{}
			header.Add("auth", "xxxx")
			req.Header = header

			w := httptest.NewRecorder()
			handler(w, req)

			resp := w.Result()
			require.Equal(t, http.StatusOK, resp.StatusCode, "HTTP status")

			require.Len(t, notifier.notifications, tt.expectedNotifyCount, tt.description)

			for i, expectedTitle := range tt.expectedTitles {
				assert.Equal(t, expectedTitle, string(notifier.notifications[i].title))
				assert.Equal(t, tt.expectedMessages[i], string(notifier.notifications[i].message))
			}
		})
	}

	_ = os.Chdir(cwd) //nolint:errcheck // don't care
}

func TestPostDataHandlerBatteryRecovery(t *testing.T) {
	cwd, _ := os.Getwd() //nolint:errcheck // don't care
	err := os.Chdir("..")
	require.Nil(t, err, "changing directory to where zone kml's are")

	now := func() time.Time {
		t, err := time.Parse(time.DateTime, "2025-09-04 13:21:42") // midday so battery notification happens
		if err != nil {
			panic("parsing time")
		}
		return t
	}

	storer := &FakeStorer{}
	notifier := &FakeNotifier{}
	handler := newDataPostHandler(storer, notifier, "xxxx", now)

	// First: Send low battery data to trigger the "set" state
	lowBatterySample := `{
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
          "AnalogueData": {
            "1": 3999,
            "3": 9999,
            "4": 9999,
            "5": 9999
          },
          "FType": 6
        }
      ]
    }
  ]
}`

	body := strings.NewReader(lowBatterySample)
	req := httptest.NewRequest(http.MethodPost, "http://example.com/foo", body)
	header := http.Header{}
	header.Add("auth", "xxxx")
	req.Header = header

	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	require.Equal(t, http.StatusOK, resp.StatusCode, "HTTP status")

	// Should have low battery notification
	require.Len(t, notifier.notifications, 1, "Should have low battery notification")
	assert.Equal(t, "RUEGER's battery low", string(notifier.notifications[0].title))
	assert.Equal(t, "Battery voltage: 3.999 V", string(notifier.notifications[0].message))

	// Clear notifications for next test
	notifier.notifications = nil

	// Second: Send recovery battery data (above reset threshold)
	recoverySample := `{
  "SerNo": 810095,
  "IMEI": "353785725680796",
  "ICCID": "89610180004127201829",
  "ProdId": 97,
  "FW": "97.2.1.11",
  "Records": [
    {
      "SeqNo": 7495,
      "Reason": 11,
      "DateUTC": "2023-10-21 23:23:42",
      "Fields": [
        {
          "AnalogueData": {
            "1": 4101,
            "3": 9999,
            "4": 9999,
            "5": 9999
          },
          "FType": 6
        }
      ]
    }
  ]
}`

	body2 := strings.NewReader(recoverySample)
	req2 := httptest.NewRequest(http.MethodPost, "http://example.com/foo", body2)
	req2.Header = header

	w2 := httptest.NewRecorder()
	handler(w2, req2)

	resp2 := w2.Result()
	require.Equal(t, http.StatusOK, resp2.StatusCode, "HTTP status")

	// Should now have "new battery detected" reset notification
	require.Len(t, notifier.notifications, 1, "Should have new battery detected notification")
	assert.Equal(t, "New battery for RUEGER detected", string(notifier.notifications[0].title))
	assert.Equal(t, "Battery voltage: 4.101 V", string(notifier.notifications[0].message))

	_ = os.Chdir(cwd) //nolint:errcheck // don't care
}

func TestPostDataHandlerAuth(t *testing.T) {
	cwd, _ := os.Getwd() //nolint:errcheck // don't care
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

	_ = os.Chdir(cwd) //nolint:errcheck // don't care
}
