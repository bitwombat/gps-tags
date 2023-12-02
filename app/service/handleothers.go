package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bitwombat/gps-tags/notify"
)

func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
	if !lastWasHealthCheck {
		debugLogger.Println("Got a health check [repeats will be hidden].")
	}

	lastWasHealthCheck = true
}

func newTestNotifyHandler(n notify.Notifier) func(http.ResponseWriter, *http.Request) {
	notifier := n

	return func(w http.ResponseWriter, r *http.Request) {
		debugLogger.Println("Got a request to send a test notification.")

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		err := notifier.Notify(ctx, "Test notification", "This is a test notification.")
		if err != nil {
			errorLogger.Printf("Error sending test notification: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
