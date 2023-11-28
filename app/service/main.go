package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/bitwombat/gps-tags/notify"
	"github.com/bitwombat/gps-tags/storage"
)

// systemd recognises these prefixes and colors accordingly. Also allows filtering priorities with journalctl.
var fatalLogger = log.New(os.Stdout, "<2>", log.LstdFlags)
var errorLogger = log.New(os.Stdout, "<3>", log.LstdFlags)
var warningLogger = log.New(os.Stdout, "<4>", log.LstdFlags)
var infoLogger = log.New(os.Stdout, "<6>", log.LstdFlags)
var debugLogger = log.New(os.Stdout, "<7>", log.LstdFlags)

// Start the service - connect to Mongo, set up notification, set up endpoints, and start the HTTP servers.
func main() {
	infoLogger.Println("Starting Dog Tag service.")

	mongoURL := os.Getenv("MONGO_URL")
	if mongoURL == "" {
		fatalLogger.Fatal("MONGO_URL not set")
	}

	infoLogger.Println("Connecting to MongoDB.")
	collection, err := storage.NewMongoConnection(mongoURL, "dogs")
	if err != nil {
		fatalLogger.Fatal(fmt.Errorf("getting a Mongo connection: %v", err))
	}

	ntfySubscriptionId := os.Getenv("NTFY_SUBSCRIPTION_ID")
	if ntfySubscriptionId == "" {
		warningLogger.Print("WARNING: NTFY_SUBSCRIPTION_ID not set. Notifications will not be sent.")
	}

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

	// Start servers
	infoLogger.Println("Starting servers.")

	go func() {
		server1 := &http.Server{
			Addr:    ":80",
			Handler: httpMux,
		}
		infoLogger.Println("Server running on :80")
		fatalLogger.Fatal(server1.ListenAndServe())
	}()

	go func() {
		server2 := &http.Server{
			Addr:    ":443",
			Handler: httpsMux,
		}
		infoLogger.Println("Server running on :443")
		fatalLogger.Fatal(server2.ListenAndServe())
	}()

	// Block forever
	select {}
}
