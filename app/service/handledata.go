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
	"github.com/bitwombat/gps-tags/oneshot"
	"github.com/bitwombat/gps-tags/poly"
	"github.com/bitwombat/gps-tags/storage"
	zonespkg "github.com/bitwombat/gps-tags/zones"
)

var lastWasHealthCheck bool // Used to clean up the log output.

func makeNotifier(notifier notify.Notifier, ctx context.Context, title, message string) func() error {
	return func() error {
		err := notifier.Notify(ctx, title, message)
		logIfErr(err)
		return err
	}
}

func newDataPostHandler(storer storage.Storage, notifier notify.Notifier, tagAuthKey string) func(http.ResponseWriter, *http.Request) {
	persistentState := make(map[string]bool)

	NamedZones, err := zonespkg.ReadKMLDir("named_zones")
	if err != nil {
		errorLogger.Printf("Error reading KML files: %v", err)
		// not a critical error, keep going
	}

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

		dogName, ok := idToName[float64(tagData.SerNo)]
		if !ok {
			errorLogger.Printf("Unknown tag number: %v", tagData.SerNo)
		}
		dogName = strings.ToUpper(dogName) // Just looks better and stands out in notifications

		var latestRecord Record

		// Process the records
		for _, r := range tagData.Records {
			var thisZoneText string

			// Figure out the most recent record for notifications later
			if r.SeqNo > latestRecord.SeqNo {
				latestRecord = r
			}

			gpsField := r.Fields[0]

			thisZoneText = zonespkg.NameThatZone(NamedZones, zonespkg.Point{Latitude: gpsField.Lat, Longitude: gpsField.Long})

			reason, ok := device.ReasonToText[int64(r.Reason)]
			if !ok {
				errorLogger.Printf("Error: Unknown reason code: %v\n", r.Reason)
				reason = "Unknown reason"
			}

			infoLogger.Printf("%v/%s  %s (%s ago) \"%v\"  %s (%s ago) %0.7f,%0.7f \"%s\"\n", tagData.SerNo, dogName, r.DateUTC, timeAgoAsText(r.DateUTC), reason, gpsField.GpsUTC, timeAgoAsText(gpsField.GpsUTC), gpsField.Lat, gpsField.Long, thisZoneText)

		}

		// Send notifications
		notifyAboutBattery(ctx, latestRecord, dogName, persistentState, notifier)
		notifyAboutZones(ctx, latestRecord, NamedZones, dogName, persistentState, notifier)

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

func notifyAboutBattery(ctx context.Context, latestRecord Record, dogName string, persistentState map[string]bool, notifier notify.Notifier) {
	var batteryVoltage float64

	for _, f := range latestRecord.Fields {
		if f.FType == AnalogueDataFType {
			batteryVoltage = float64(f.AnalogueData.Num1) / 1000
		}
	}

	// We don't want to hear about low battery in the middle of the night.
	nowIsWakingHours := time.Now().Hour() >= 8 && time.Now().Hour() <= 22

	if batteryVoltage == 0 {
		debugLogger.Println("No battery voltage in record")
	} else {
		debugLogger.Printf("Battery voltage: %.3f V\n", batteryVoltage)

		_ = oneshot.SetOrReset(dogName+"lowBattery", persistentState,
			oneshot.Config{
				SetIf:   (batteryVoltage < BatteryLowThreshold) && nowIsWakingHours,
				OnSet:   makeNotifier(notifier, ctx, fmt.Sprintf("%s's battery low", dogName), fmt.Sprintf("Battery voltage: %.3f V", batteryVoltage)),
				ResetIf: batteryVoltage > BatteryLowThreshold+BatteryHysteresis,
				OnReset: makeNotifier(notifier, ctx, fmt.Sprintf("New battery for %s detected", dogName), fmt.Sprintf("Battery voltage: %.3f V", batteryVoltage)),
			})

		_ = oneshot.SetOrReset(dogName+"criticalBattery", persistentState,
			oneshot.Config{
				SetIf:   (batteryVoltage < BatteryCriticalThreshold) && nowIsWakingHours,
				OnSet:   makeNotifier(notifier, ctx, fmt.Sprintf("%s's battery critical", dogName), fmt.Sprintf("Battery voltage: %.3f V", batteryVoltage)),
				ResetIf: batteryVoltage > BatteryLowThreshold,
			})

	}
}

func notifyAboutZones(ctx context.Context, latestRecord Record, NamedZones []zonespkg.Zone, dogName string, persistentState map[string]bool, notifier notify.Notifier) {
	latestGPSRecord := latestRecord.Fields[0]

	var thisZoneText string

	if NamedZones != nil {
		thisZoneText = "Last seen " + zonespkg.NameThatZone(NamedZones, zonespkg.Point{Latitude: latestGPSRecord.Lat, Longitude: latestGPSRecord.Long})
	} else {
		thisZoneText = "<No zones loaded>"
	}

	currentLocation := poly.Point{X: latestGPSRecord.Lat, Y: latestGPSRecord.Long}
	isOutsidePropertyBoundary := !poly.IsInside(propertyBoundary, currentLocation)
	isOutsideSafeZoneBoundary := !poly.IsInside(safeZoneBoundary, currentLocation)

	_ = oneshot.SetOrReset(dogName+"offProperty", persistentState,
		oneshot.Config{
			SetIf:   isOutsidePropertyBoundary,
			OnSet:   makeNotifier(notifier, ctx, fmt.Sprintf("%s is off the property", dogName), thisZoneText),
			ResetIf: !isOutsidePropertyBoundary,
			OnReset: makeNotifier(notifier, ctx, fmt.Sprintf("%s is now back on the property", dogName), thisZoneText),
		})

	_ = oneshot.SetOrReset(dogName+"outsideSafeZone", persistentState,
		oneshot.Config{
			SetIf:   isOutsideSafeZoneBoundary,
			OnSet:   makeNotifier(notifier, ctx, fmt.Sprintf("%s is getting far from the house", dogName), thisZoneText),
			ResetIf: !isOutsideSafeZoneBoundary,
			OnReset: makeNotifier(notifier, ctx, fmt.Sprintf("%s is now back close to the house", dogName), thisZoneText),
		})
}
