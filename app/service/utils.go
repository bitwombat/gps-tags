package main

import (
	"log"
	"time"

	"github.com/bitwombat/tag/storage"
)

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
