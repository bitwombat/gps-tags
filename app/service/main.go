package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bitwombat/gps-tags/notify"
	"github.com/bitwombat/gps-tags/storage"
)

const (
	BatteryLowThreshold      = 4.0
	BatteryCriticalThreshold = 3.8
	BatteryHysteresis        = 0.1
)

// systemd recognises these prefixes and colors accordingly. Also allows filtering priorities with journalctl.
var (
	fatalLogger   = log.New(os.Stdout, "<2>", log.LstdFlags)
	errorLogger   = log.New(os.Stdout, "<3>", log.LstdFlags)
	warningLogger = log.New(os.Stdout, "<4>", log.LstdFlags)
	infoLogger    = log.New(os.Stdout, "<6>", log.LstdFlags)
	debugLogger   = log.New(os.Stdout, "<7>", log.LstdFlags)
)

func hostnameBasedFileServer() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Determine the directory to serve filess from based on the hostname in
		// the request.
		host := strings.ToLower(r.Host)
		var dir string
		switch host {
		case "tags.bitwombat.com.au":
			dir = "./public_html"
		case "photos.bitwombat.com.au":
			dir = "./public_html.photos"
		default:
			dir = "./public_html"
		}

		// Create a new file server for that directory
		fs := http.FileServer(http.Dir(dir))
		// Serve using the file server
		fs.ServeHTTP(w, r)
	})
}

// Start the service - connect to Mongo, set up notification, set up endpoints,
// and start the HTTP servers.
func main() {
	infoLogger.Println("Starting Dog Tag service.")

	storer, err := storage.NewSQLiteStorer("dogs") // TODO: Get consistent between "dogs", "tags" and "dogtags" throughout codebase, including SQL
	if err != nil {
		fatalLogger.Fatal(fmt.Errorf("getting an sqlite storer: %w", err))
	}

	ntfySubscriptionID := os.Getenv("NTFY_SUBSCRIPTION_ID")
	if ntfySubscriptionID == "" {
		warningLogger.Print("WARNING: NTFY_SUBSCRIPTION_ID not set. Notifications will not be sent.")
	}

	// Set up endpoints
	httpsMux := http.NewServeMux()

	var notifier notify.Notifier
	if os.Getenv("NONOTIFY") != "" {
		warningLogger.Print("WARNING: NONOTIFY env var set. Null notifier being used. No notifications will be sent.")
		notifier = notify.NewNullNotifier()
	} else {
		notifier = notify.NewNtfyNotifier(ntfySubscriptionID)
	}
	loggingNotifier := notify.NewLoggingNotifier(notifier, debugLogger)

	tagAuthKey := os.Getenv("TAG_AUTH_KEY")
	if tagAuthKey == "" {
		fatalLogger.Fatal("TAG_AUTH_KEY not set")
	}

	// Current location map page
	httpsMux.HandleFunc("/current", newCurrentMapPageHandler(storer, time.Now))

	// Paths travelled page
	httpsMux.HandleFunc("/paths", newPathsMapPageHandler(storer, time.Now))

	// Data upload endpoint
	dataPostHandler := newDataPostHandler(storer, loggingNotifier, tagAuthKey, time.Now)
	httpsMux.HandleFunc("/upload", dataPostHandler)

	// Notification testing endpoint and aliases
	httpsMux.HandleFunc("/testnotify", newTestNotifyHandler(loggingNotifier))
	httpsMux.HandleFunc("/notifytest", newTestNotifyHandler(loggingNotifier))
	httpsMux.HandleFunc("/testnotification", newTestNotifyHandler(loggingNotifier))
	httpsMux.HandleFunc("/notificationtest", newTestNotifyHandler(loggingNotifier))

	// Health check endpoint (from load balancer)
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/health", handleHealthCheck)

	// Static file serving
	httpsMux.Handle("/", hostnameBasedFileServer())

	// Start servers
	infoLogger.Println("Starting servers.")

	go func() {
		server1 := &http.Server{
			Addr:              ":80",
			Handler:           httpMux,
			ReadHeaderTimeout: 10 * time.Second,
		}
		infoLogger.Println("Server running on :80")
		fatalLogger.Fatal(server1.ListenAndServe())
	}()

	go func() {
		server2 := &http.Server{
			Addr:              ":443",
			Handler:           httpsMux,
			ReadHeaderTimeout: 10 * time.Second,
		}
		infoLogger.Println("Server running on :443")
		fatalLogger.Fatal(server2.ListenAndServe())
	}()

	// Block forever
	select {}
}
