package main

import (
	"fmt"
	"net/http"

	"github.com/bitwombat/gps-tags/device"
	"github.com/bitwombat/gps-tags/storage"
	"github.com/bitwombat/gps-tags/substitute"
	"github.com/bitwombat/gps-tags/types"
)

func newPathsMapPageHandler(storer storage.Storage) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		debugLogger.Println("Got a current map page request.")
		lastWasHealthCheck = false

		pathpoints, err := storer.GetLastNPositions(30)
		if err != nil {
			errorLogger.Printf("Error getting last N positions from storage: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		subs := make(map[string]string)

		for _, tag := range pathpoints {
			// Start the JavaScript array
			pathpointStr := "["
			name := device.SerNoToName[tag.SerNo]
			for i, pathpoint := range tag.PathPoints {
				if i == 0 { // First point is most recent. Use this for the marker.
					subs[name+"Lat"] = fmt.Sprintf("%.7f", pathpoint.Latitude)
					subs[name+"Lng"] = fmt.Sprintf("%.7f", pathpoint.Longitude)
				}
				pathpointStr += fmt.Sprintf("{lat: %.7f, lng: %.7f},", pathpoint.Latitude, pathpoint.Longitude)
			}
			// Take the trailing comma off.
			pathpointStr = pathpointStr[:len(pathpointStr)-1]
			// Close out the array
			pathpointStr += "]"

			subs[name+"Path"] = pathpointStr
		}

		mapPage, err := substitute.ContentsOf("public_html/paths.html", subs)

		if err != nil {
			errorLogger.Printf("Error getting contents of paths.html: %v\n", err)
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
			name := device.SerNoToName[tag.SerNo]
			subs[name+"Lat"] = fmt.Sprintf("%.7f", tag.Latitude)
			subs[name+"Lng"] = fmt.Sprintf("%.7f", tag.Longitude)
			subs[name+"AccuracyRadius"] = fmt.Sprintf("%.7f", tag.PosAcc)
			subs[name+"Note"] = "Last GPS: " + timeAgoAsText(tag.GpsUTC) + " ago<br>Last Checkin: " + timeAgoAsText(tag.DateUTC) + " ago<br>Reason: " + types.ReasonCode(tag.Reason).String() + "<br>Battery: " + fmt.Sprintf("%.2f", tag.Battery/1000) + "V"
			subs[name+"Colour"] = timeAgoInColour(tag.GpsUTC)
		}

		mapPage, err := substitute.ContentsOf("public_html/current-map.html", subs)

		if err != nil {
			errorLogger.Printf("Error getting contents of current-map.html: %v\n", err)
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
