package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	//	"go.mongodb.org/mongo-driver/bson"
	"github.com/bitwombat/tag/notify"
	"github.com/bitwombat/tag/poly"
	"github.com/bitwombat/tag/storage"
	"github.com/bitwombat/tag/substitute"
	zonespkg "github.com/bitwombat/tag/zones"
)

type TagData struct {
	SerNo   int      `json:"SerNo"`
	Imei    string   `json:"IMEI"`
	Iccid   string   `json:"ICCID"`
	ProdID  int      `json:"ProdId"`
	Fw      string   `json:"FW"`
	Records []Record `json:"Records"`
}

type AnalogueData struct {
	Num1 int `json:"1"`
	Num3 int `json:"3"`
	Num4 int `json:"4"`
	Num5 int `json:"5"`
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

type Record struct {
	SeqNo   int     `json:"SeqNo"`
	Reason  int     `json:"Reason"`
	DateUTC string  `json:"DateUTC"`
	Fields  []Field `json:"Fields"`
}

var propertyOutline = []poly.Point{
	{X: -31.4586212322512, Y: 152.6422124774594},
	{X: -31.4595509701308, Y: 152.6438560831193},
	{X: -31.45812972583087, Y: 152.6451090582995},
	{X: -31.45580841978974, Y: 152.6409669973841},
	{X: -31.45613159545191, Y: 152.6404602174576},
	{X: -31.4586212322512, Y: 152.6422124774594},
}

var safeZoneOutline = []poly.Point{
	{X: -31.45682907060356, Y: 152.6423896947185},
	{X: -31.4566607599576, Y: 152.6418404324234},
	{X: -31.45698678814134, Y: 152.641332228427},
	{X: -31.45717818300943, Y: 152.6413883863712},
	{X: -31.45785626006774, Y: 152.6418268829714},
	{X: -31.4583193861515, Y: 152.6422250491253},
	{X: -31.45863308808003, Y: 152.6432598845506},
	{X: -31.45757700831948, Y: 152.6435985457414},
	{X: -31.45682907060356, Y: 152.6423896947185},
}

var reasonToText = map[int64]string{
	1:  "Start of trip",
	2:  "End of trip",
	3:  "Elapsed time",
	6:  "Distance travelled",
	11: "Heartbeat",
}

var idToName = map[float64]string{
	810095: "rueger",
	810243: "tucker",
}

var lastWasHealthCheck bool // Used to clean up the log output. TODO: Same persistence problem here.

func logIfErr(err error) {
	if err != nil {
		log.Printf("error sending notification: %v", err)
	}
}

// Just to clean up the call - we always use time.Now in a non-test environment.
func timeAgoAsText(timeStr string) string {
	return storage.TimeAgoAsText(timeStr, time.Now)
}

func timeAgoInColour(timeStr string) string {
	return storage.TimeAgoInColour(timeStr, time.Now)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
	if !lastWasHealthCheck {
		log.Println("Got a health check [repeats will be hidden].")
	}

	lastWasHealthCheck = true
}

func NewCurrentMapPageHandler(storer storage.Storage) func(http.ResponseWriter, *http.Request) {

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
			reason, ok := reasonToText[tag.Reason]
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

func NewDataPostHandler(s storage.Storage) func(http.ResponseWriter, *http.Request) {
	storer := s // TODO: Is this necessary for a closure?

	prevInsideSafeZoneBoundary := true // TODO: this saved state won't work if on CloudRun, since the process comes and goes!
	prevInsidePropertyBoundary := true

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

		log.Println("Got a data post!")
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

			reason, ok := reasonToText[int64(r.Reason)]
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

		// Notify on changes
		if prevInsidePropertyBoundary && !currInsidePropertyBoundary {
			err = notify.Notify(ctx, fmt.Sprintf("%s is off the property", dogName), thisZoneText)
			logIfErr(err)
		}

		if !prevInsidePropertyBoundary && currInsidePropertyBoundary {
			err = notify.Notify(ctx, fmt.Sprintf("%s is now back on property", dogName), thisZoneText)
			logIfErr(err)
		}

		if prevInsideSafeZoneBoundary && !currInsideSafeZoneBoundary {
			err = notify.Notify(ctx, fmt.Sprintf("%s is getting far from home base", dogName), thisZoneText)
			logIfErr(err)
		}

		if !prevInsideSafeZoneBoundary && currInsideSafeZoneBoundary {
			err = notify.Notify(ctx, fmt.Sprintf("%s is now back close to home base", dogName), thisZoneText)
			logIfErr(err)
		}

		prevInsidePropertyBoundary = currInsidePropertyBoundary
		prevInsideSafeZoneBoundary = currInsideSafeZoneBoundary

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

func main() {
	log.Println("Starting Dog Tag application.")

	lastWasHealthCheck = false

	mongoURL := os.Getenv("MONGO_URL")
	if mongoURL == "" {
		log.Fatal("MONGO_URL not set")
	}

	collection, err := storage.NewMongoConnection(mongoURL, "dogs")
	if err != nil {
		log.Fatal(fmt.Errorf("getting a Mongo connection: %v", err))
	}

	log.Println("Connected to MongoDB!")

	log.Println("Setting up handlers.")

	// HTTP endpoints
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/health", handleHealth)

	// HTTPS endpoints
	httpsMux := http.NewServeMux()
	storer := storage.NewMongoStorer(collection)
	httpsMux.HandleFunc("/current", NewCurrentMapPageHandler(storer))
	dataPostHandler := NewDataPostHandler(storer)
	httpsMux.HandleFunc("/upload", dataPostHandler)

	// Static file serving
	fs := http.FileServer(http.Dir("./public_html"))
	httpsMux.Handle("/", fs)

	log.Println("Starting servers")
	go func() {
		server1 := &http.Server{
			Addr:    ":80",
			Handler: httpMux,
		}
		log.Println("Server running on :80")
		log.Fatal(server1.ListenAndServe())
	}()

	go func() {
		server2 := &http.Server{
			Addr:    ":443",
			Handler: httpsMux,
		}
		log.Println("Server running on :443")
		log.Fatal(server2.ListenAndServe())
	}()

	// Block forever
	select {}
}
