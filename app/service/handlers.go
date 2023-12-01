package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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

const AnalogueDataFType = 6

var idToName = map[float64]string{
	810095: "rueger",
	810243: "tucker",
}

var lastWasHealthCheck bool // Used to clean up the log output.

// A data structure for persistent ["dog"]["boundary"] = true/false states.
type statesType map[string]bool
type dogStatesType map[string]statesType

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

func newDataPostHandler(storer storage.Storage, notifier notify.Notifier, tagAuthKey string) func(http.ResponseWriter, *http.Request) {
	persistentState := make(dogStatesType)

	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		// Validate the request
		if r.Method != http.MethodPost {
			errorLogger.Println("Got a request to /upload that was not a POST")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		authKey := r.Header[http.CanonicalHeaderKey("auth")][0]

		if authKey == "" {
			errorLogger.Printf("Got an empty auth key: %v\n", authKey)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if authKey != tagAuthKey {
			errorLogger.Printf("Got a bad auth key: %v\n", authKey)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		debugLogger.Println("Got a data post.")
		lastWasHealthCheck = false

		// Read and decode the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			errorLogger.Printf("Error reading body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		debugLogger.Println(string(body))

		var tagData TagData

		err = json.Unmarshal(body, &tagData)
		if err != nil {
			errorLogger.Printf("Error unmarshalling JSON: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var thisZoneText string

		NamedZones, err := zonespkg.ReadKMLDir("named_zones")
		if err != nil {
			errorLogger.Printf("Error reading KML files: %v", err)
			// not a critical error, keep going
			return
		}

		dogName, ok := idToName[float64(tagData.SerNo)]
		if !ok {
			errorLogger.Printf("Unknown tag number: %v", tagData.SerNo)
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
				errorLogger.Printf("Error: Unknown reason code: %v\n", r.Reason)
				reason = "Unknown reason"
			}

			infoLogger.Printf("%v/%s  %s (%s ago) \"%v\"  %s (%s ago) %0.7f,%0.7f \"%s\"\n", tagData.SerNo, dogName, r.DateUTC, timeAgoAsText(r.DateUTC), reason, gpsField.GpsUTC, timeAgoAsText(gpsField.GpsUTC), gpsField.Lat, gpsField.Long, thisZoneText)

		}

		// --------------------------------------------------------------------
		// Notify about battery condition -------------------------------------
		// --------------------------------------------------------------------
		var batteryVoltage float64

		for _, f := range latestRecord.Fields {
			if f.FType == AnalogueDataFType {
				batteryVoltage = float64(f.AnalogueData.Num1) / 1000
			}
		}

		if batteryVoltage == 0 {
			debugLogger.Println("No battery voltage in record")
		} else {
			debugLogger.Printf("Battery voltage: %.3f V\n", batteryVoltage)

			if persistentState[dogName] == nil {
				persistentState[dogName] = make(statesType)
			}

			if batteryVoltage < BatteryLowThreshold && !persistentState[dogName]["lowBattery"] {
				err = notifier.Notify(ctx, fmt.Sprintf("%s's battery low", dogName), fmt.Sprintf("Battery voltage: %.3f V", batteryVoltage))
				logIfErr(err)
				if err == nil {
					persistentState[dogName]["lowBattery"] = true
				}
			}

			if batteryVoltage < BatteryCriticalThreshold && !persistentState[dogName]["criticalBattery"] {
				err = notifier.Notify(ctx, fmt.Sprintf("%s's battery critical", dogName), fmt.Sprintf("Battery voltage: %.3f V", batteryVoltage))
				logIfErr(err)
				if err == nil {
					persistentState[dogName]["criticalBattery"] = true
				}
			}

			if batteryVoltage > BatteryLowThreshold+BatteryHysteresis && (persistentState[dogName]["lowBattery"] || persistentState[dogName]["criticalBattery"]) { // The higher of the two thresholds
				// Battery must have been replaced
				persistentState[dogName]["lowBattery"] = false
				persistentState[dogName]["criticalBattery"] = false
				err = notifier.Notify(ctx, fmt.Sprintf("New battery for %s detected", dogName), fmt.Sprintf("Battery voltage: %.3f V", batteryVoltage))
				logIfErr(err)
			}
		}

		// --------------------------------------------------------------------
		// Notify about zones and boundaries ----------------------------------
		// --------------------------------------------------------------------
		latestGPSRecord := latestRecord.Fields[0]

		if NamedZones != nil {
			thisZoneText = "Last seen " + zonespkg.NameThatZone(NamedZones, zonespkg.Point{Latitude: latestGPSRecord.Lat, Longitude: latestGPSRecord.Long})
		} else {
			thisZoneText = "<No zones loaded>"
		}

		currLocationPoint := poly.Point{X: latestGPSRecord.Lat, Y: latestGPSRecord.Long}
		currInsidePropertyBoundary := poly.IsInside(propertyOutline, currLocationPoint)
		currInsideSafeZoneBoundary := poly.IsInside(safeZoneOutline, currLocationPoint)

		// Notify on changes
		if persistentState[dogName]["propertyBoundary"] && !currInsidePropertyBoundary {
			err = notifier.Notify(ctx, fmt.Sprintf("%s is off the property", dogName), thisZoneText)
			logIfErr(err)
		}

		if !persistentState[dogName]["propertyBoundary"] && currInsidePropertyBoundary {
			err = notifier.Notify(ctx, fmt.Sprintf("%s is now back on property", dogName), thisZoneText)
			logIfErr(err)
		}

		if persistentState[dogName]["safeZoneBoundary"] && !currInsideSafeZoneBoundary {
			err = notifier.Notify(ctx, fmt.Sprintf("%s is getting far from home base", dogName), thisZoneText)
			logIfErr(err)
		}

		if !persistentState[dogName]["safeZoneBoundary"] && currInsideSafeZoneBoundary {
			err = notifier.Notify(ctx, fmt.Sprintf("%s is now back close to home base", dogName), thisZoneText)
			logIfErr(err)
		}

		persistentState[dogName]["propertyBoundary"] = currInsidePropertyBoundary
		persistentState[dogName]["safeZoneBoundary"] = currInsideSafeZoneBoundary

		// Insert the document into storage
		id, err := storer.WriteCommit(ctx, string(body))
		if err != nil {
			errorLogger.Printf("Error inserting document: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// All happy
		debugLogger.Print("Successfully inserted document, id: ", id)
		w.WriteHeader(http.StatusOK)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
	if !lastWasHealthCheck {
		debugLogger.Println("Got a health check [repeats will be hidden].")
	}

	lastWasHealthCheck = true
}

func newTestNotifyHandler(n notify.Notifier) func(http.ResponseWriter, *http.Request) {
	notifier := n

	return func(w http.ResponseWriter, r *http.Request) {
		debugLogger.Println("Got a request to send a test notification.")

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		err := notifier.Notify(ctx, "Test notification", "This is a test notification.")
		if err != nil {
			errorLogger.Printf("Error sending test notification: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
