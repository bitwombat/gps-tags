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
	"github.com/bitwombat/gps-tags/model"
	"github.com/bitwombat/gps-tags/notify"
	oshotpkg "github.com/bitwombat/gps-tags/oneshot"
	"github.com/bitwombat/gps-tags/poly"
	"github.com/bitwombat/gps-tags/storage"
	zonespkg "github.com/bitwombat/gps-tags/zones"
)

var lastWasHealthCheck bool // Used to clean up the log output.

func makeNotifier(ctx context.Context, notifier notify.Notifier, title, message string) func() error {
	return func() error {
		err := notifier.Notify(ctx, title, message)
		logIfErr(err)
		return err
	}
}

func newDataPostHandler(storer storage.Storage, notifier notify.Notifier, tagAuthKey string, now func() time.Time) func(http.ResponseWriter, *http.Request) {
	oneShot := oshotpkg.NewOneShot()

	NamedZones, err := zonespkg.ReadKMLDir("named_zones")
	if err != nil {
		errorLogger.Printf("Error reading KML files: %v", err)
		// not a critical error, keep going
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
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

		var tagData device.TagTx

		err = json.Unmarshal(body, &tagData)
		if err != nil {
			errorLogger.Printf("Error unmarshalling JSON: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		dogName, ok := device.SerNoToName[tagData.SerNo]
		if !ok {
			errorLogger.Printf("Unknown tag number: %v", tagData.SerNo)
		}
		dogName = strings.ToUpper(dogName) // Just looks better and stands out in notifications

		var latestRecord device.Record

		// Process the records
		for _, r := range tagData.Records {
			var thisZoneText string

			// Figure out the most recent record for notifications later
			if r.SeqNo > latestRecord.SeqNo { // TODO: Totally wrong. Need to independently track the Reading* fields.
				latestRecord = r
			}

			if r.GPSReading != nil {
				thisZoneText = zonespkg.NameThatZone(NamedZones, zonespkg.Point{Latitude: r.GPSReading.Lat, Longitude: r.GPSReading.Long})

				infoLogger.Printf("%v/%s  %s (%s ago) \"%v\"  %s (%s ago) %0.7f,%0.7f \"%s\"\n", tagData.SerNo, dogName, r.DateUTC, storage.StrTimeAgoAsText(r.DateUTC, now), model.ReasonCode(r.Reason), r.GPSReading.GpsUTC, storage.StrTimeAgoAsText(r.GPSReading.GpsUTC, now), r.GPSReading.Lat, r.GPSReading.Long, thisZoneText)
			} else {
				infoLogger.Printf("%v/%s  %s (%s ago) \"%v\"\n", tagData.SerNo, dogName, r.DateUTC, storage.StrTimeAgoAsText(r.DateUTC, now), model.ReasonCode(r.Reason))
			}

		}

		if latestRecord.GPSReading.Lat == 0 || latestRecord.GPSReading.Long == 0 { // Oddball, bogus GPS result.
			errorLogger.Print("Got 0 for lat or long... not committing record")
			// TODO: nil out the pointer in the Record, rather than returning
			// early - there are other records to go through in the loop
			w.WriteHeader(http.StatusOK) // Say "OK" because we don't want the system re-sending this record.  // TODO: Delete this
			return
		}

		// Send notifications
		notifyAboutBattery(ctx, latestRecord, dogName, oneShot, notifier)
		notifyAboutZones(ctx, latestRecord, NamedZones, dogName, oneShot, notifier)

		// Insert the document into storage
		id, err := storer.WriteTx(ctx, tagData)
		if err != nil {
			errorLogger.Printf("Error inserting transmission: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// All happy
		debugLogger.Print("Successfully inserted transmission, id: ", id)
		w.WriteHeader(http.StatusOK)
	}
}

func notifyAboutBattery(ctx context.Context, latestRecord device.Record, dogName string, oneShot oshotpkg.OneShot, notifier notify.Notifier) {
	var batteryVoltage float64

	batteryVoltage = float64(latestRecord.AnalogueReading.AnalogueData.Num1) / 1000 // TODO: Remove this extra level of structure.

	// We don't want to hear about low battery in the middle of the night.
	nowIsWakingHours := time.Now().Hour() >= 8 && time.Now().Hour() <= 22

	err := oneShot.SetReset(dogName+"lowBattery",
		oshotpkg.Config{
			SetIf:   (batteryVoltage < BatteryLowThreshold) && nowIsWakingHours,
			OnSet:   makeNotifier(ctx, notifier, fmt.Sprintf("%s's battery low", dogName), fmt.Sprintf("Battery voltage: %.3f V", batteryVoltage)),
			ResetIf: batteryVoltage > BatteryLowThreshold+BatteryHysteresis,
			OnReset: makeNotifier(ctx, notifier, fmt.Sprintf("New battery for %s detected", dogName), fmt.Sprintf("Battery voltage: %.3f V", batteryVoltage)),
		})
	if err != nil {
		debugLogger.Println("error when setting: ", err) // TODO: Should this return an error?

		return
	}

	err = oneShot.SetReset(dogName+"criticalBattery", // TODO: Don't ignore return value
		oshotpkg.Config{
			SetIf:   (batteryVoltage < BatteryCriticalThreshold) && nowIsWakingHours,
			OnSet:   makeNotifier(ctx, notifier, fmt.Sprintf("%s's battery critical", dogName), fmt.Sprintf("Battery voltage: %.3f V", batteryVoltage)),
			ResetIf: batteryVoltage > BatteryLowThreshold,
		})
	if err != nil {
		debugLogger.Println("error when setting: ", err) // TODO: Should this return an error?

		return
	}
}

func notifyAboutZones(ctx context.Context, latestRecord device.Record, namedZones []zonespkg.Zone, dogName string, oneShot oshotpkg.OneShot, notifier notify.Notifier) {
	if latestRecord.GPSReading == nil {
		debugLogger.Println("No GPS reading in record") // TODO: Should this return an error?

		return
	}

	var thisZoneText string

	if namedZones != nil {
		thisZoneText = "Last seen " + zonespkg.NameThatZone(namedZones, zonespkg.Point{Latitude: latestRecord.GPSReading.Lat, Longitude: latestRecord.GPSReading.Long})
	} else {
		thisZoneText = "<No zones loaded>"
	}

	currentLocation := poly.Point{X: latestRecord.GPSReading.Lat, Y: latestRecord.GPSReading.Long}
	isOutsidePropertyBoundary := !poly.IsInside(propertyBoundary, currentLocation)
	isOutsideSafeZoneBoundary := !poly.IsInside(safeZoneBoundary, currentLocation)

	err := oneShot.SetReset(dogName+"offProperty",
		oshotpkg.Config{
			SetIf:   isOutsidePropertyBoundary,
			OnSet:   makeNotifier(ctx, notifier, fmt.Sprintf("%s is off the property", dogName), thisZoneText),
			ResetIf: !isOutsidePropertyBoundary,
			OnReset: makeNotifier(ctx, notifier, fmt.Sprintf("%s is now back on the property", dogName), thisZoneText),
		})
	if err != nil {
		debugLogger.Println("error when setting: ", err) // TODO: Should this return an error?

		return
	}

	err = oneShot.SetReset(dogName+"outsideSafeZone",
		oshotpkg.Config{
			SetIf:   isOutsideSafeZoneBoundary,
			OnSet:   makeNotifier(ctx, notifier, fmt.Sprintf("%s is getting far from the house", dogName), thisZoneText),
			ResetIf: !isOutsideSafeZoneBoundary,
			OnReset: makeNotifier(ctx, notifier, fmt.Sprintf("%s is now back close to the house", dogName), thisZoneText),
		})
	if err != nil {
		debugLogger.Println("error when setting: ", err) // TODO: Should this return an error?

		return
	}
}
