package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/bitwombat/gps-tags/notify"
	"github.com/bitwombat/gps-tags/storage"
)

const BatteryLowThreshold = 4.0
const BatteryCriticalThreshold = 3.8
const BatteryHysteresis = 0.1

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
		fatalLogger.Fatal(fmt.Errorf("getting a Mongo connection: %w", err))
	}

	ntfySubscriptionId := os.Getenv("NTFY_SUBSCRIPTION_ID")
	if ntfySubscriptionId == "" {
		warningLogger.Print("WARNING: NTFY_SUBSCRIPTION_ID not set. Notifications will not be sent.")
	}

	tagAuthKey := os.Getenv("TAG_AUTH_KEY")
	if tagAuthKey == "" {
		fatalLogger.Fatal("TAG_AUTH_KEY not set")
	}

	// Web page endpoints
	httpsMux := http.NewServeMux()

	storer := storage.NewMongoStorer(collection)

	var notifier notify.Notifier
	if os.Getenv("NONOTIFY") != "" {
		warningLogger.Print("WARNING: NONOTIFY env var set. Null notifier being used. No notifications will be sent.")
		notifier = notify.NewNullNotifier()
	} else {
		notifier = notify.NewNtfyNotifier(ntfySubscriptionId)
	}
	loggingNotifier := notify.NewLoggingNotifier(notifier, debugLogger)

	// Current location map page
	httpsMux.HandleFunc("/current", newCurrentMapPageHandler(storer))

	// Paths travelled page
	httpsMux.HandleFunc("/paths", newPathsMapPageHandler(storer))

	// Data upload endpoint
	dataPostHandler := newDataPostHandler(storer, loggingNotifier, tagAuthKey)
	httpsMux.HandleFunc("/upload", dataPostHandler)

	// Notification testing endpoint and aliases
	httpsMux.HandleFunc("/testnotify", newTestNotifyHandler(loggingNotifier))
	httpsMux.HandleFunc("/notifytest", newTestNotifyHandler(loggingNotifier))
	httpsMux.HandleFunc("/testnotification", newTestNotifyHandler(loggingNotifier))
	httpsMux.HandleFunc("/notificationtest", newTestNotifyHandler(loggingNotifier))

	// Health check endpoint (from load balancer)
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/health", handleHealthCheck)

	httpMux.HandleFunc("/owntracks", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		body, _ := io.ReadAll(r.Body)
		debugLogger.Println(string(body))
	})

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
