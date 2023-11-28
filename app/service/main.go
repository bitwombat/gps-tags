package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bitwombat/tag/notify"
	"github.com/bitwombat/tag/poly"
	"github.com/bitwombat/tag/storage"
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

type boundaryStatesType map[string]bool
type dogBoundaryStatesType map[string]boundaryStatesType

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

	ntfy_subscription_id := os.Getenv("NTFY_SUBSCRIPTION_ID")
	if ntfy_subscription_id == "" {
		log.Print("WARNING: NTFY_SUBSCRIPTION_ID not set. Notifications will not be sent.")
	}

	log.Println("Setting up handlers.")

	// HTTP endpoints
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/health", handleHealth)

	// HTTPS endpoints
	httpsMux := http.NewServeMux()
	storer := storage.NewMongoStorer(collection)
	notifier := notify.NewNtfyNotifier(ntfy_subscription_id)
	httpsMux.HandleFunc("/current", NewCurrentMapPageHandler(storer))
	dataPostHandler := NewDataPostHandler(storer, notifier)
	httpsMux.HandleFunc("/upload", dataPostHandler)

	// Static file serving
	fs := http.FileServer(http.Dir("./public_html"))
	httpsMux.Handle("/", fs)

	// Notification testing
	httpsMux.HandleFunc("/testnotify",
		func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()

			err := notifier.Notify(ctx, "Test notification", "This is a test notification.")
			if err != nil {
				log.Printf("Error sending test notification: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
		})

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
