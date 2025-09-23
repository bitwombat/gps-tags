package main

import (
	"context"
	"fmt"

	"github.com/bitwombat/gps-tags/model"
	"github.com/bitwombat/gps-tags/notify"
	oshotpkg "github.com/bitwombat/gps-tags/oneshot"
	"github.com/bitwombat/gps-tags/poly"
	"github.com/bitwombat/gps-tags/zones"
)

type zoneNotifier struct {
	namedZones []zones.Zone
	oneShot    oshotpkg.OneShot
	notifier   notify.Notifier
}

func (zn zoneNotifier) Notify(ctx context.Context, tagData model.TagTx) {
	var latestGPS struct {
		gr    *model.GPSReading
		seqNo int
	}

	for _, r := range tagData.Records {
		if r.GPSReading != nil {
			if r.SeqNo > latestGPS.seqNo {
				latestGPS.seqNo = r.SeqNo
				latestGPS.gr = r.GPSReading
			}
		}
	}

	dogName := model.UpperSerNoToName(tagData.SerNo)
	notifyAboutZones(ctx, latestGPS.gr, zn.namedZones, dogName, zn.oneShot, zn.notifier)
}

func notifyAboutZones(ctx context.Context, latestGPS *model.GPSReading, namedZones []zones.Zone, dogName string, oneShot oshotpkg.OneShot, notifier notify.Notifier) {
	if latestGPS == nil {
		debugLogger.Println("No GPS reading in transmission")

		return
	}

	var thisZoneText string

	if namedZones != nil {
		thisZoneText = "Last seen " + zones.NameThatZone(namedZones, zones.Point{Latitude: latestGPS.Lat, Longitude: latestGPS.Long})
	} else {
		thisZoneText = "<No zones loaded>"
	}

	currentLocation := poly.Point{X: latestGPS.Lat, Y: latestGPS.Long}
	isOutsidePropertyBoundary := !poly.IsInside(propertyBoundary, currentLocation)
	isOutsideSafeZoneBoundary := !poly.IsInside(safeZoneBoundary, currentLocation)

	err := oneShot.SetReset(dogName+"offProperty",
		oshotpkg.Config{
			SetIf: isOutsidePropertyBoundary,
			OnSet: makeNotifier(
				ctx,
				notifier,
				notify.Title(fmt.Sprintf("%s is off the property", dogName)),
				notify.Message(thisZoneText),
			),
			ResetIf: !isOutsidePropertyBoundary,
			OnReset: makeNotifier(
				ctx,
				notifier,
				notify.Title(fmt.Sprintf("%s is back on the property", dogName)),
				notify.Message(thisZoneText),
			),
		})
	if err != nil {
		debugLogger.Println("error when setting: ", err) // notifications are not important enough to return an error.

		return
	}

	err = oneShot.SetReset(dogName+"outsideSafeZone",
		oshotpkg.Config{
			SetIf: isOutsideSafeZoneBoundary,
			OnSet: makeNotifier(
				ctx,
				notifier,
				notify.Title(fmt.Sprintf("%s is getting far from the house", dogName)),
				notify.Message(thisZoneText),
			),
			ResetIf: !isOutsideSafeZoneBoundary,
			OnReset: makeNotifier(
				ctx,
				notifier,
				notify.Title(fmt.Sprintf("%s is back close to the house", dogName)),
				notify.Message(thisZoneText),
			),
		})
	if err != nil {
		debugLogger.Println("error when setting: ", err) // notifications are not important enough to return an error.

		return
	}
}
