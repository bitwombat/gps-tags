package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bitwombat/gps-tags/device"
	"github.com/bitwombat/gps-tags/model"
	"github.com/bitwombat/gps-tags/storage"
	"github.com/bitwombat/gps-tags/substitute"
)

func newPathsMapPageHandler(storer storage.Storage) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		debugLogger.Println("Got a current map page request.")
		lastWasHealthCheck = false

		ctx := context.Background() // TODO: Correct? Should this be coming in from somewhere?
		pathpoints, err := storer.GetLastNPositions(ctx, 30)
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
			// Take the trailing comma off
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

		// Don't need w.WriteHeader(http.StatusOK) - it's taken care of by w.Write
	}
}

func newCurrentMapPageHandler(storer storage.Storage) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		debugLogger.Println("Got a current map page request.")
		lastWasHealthCheck = false

		ctx := context.Background() // TODO: Correct? Should this be coming in from somewhere?

		tags, err := storer.GetLastPositions(ctx)
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
			subs[name+"AccuracyRadius"] = fmt.Sprintf("%v", tag.PosAcc)
			subs[name+"Note"] = "Last GPS: " + timeAgoAsText(tag.GpsUTC) + " ago<br>Last Checkin: " + timeAgoAsText(tag.DateUTC) + " ago<br>Reason: " + model.ReasonCode(tag.Reason).String() + "<br>Battery: " + fmt.Sprintf("%.2f", float64(tag.Battery)/1000.) + "V"
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
