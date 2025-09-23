// Tests the the live page by comparing it to a golden copy, ignoring dynamic
// data like coordinates.
// Weird test to run every time, so the -test.short flag is default in Makefile
package main

import (
	"errors"
	"io"
	"net/http"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLiveWebApplicationGolden(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	// Download the live web page
	url := "https://tags.bitwombat.com.au/current"
	resp, err := http.Get(url)
	require.NoError(t, err, "downloading web page")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "web page should return 200 OK")

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "reading response body")

	htmlContent := string(body)

	// Ensure we got some expected content
	require.Contains(t, htmlContent, "Rueger", "page should contain dog name")
	require.Contains(t, htmlContent, "google.maps", "page should contain Google Maps library")

	// Use our custom assertion that handles dynamic data
	assertGoldenWithDynamicData(t, "current_page_live", htmlContent)
}

// assertGoldenWithDynamicData compares HTML content against a golden file,
// but normalizes dynamic data like coordinates before comparison
func assertGoldenWithDynamicData(tb testing.TB, fileBasename, got string) {
	tb.Helper()

	normalizedGot := normalizeDynamicData(got)

	goldenFilename := "test-output/" + fileBasename + ".golden.html"
	golden, err := os.ReadFile(goldenFilename)
	if errors.Is(err, os.ErrNotExist) {
		// Create the golden file automatically on first run
		err := os.WriteFile(goldenFilename, []byte(normalizedGot), 0o644) //nolint:gosec  // Test code, don't care
		if err != nil {
			tb.Fatalf("Couldn't create golden file %s: %v", goldenFilename, err)
		}
		tb.Logf("Created golden file %s", goldenFilename)
		return
	}
	if err != nil {
		tb.Fatalf("error reading %s: %v", goldenFilename, err)
	}

	require.Equal(tb, string(golden), normalizedGot)
}

// normalizeDynamicData replaces dynamic values with placeholders
// so that coordinate changes don't break the test
func normalizeDynamicData(html string) string {
	// Normalize latitude/longitude values (floating point numbers in various contexts)
	// This handles coordinates in JavaScript, HTML attributes, etc.

	// Match latitude/longitude pairs in various formats:
	// - JavaScript arrays: [lat, lng]
	// - Object properties: lat: value, lng: value
	// - HTML data attributes: data-lat="value"
	// - Common coordinate patterns

	replacements := []struct {
		regex *regexp.Regexp
		value string
	}{
		// JavaScript coordinate arrays like [lat, lng]
		{
			regexp.MustCompile(`\[-?\d+\.\d+,\s*-?\d+\.\d+\]`),
			"[LAT_PLACEHOLDER, LNG_PLACEHOLDER]",
		},
		// Object properties like lat: -31.123, lng: 152.456
		{
			regexp.MustCompile(`(lat|latitude):\s*-?\d+\.\d+`),
			"${1}: LAT_PLACEHOLDER",
		},
		{
			regexp.MustCompile(`(lng|longitude|long):\s*-?\d+\.\d+`),
			"${1}: LNG_PLACEHOLDER",
		},
		// HTML data attributes like data-lat="-31.123"
		{
			regexp.MustCompile(`data-(lat|latitude)="[^"]*"`),
			`data-${1}="LAT_PLACEHOLDER"`,
		},
		{
			regexp.MustCompile(`data-(lng|longitude|long)="[^"]*"`),
			`data-${1}="LNG_PLACEHOLDER"`,
		},
		// Common JavaScript variable assignments
		{
			regexp.MustCompile(`(var|let|const)\s+(lat|latitude)\s*=\s*-?\d+\.\d+`),
			"${1} ${2} = LAT_PLACEHOLDER",
		},
		{
			regexp.MustCompile(`(var|let|const)\s+(lng|longitude|long)\s*=\s*-?\d+\.\d+`),
			"${1} ${2} = LNG_PLACEHOLDER",
		},
		// Timestamps (in case there are any dynamic timestamps)
		{
			regexp.MustCompile(`\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2}`),
			"TIMESTAMP_PLACEHOLDER",
		},
		// Dynamic time-ago messages (e.g., "2 minutes ago", "1 hour ago")
		{
			regexp.MustCompile(`\d+\s+days,\s+\d+\s+hours,\s+\d+\s+minutes\s+ago`),
			"TIME_AGO_PLACEHOLDER",
		},
		{
			regexp.MustCompile(`\d+\s+hours,\s+\d+\s+minutes\s+ago`),
			"TIME_AGO_PLACEHOLDER",
		},
		{
			regexp.MustCompile(`\d+\s+minutes\s+ago`),
			"TIME_AGO_PLACEHOLDER",
		},
		// Dynamic time-ago marker colour
		{
			regexp.MustCompile(`"(red|#a23535|#8d8d8d)"`),
			"COLOUR_AGO_PLACEHOLDER",
		},
		// Any standalone floating point numbers that might be coordinates
		// (be more specific to avoid false positives)
		{
			regexp.MustCompile(`-?\d{1,3}\.\d{4,}`), // latitude/longitude typically have many decimal places
			"COORDINATE_PLACEHOLDER",
		},
	}

	normalizedHTML := html
	for _, replacement := range replacements {
		normalizedHTML = replacement.regex.ReplaceAllString(normalizedHTML, replacement.value)
	}

	return normalizedHTML
}
