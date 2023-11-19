package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	//	"go.mongodb.org/mongo-driver/bson"
	"github.com/bitwombat/tag/storage"
	"github.com/bitwombat/tag/sub"
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

var lastWasHealthCheck bool // Used to clean up the log output

// Just to clean up the call - we always use time.Now in a non-test environment.
func timeAgo(timeStr string) string {
	return storage.TimeAgo(timeStr, time.Now)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
	if !lastWasHealthCheck {
		log.Println("Got a health check [repeats will be hidden].")
	}

	lastWasHealthCheck = true
}

func NewMapPageHandler(storer storage.Storage) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got a map page request.")
		lastWasHealthCheck = false

		tags, err := storer.GetLastPositions()
		if err != nil {
			log.Printf("Error getting last position from storage: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		subs := make(map[string]string)

		var idToName = map[float64]string{
			810095: "rueger",
			810243: "tucker",
		}

		for _, tag := range tags {
			name := idToName[tag.SerNo]
			subs[name+"Lat"] = fmt.Sprintf("%.7f", tag.Latitude)
			subs[name+"Long"] = fmt.Sprintf("%.7f", tag.Longitude)
			subs[name+"Note"] = timeAgo(tag.GpsUTC) + " ago"
		}

		mapPage, err := sub.GetContents("public_html/index.html", subs)

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

func NewDataPostHandler(storer storage.Storage) func(http.ResponseWriter, *http.Request) {

	strer := storer // TODO: Is this necessary for a closure?

	return func(w http.ResponseWriter, r *http.Request) {
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

		// Log the records, for debugging
		for _, r := range tagData.Records {
			gpsField := r.Fields[0]
			log.Printf("%v  %s (%s ago) %v  %s (%s ago) %0.7f,%0.7f\n", tagData.SerNo, r.DateUTC, timeAgo(r.DateUTC), r.Reason, gpsField.GpsUTC, timeAgo(gpsField.GpsUTC), gpsField.Lat, gpsField.Long)
		}

		// Insert the document into storage
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		err = strer.WriteCommit(ctx, string(body))
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
		log.Fatal(fmt.Errorf("getting a Mongo connection: %w", err))
	}

	storer := storage.NewMongoStorer(collection)

	log.Println("Connected to MongoDB!")

	log.Println("Setting up handlers.")

	// HTTP endpoints
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/health", handleHealth)

	// HTTPS endpoints
	httpsMux := http.NewServeMux()
	httpsMux.HandleFunc("/map", NewMapPageHandler(storer))
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
