package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/bitwombat/gps-tags/device"
	"github.com/bitwombat/gps-tags/notify"
	"github.com/bitwombat/gps-tags/poly"
	"github.com/bitwombat/gps-tags/storage"
	"github.com/bitwombat/gps-tags/substitute"
	zonespkg "github.com/bitwombat/gps-tags/zones"
)

type TagData struct {
	SerNo   int      `json:"SerNo"`
	Imei    string   `json:"IMEI"`
	Iccid   string   `json:"ICCID"`
	ProdID  int      `json:"ProdId"`
	Fw      string   `json:"FW"`
	Records []Record `json:"Records"`
}

type Record struct {
	SeqNo   int     `json:"SeqNo"`
	Reason  int     `json:"Reason"`
	DateUTC string  `json:"DateUTC"`
	Fields  []Field `json:"Fields"`
}

type Field struct {
	GpsUTC       string       `json:"GpsUTC,omitempty"`
	Lat          float64      `json:"Lat,omitempty"`
	Long         float64      `json:"Long,omitempty"`
	Alt          int          `json:"Alt,omitempty"`
	Spd          int          `json:"Spd,omitempty"`
	SpdAcc       int          `json:"SpdAcc,omitempty"`
	Head         int          `json:"Head,omitempty"`
	Pdop         int          `json:"PDOP,omitempty"`
	PosAcc       int          `json:"PosAcc,omitempty"`
	GpsStat      int          `json:"GpsStat,omitempty"`
	FType        int          `json:"FType"`
	DIn          int          `json:"DIn,omitempty"`
	DOut         int          `json:"DOut,omitempty"`
	DevStat      int          `json:"DevStat,omitempty"`
	AnalogueData AnalogueData `json:"AnalogueData,omitempty"`
}

type AnalogueData struct {
	Num1 int `json:"1"`
	Num3 int `json:"3"`
	Num4 int `json:"4"`
	Num5 int `json:"5"`
}

var idToName = map[float64]string{
	810095: "rueger",
	810243: "tucker",
}

var lastWasHealthCheck bool // Used to clean up the log output.

type boundaryStatesType map[string]bool
type dogBoundaryStatesType map[string]boundaryStatesType

func newCurrentMapPageHandler(storer storage.Storage) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got a current map page request.")
		lastWasHealthCheck = false

		tags, err := storer.GetLastPositions()
		if err != nil {
			log.Printf("Error getting last position from storage: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		subs := make(map[string]string)

		for _, tag := range tags {
			name := idToName[tag.SerNo]
			reason, ok := device.ReasonToText[tag.Reason]
			if !ok {
				log.Printf("Error: Unknown reason code: %v\n", tag.Reason)
				reason = "Unknown reason"
			}
			subs[name+"Lat"] = fmt.Sprintf("%.7f", tag.Latitude)
			subs[name+"Lng"] = fmt.Sprintf("%.7f", tag.Longitude)
			subs[name+"Note"] = "Last GPS: " + timeAgoAsText(tag.GpsUTC) + " ago<br>Last Checkin: " + timeAgoAsText(tag.DateUTC) + " ago<br>Reason: " + reason + "<br>Battery: " + fmt.Sprintf("%.2f", tag.Battery/1000) + "V"
			subs[name+"Colour"] = timeAgoInColour(tag.GpsUTC)
		}

		mapPage, err := substitute.ContentsOf("public_html/current-map.html", subs)

		if err != nil {
			log.Printf("Error getting contents of index.html: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = w.Write([]byte(mapPage))
		if err != nil {
			log.Printf("Error writing response: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Don't need this - it's taken care of by w.Write:  w.WriteHeader(http.StatusOK)
	}
}

func newDataPostHandler(s storage.Storage, n notify.Notifier) func(http.ResponseWriter, *http.Request) {
	storer := s // TODO: Is this necessary for a closure?
	notifier := n

	prevDogBoundaryState := make(dogBoundaryStatesType)

	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		// Validate the request
		if r.Method != http.MethodPost {
			log.Println("Got a request to /upload that was not a POST")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		authKey := r.Header[http.CanonicalHeaderKey("auth")][0]

		if authKey == "" {
			log.Printf("Got an empty auth key: %v\n", authKey)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if authKey != "6ebaa65ed27455fd6d32bfd4c01303cd" {
			log.Printf("Got a bad auth key: %v\n", authKey)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		log.Println("Got a data post.")
		lastWasHealthCheck = false

		// Read and decode the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Println(string(body))

		var tagData TagData

		err = json.Unmarshal(body, &tagData)
		if err != nil {
			log.Printf("Error unmarshalling JSON: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var thisZoneText string

		NamedZones, err := zonespkg.ReadKMLDir("named_zones")
		if err != nil {
			log.Printf("Error reading KML files: %v", err)
			// not a critical error, keep going
			return
		}

		dogName, ok := idToName[float64(tagData.SerNo)]
		if !ok {
			log.Printf("Unknown tag number: %v", tagData.SerNo)
		}
		dogName = strings.ToUpper(dogName) // Just looks better and stands out in notifications

		var latestRecord Record

		// Process the records
		for _, r := range tagData.Records {

			// Figure out the most recent record for notifications later
			if r.SeqNo > latestRecord.SeqNo {
				latestRecord = r
			}

			gpsField := r.Fields[0]

			if NamedZones != nil {
				thisZoneText = zonespkg.NameThatZone(NamedZones, zonespkg.Point{Latitude: gpsField.Lat, Longitude: gpsField.Long})
			} else {
				thisZoneText = "No zones loaded"
			}

			reason, ok := device.ReasonToText[int64(r.Reason)]
			if !ok {
				log.Printf("Error: Unknown reason code: %v\n", r.Reason)
				reason = "Unknown reason"
			}

			log.Printf("%v/%s  %s (%s ago) \"%v\"  %s (%s ago) %0.7f,%0.7f \"%s\"\n", tagData.SerNo, dogName, r.DateUTC, timeAgoAsText(r.DateUTC), reason, gpsField.GpsUTC, timeAgoAsText(gpsField.GpsUTC), gpsField.Lat, gpsField.Long, thisZoneText)

		}

		// Notify based on most recent record in the set just sent
		latestGPSRecord := latestRecord.Fields[0]

		if NamedZones != nil {
			thisZoneText = "Last seen " + zonespkg.NameThatZone(NamedZones, zonespkg.Point{Latitude: latestGPSRecord.Lat, Longitude: latestGPSRecord.Long})
		} else {
			thisZoneText = "<No zones loaded>"
		}

		currLocationPoint := poly.Point{X: latestGPSRecord.Lat, Y: latestGPSRecord.Long}
		currInsidePropertyBoundary := poly.IsInside(propertyOutline, currLocationPoint)
		currInsideSafeZoneBoundary := poly.IsInside(safeZoneOutline, currLocationPoint)

		if prevDogBoundaryState[dogName] == nil {
			prevDogBoundaryState[dogName] = make(boundaryStatesType)
		}

		// Notify on changes
		if prevDogBoundaryState[dogName]["propertyBoundary"] && !currInsidePropertyBoundary {
			err = notifier.Notify(ctx, fmt.Sprintf("%s is off the property", dogName), thisZoneText)
			logIfErr(err)
		}

		if !prevDogBoundaryState[dogName]["propertyBoundary"] && currInsidePropertyBoundary {
			err = notifier.Notify(ctx, fmt.Sprintf("%s is now back on property", dogName), thisZoneText)
			logIfErr(err)
		}

		if prevDogBoundaryState[dogName]["safeZoneBoundary"] && !currInsideSafeZoneBoundary {
			err = notifier.Notify(ctx, fmt.Sprintf("%s is getting far from home base", dogName), thisZoneText)
			logIfErr(err)
		}

		if !prevDogBoundaryState[dogName]["safeZoneBoundary"] && currInsideSafeZoneBoundary {
			err = notifier.Notify(ctx, fmt.Sprintf("%s is now back close to home base", dogName), thisZoneText)
			logIfErr(err)
		}

		prevDogBoundaryState[dogName]["propertyBoundary"] = currInsidePropertyBoundary
		prevDogBoundaryState[dogName]["safeZoneBoundary"] = currInsideSafeZoneBoundary

		// Insert the document into storage
		err = storer.WriteCommit(ctx, string(body))
		if err != nil {
			log.Printf("Error inserting document: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// All happy
		log.Printf("Successfully inserted document")
		w.WriteHeader(http.StatusOK)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
	if !lastWasHealthCheck {
		log.Println("Got a health check [repeats will be hidden].")
	}

	lastWasHealthCheck = true
}

func newTestNotifyHandler(n notify.Notifier) func(http.ResponseWriter, *http.Request) {
	notifier := n

	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		err := notifier.Notify(ctx, "Test notification", "This is a test notification.")
		if err != nil {
			log.Printf("Error sending test notification: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
