package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bitwombat/gps-tags/model"
	"github.com/bitwombat/gps-tags/storage"
	"github.com/bitwombat/gps-tags/substitute"
)

func newPathsMapPageHandler(storer storage.Storage) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		debugLogger.Println("Got a current map page request.")
		lastWasHealthCheck = false

		ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
		defer cancel()

		pathpoints, err := storer.GetLastNCoords(ctx, 30)
		if err != nil {
			errorLogger.Printf("Error getting last N coordinates from storage: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		subs := make(map[string]string)

		for tag, points := range pathpoints {
			// Start the JavaScript array
			pathpointStr := "["
			name := model.SerNoToName[int(tag)]
			for i, pathpoint := range points {
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

func newCurrentMapPageHandler(storer storage.Storage, now func() time.Time) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		debugLogger.Println("Got a current map page request.")
		lastWasHealthCheck = false

		ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
		defer cancel()

		tagStatuses, err := storer.GetLastStatuses(ctx)
		if err != nil {
			errorLogger.Printf("Error getting last statuses from storage: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		subs := make(map[string]string)

		for _, tag := range tagStatuses {
			name := model.SerNoToName[int(tag.SerNo)]
			subs[name+"Lat"] = fmt.Sprintf("%.7f", tag.Latitude)
			subs[name+"Lng"] = fmt.Sprintf("%.7f", tag.Longitude)
			subs[name+"AccuracyRadius"] = fmt.Sprintf("%v", tag.PosAcc)
			subs[name+"Note"] = "Last GPS: " + storage.TimeAgoAsText(tag.GpsUTC, now) + " ago<br>Last Checkin: " + storage.TimeAgoAsText(tag.DateUTC, now) + " ago<br>Reason: " + tag.Reason.String() + "<br>Battery: " + fmt.Sprintf("%.2f", float64(tag.Battery)/1000.) + "V"
			subs[name+"Colour"] = storage.TimeAgoInColour(tag.GpsUTC, now)
		}

		mapPage, err := substitute.ContentsOf("public_html/current-map.html", subs)
		if err != nil {
			errorLogger.Printf("Error getting contents of current-map.html: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = w.Write([]byte(mapPage)) // NOTE: writes http.StatusOK header
		if err != nil {
			errorLogger.Printf("Error writing response: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
