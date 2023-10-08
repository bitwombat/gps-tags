package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/bitwombat/tag/sub"
)

func main() {
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
		log.Println("Got a health check.")
	})

	httpsMux := http.NewServeMux()
	httpsMux.HandleFunc("/map", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got a map page request.")
		mapPage, err := sub.GetContents("public_html/index.html", map[string]string{
			"tuckerPositions": `{lat:-31.457435,lng:152.641985}`,
		})

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
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Println(string(body))

		var result map[string]interface{}

		err = json.Unmarshal(body, result)
		if err != nil {
			log.Printf("Error unmarshalling JSON: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if value, ok := result["SerNo"]; ok {
			log.Printf("Value of 'SerNo': %v\n", value)
		} else {
			fmt.Println("Key 'SerNo' does not exist.")
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
