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
	"github.com/bitwombat/tag/storage"
	"github.com/bitwombat/tag/sub"
	"go.mongodb.org/mongo-driver/mongo"
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

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
	if !lastWasHealthCheck {
		log.Println("Got a health check [repeats will be hidden].")
	}

	lastWasHealthCheck = true
}

func readDataFromDisk(filename string) ([]string, error) {
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	fields := strings.Split(string(body), " ")
	if len(fields) != 2 {
		return nil, fmt.Errorf("error in format of file. Length of fields is %d, expected 2", len(fields))
	}

	return fields, nil
}

func handleMapPage(w http.ResponseWriter, r *http.Request) {
	log.Println("Got a map page request.")
	lastWasHealthCheck = false

	subs := make(map[string]string)

	//------
	fields, err := readDataFromDisk("810095")
	if err != nil {
		log.Printf("Error reading %s file: %v\n", "810095", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	subs["ruegerPositions"] = fmt.Sprintf("{lat: %s, lng: %s}", fields[0], fields[1])

	//------
	fields, err = readDataFromDisk("810243")
	if err != nil {
		log.Printf("Error reading %s file: %v\n", "810243", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	subs["tuckerPositions"] = fmt.Sprintf("{lat: %s, lng: %s}", fields[0], fields[1])

	mapPage, err := sub.GetContents("public_html/index.html", subs)

	if err != nil {
		log.Printf("Error getting contents: %v\n", err)
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

func NewDataPostHandler(collection *mongo.Collection) func(http.ResponseWriter, *http.Request) {

	coll := collection // TODO: Understand if this is necessary. Yes because pointer coming in?

	return func(w http.ResponseWriter, r *http.Request) {
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

		log.Printf("Got %d Records, keeping last one\n", len(tagData.Records))

		for i, r := range tagData.Records {
			gpsField := r.Fields[0]
			log.Printf("%v  %s  %v  %s  %0.7f,%0.7f\n", tagData.SerNo, r.DateUTC, r.Reason, gpsField.GpsUTC, gpsField.Lat, gpsField.Long)

			if i == len(tagData.Records)-1 {
				err = os.WriteFile(fmt.Sprintf("%d", tagData.SerNo), []byte(fmt.Sprintf("%0.7f %0.7f", gpsField.Lat, gpsField.Long)), 0644)
			}

			if err != nil {
				log.Printf("Error writing to file: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		_ = coll

		// Unmarshal JSON to map
		var data map[string]interface{}
		err = json.Unmarshal([]byte(body), &data)
		if err != nil {
			log.Printf("Error unmarshaling: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Insert into MongoDB
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		insertResult, err := coll.InsertOne(ctx, data)
		if err != nil {
			log.Printf("Error inserting document: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Println("Inserted document: ", insertResult.InsertedID)

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

	log.Println("Connected to MongoDB!")

	log.Println("Setting up handlers.")

	// HTTP endpoints
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/health", handleHealth)

	// HTTPS endpoints
	httpsMux := http.NewServeMux()
	httpsMux.HandleFunc("/map", handleMapPage)
	dataPostHandler := NewDataPostHandler(collection)
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
