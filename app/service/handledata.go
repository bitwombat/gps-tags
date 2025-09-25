package main

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/bitwombat/gps-tags/device"
	"github.com/bitwombat/gps-tags/model"
	"github.com/bitwombat/gps-tags/notify"
)

// TxWriter handles writing transmission data.
type TxWriter interface {
	WriteTx(context.Context, model.TagTx) (string, error)
}

var lastWasHealthCheck bool // Used to clean up the log output.

func makeNotifier(ctx context.Context, notifier notify.Notifier, title notify.Title, message notify.Message) func() error {
	return func() error {
		err := notifier.Notify(ctx, title, message)
		if err != nil {
			errorLogger.Printf("error sending notification: %v", err)
		}
		return err
	}
}

func newDataPostHandler(
	storer TxWriter,
	txLogger txLogger,
	batteryNotifier batteryNotifier,
	zoneNotifier zoneNotifier,
	tagAuthKey string,
	now func() time.Time,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
		defer cancel()

		// Validate the request
		if r.Method != http.MethodPost {
			errorLogger.Println("Got a request to /upload that was not a POST")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		auths, ok := r.Header[http.CanonicalHeaderKey("auth")]
		if !ok || len(auths) != 1 {
			errorLogger.Printf("Auth key not set in header, or too many set\n")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		authKey := auths[0]

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

		// TODO: This is temporary while I watch GPS readings come in and try to figure out why it
		// repeats Lat/Lng when it claims to have gotten a good fix (GpsStat = 3).
		debugLogger.Println(string(body))

		tagData, err := device.Unmarshal(body)
		if err != nil {
			errorLogger.Printf("Error unmarshalling transmission body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		id, err := storer.WriteTx(ctx, tagData)
		if err != nil {
			errorLogger.Printf("Error inserting transmission: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		debugLogger.Print("Successfully inserted transmission, id: ", id)

		_, ok = model.SerNoToName[tagData.SerNo]
		if !ok {
			errorLogger.Printf("Unknown tag number: %v", tagData.SerNo)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		cleanGPSReadings(tagData)
		txLogger.Log(now, tagData)
		batteryNotifier.Notify(ctx, now, tagData)
		zoneNotifier.Notify(ctx, tagData)

		w.WriteHeader(http.StatusOK)
	}
}

func cleanGPSReadings(tagData model.TagTx) { // TODO: Maybe move this to the model or device
	for i, r := range tagData.Records {
		if r.GPSReading != nil {
			if r.GPSReading.Lat == 0 || r.GPSReading.Long == 0 { // Oddball, bogus GPS result.
				errorLogger.Print("Got 0 for lat or long... not committing record")
				tagData.Records[i].GPSReading = nil
			}
		}
	}
}
