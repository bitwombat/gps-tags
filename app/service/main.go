package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bitwombat/gps-tags/notify"
	oshotpkg "github.com/bitwombat/gps-tags/oneshot"
	"github.com/bitwombat/gps-tags/storage"
	zonespkg "github.com/bitwombat/gps-tags/zones"
	"golang.org/x/sync/errgroup"
)

const (
	BatteryLowThreshold      = 4.0
	BatteryCriticalThreshold = 3.8
	BatteryHysteresis        = 0.1
)

// systemd recognises these prefixes and colors accordingly. Also allows filtering priorities with journalctl.
var (
	fatalLog = func(errcode int, msg string) int {
		log.New(os.Stdout, "<2>", log.LstdFlags).Print(msg)

		return errcode
	}
	errorLogger   = log.New(os.Stdout, "<3>", log.LstdFlags)
	warningLogger = log.New(os.Stdout, "<4>", log.LstdFlags)
	infoLogger    = log.New(os.Stdout, "<6>", log.LstdFlags)
	debugLogger   = log.New(os.Stdout, "<7>", log.LstdFlags)
)

func hostnameBasedFileServer() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Determine the directory to serve files from based on the hostname in
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

func main() {
	os.Exit(run())
}

// Start the service - connect to Mongo, set up notification, set up endpoints,
// and start the HTTP servers.
func run() int {
	infoLogger.Println("Starting Dog Tag service.")

	storer, err := storage.NewSQLiteStorer("dogtags")
	if err != nil {
		return fatalLog(1, fmt.Sprintf("getting an sqlite storer: %v", err))
	}

	ntfySubscriptionID := os.Getenv("NTFY_SUBSCRIPTION_ID")
	if ntfySubscriptionID == "" {
		warningLogger.Print("WARNING: NTFY_SUBSCRIPTION_ID not set. Notifications will not be sent.")
	}

	// Set up endpoints
	httpsMux := http.NewServeMux()

	tagAuthKey := os.Getenv("TAG_AUTH_KEY")
	if tagAuthKey == "" {
		return fatalLog(1, "TAG_AUTH_KEY not set")
	}

	// Current location map page
	httpsMux.HandleFunc("/current", newCurrentMapPageHandler(storer, time.Now))

	// Paths travelled page
	httpsMux.HandleFunc("/paths", newPathsMapPageHandler(storer))

	var notifier notify.Notifier
	if os.Getenv("NONOTIFY") != "" {
		warningLogger.Print("WARNING: NONOTIFY env var set. Null notifier being used. No notifications will be sent.")
		notifier = notify.NewNullNotifier()
	} else {
		notifier = notify.NewNtfyNotifier(ntfySubscriptionID)
	}
	loggingNotifier := notify.NewLoggingNotifier(notifier, debugLogger)

	oneShot := oshotpkg.NewOneShot()

	namedZones, err := zonespkg.ReadKMLDir("named_zones")
	if err != nil {
		errorLogger.Printf("Error reading KML files: %v", err)
		// not a critical error, keep going
	}

	txLogger := txLogger{namedZones}

	batteryNotifier := batteryNotifier{
		namedZones: namedZones,
		oneShot:    oneShot,
		notifier:   loggingNotifier,
	}

	zoneNotifier := zoneNotifier{
		namedZones: namedZones,
		oneShot:    oneShot,
		notifier:   loggingNotifier,
	}

	// Data upload endpoint
	dataPostHandler := newDataPostHandler(storer, txLogger, batteryNotifier, zoneNotifier, tagAuthKey, time.Now)
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

	var g errgroup.Group

	g.Go(func() error {
		server1 := &http.Server{
			Addr:              ":80",
			Handler:           httpMux,
			ReadHeaderTimeout: 10 * time.Second,
		}
		infoLogger.Println("Server running on :80")

		return server1.ListenAndServe()
	})

	g.Go(func() error {
		server2 := &http.Server{
			Addr:              ":443",
			Handler:           httpsMux,
			ReadHeaderTimeout: 10 * time.Second,
		}
		infoLogger.Println("Server running on :443")
		return server2.ListenAndServe()
	})

	if err := g.Wait(); err != nil {
		return fatalLog(2, fmt.Sprintf("error starting one of the servers: %v", err))
	}

	// Block forever
	select {}
}
