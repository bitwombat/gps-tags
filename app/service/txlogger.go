package main

import (
	"time"

	"github.com/bitwombat/gps-tags/model"
	"github.com/bitwombat/gps-tags/zones"
)

type txLogger struct {
	namedZones []zones.Zone
}

func (txl txLogger) Log(now func() time.Time, tagData model.TagTx) {
	dogName := model.UpperSerNoToName(tagData.SerNo)

	for _, r := range tagData.Records {
		var thisZoneText string

		if r.GPSReading != nil {
			thisZoneText = zones.NameThatZone(txl.namedZones, zones.Point{Latitude: r.GPSReading.Lat, Longitude: r.GPSReading.Long})

			infoLogger.Printf("%v/%s  %s (%s ago) \"%v\"  %s (%s ago) %0.7f,%0.7f \"%s\"\n",
				tagData.SerNo,
				dogName,
				r.DateUTC,
				timeAgoAsText(r.DateUTC.T, now),
				r.Reason,
				r.GPSReading.GpsUTC,
				timeAgoAsText(r.GPSReading.GpsUTC.T, now),
				r.GPSReading.Lat,
				r.GPSReading.Long,
				thisZoneText,
			)
		} else {
			infoLogger.Printf("%v/%s  %s (%s ago) \"%v\"\n", tagData.SerNo, dogName, r.DateUTC, timeAgoAsText(r.DateUTC.T, now), r.Reason)
		}
	}
}
