package main

import (
	"fmt"
	"net/http"

	"github.com/bitwombat/gps-tags/device"
	"github.com/bitwombat/gps-tags/storage"
	"github.com/bitwombat/gps-tags/substitute"
)

func newCurrentMapPageHandler(storer storage.Storage) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		debugLogger.Println("Got a current map page request.")
		lastWasHealthCheck = false

		tags, err := storer.GetLastPositions()
		if err != nil {
			errorLogger.Printf("Error getting last position from storage: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		subs := make(map[string]string)

		for _, tag := range tags {
			name := idToName[tag.SerNo]
			reason, ok := device.ReasonToText[tag.Reason]
			if !ok {
				errorLogger.Printf("Error: Unknown reason code: %v\n", tag.Reason)
				reason = "Unknown reason"
			}
			subs[name+"Lat"] = fmt.Sprintf("%.7f", tag.Latitude)
			subs[name+"Lng"] = fmt.Sprintf("%.7f", tag.Longitude)
			subs[name+"AccuracyRadius"] = fmt.Sprintf("%.7f", tag.PosAcc)
			subs[name+"Note"] = "Last GPS: " + timeAgoAsText(tag.GpsUTC) + " ago<br>Last Checkin: " + timeAgoAsText(tag.DateUTC) + " ago<br>Reason: " + reason + "<br>Battery: " + fmt.Sprintf("%.2f", tag.Battery/1000) + "V"
			subs[name+"Colour"] = timeAgoInColour(tag.GpsUTC)
		}

		mapPage, err := substitute.ContentsOf("public_html/current-map.html", subs)

		if err != nil {
			errorLogger.Printf("Error getting contents of index.html: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = w.Write([]byte(mapPage))
		if err != nil {
			errorLogger.Printf("Error writing response: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Don't need this - it's taken care of by w.Write:  w.WriteHeader(http.StatusOK)
	}
}
