package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTimeAgoAsText(t *testing.T) {
	now := func() time.Time {
		t, err := time.Parse(time.DateTime, "2023-11-19 23:21:42")
		if err != nil {
			panic("parsing time")
		}
		return t
	}

	tests := []struct {
		future   string
		expected string
	}{
		{"2023-11-19 23:20:42", "1 minutes"},
		{"2023-11-19 22:21:42", "1 hours, 0 minutes"},
		{"2023-11-18 23:21:42", "1 days, 0 hours, 0 minutes"},
		{"2023-11-17 03:02:02", "2 days, 20 hours, 19 minutes"},
	}

	for _, tt := range tests {
		t.Run(tt.future, func(t *testing.T) {
			age := StrTimeAgoAsText(tt.future, now)
			require.Equal(t, tt.expected, age)
		})
	}
}
