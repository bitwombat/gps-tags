package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/bitwombat/tag/sub"
)

type TagData struct {
	SerNo   int       `json:"SerNo"`
	Imei    string    `json:"IMEI"`
	Iccid   string    `json:"ICCID"`
	ProdID  int       `json:"ProdId"`
	Fw      string    `json:"FW"`
	Records []Records `json:"Records"`
}

type AnalogueData struct {
	Num1 int `json:"1"`
	Num3 int `json:"3"`
	Num4 int `json:"4"`
	Num5 int `json:"5"`
}

type Fields struct {
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

type Records struct {
	SeqNo   int      `json:"SeqNo"`
	Reason  int      `json:"Reason"`
	DateUTC string   `json:"DateUTC"`
	Fields  []Fields `json:"Fields"`
}

var lastWasHealthCheck bool

func main() {

	lastWasHealthCheck = false

	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
		if !lastWasHealthCheck {
			log.Println("Got a health check [repeats will be hidden].")
		}

		lastWasHealthCheck = true
	})

	httpsMux := http.NewServeMux()
	httpsMux.HandleFunc("/map", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got a map page request.")
		lastWasHealthCheck = false

		//------
		body, err := ioutil.ReadFile("810095")
		if err != nil {
			log.Printf("Error reading 810095 (Rueger) file: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		subs := make(map[string]string)

		fields := strings.Split(string(body), " ")
		if len(fields) != 2 {
			log.Printf("Error in format of 810095 file. Length of fields is %d, expected 2", len(fields))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		subs["ruegerPositions"] = fmt.Sprintf("{lat: %s, lng: %s}", fields[0], fields[1])

		//------
		body, err = ioutil.ReadFile("810243")
		if err != nil {
			log.Printf("Error reading 810243 (Tucker) file: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		fields = strings.Split(string(body), " ")
		if len(fields) != 2 {
			log.Printf("Error in format of 810243 file. Length of fields is %d, expected 2", len(fields))
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

		w.WriteHeader(http.StatusOK)

		_, err = w.Write([]byte(mapPage))
		if err != nil {
			log.Printf("Error writing response: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	})

	httpsMux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
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

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Println(string(body))

		var result TagData

		err = json.Unmarshal(body, &result)
		if err != nil {
			log.Printf("Error unmarshalling JSON: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Println(result.SerNo)
		log.Printf("Got %d Records, keeping last one\n", len(result.Records))

		for i, r := range result.Records {
			log.Println(r.DateUTC)
			gpsField := r.Fields[0]
			log.Printf("   %0.7f, %0.7f\n", gpsField.Lat, gpsField.Long)

			if i == len(result.Records)-1 {
				err = ioutil.WriteFile(fmt.Sprintf("%d", result.SerNo), []byte(fmt.Sprintf("%0.7f %0.7f", gpsField.Lat, gpsField.Long)), 0644)
			}

			if err != nil {
				log.Printf("Error writing to file: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	})

	fs := http.FileServer(http.Dir("./public_html"))
	httpsMux.Handle("/", fs)

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

	select {} // Block forever
}
