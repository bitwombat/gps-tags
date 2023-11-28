package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/bitwombat/gps-tags/notify"
	"github.com/bitwombat/gps-tags/storage"
)

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

	log.Println("Connected to MongoDB.")

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
