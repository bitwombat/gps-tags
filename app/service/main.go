package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/bitwombat/tag/notify"
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

var idToName = map[float64]string{
	810095: "rueger",
	810243: "tucker",
}

type boundaryStatesType map[string]bool
type dogBoundaryStatesType map[string]boundaryStatesType

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

	ntfySubscriptionId := os.Getenv("NTFY_SUBSCRIPTION_ID")
	if ntfySubscriptionId == "" {
		log.Print("WARNING: NTFY_SUBSCRIPTION_ID not set. Notifications will not be sent.")
	}

	log.Println("Setting up handlers.")

	// Web page endpoints
	httpsMux := http.NewServeMux()
	storer := storage.NewMongoStorer(collection)
	notifier := notify.NewNtfyNotifier(ntfySubscriptionId)
	httpsMux.HandleFunc("/current", newCurrentMapPageHandler(storer))

	// Data upload endpoint
	dataPostHandler := newDataPostHandler(storer, notifier)
	httpsMux.HandleFunc("/upload", dataPostHandler)

	// Notification testing endpoint
	httpsMux.HandleFunc("/testnotify", newTestNotifyHandler(notifier))

	// Health check endpoint (from load balancer)
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/health", handleHealth)

	// Static file serving
	fs := http.FileServer(http.Dir("./public_html"))
	httpsMux.Handle("/", fs)

	// Servers
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
