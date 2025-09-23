package main

import (
	"fmt"
	"time"
)

func strTimeAgoAsText(ts string, now func() time.Time) string {
	// Parse the given time string
	t, err := time.Parse(time.DateTime, ts)
	if err != nil {
		return "<time parsing error>"
	}
	return timeAgoAsText(t, now)
}

func timeAgoAsText(t time.Time, now func() time.Time) string {
	// Calculate the difference
	diff := now().Sub(t)

	// Format
	if diff < time.Hour {
		return fmt.Sprintf("%d minutes", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		minutes := int(diff.Minutes()) % 60

		return fmt.Sprintf("%d hours, %d minutes", hours, minutes)
	}

	days := int(diff.Hours()) / 24
	hours := int(diff.Hours()) % 24
	minutes := int(diff.Minutes()) % 60

	return fmt.Sprintf("%d days, %d hours, %d minutes", days, hours, minutes)
}

const (
	aliveColour = "red"
	staleColour = "#a23535"
	deadColour  = "#8d8d8d"
)

func timeAgoInColour(t time.Time, now func() time.Time) string {
	const heartBeatTime = 10 * time.Minute

	// Calculate the difference
	diff := now().Sub(t)

	if diff < heartBeatTime+1*time.Minute { // if it's reported in properly, recently
		return aliveColour
	} else if diff < time.Hour { // somewhat recently, probably working
		return staleColour
	}
	return deadColour // not good
}
