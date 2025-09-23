package main

import (
	"context"
	"fmt"
	"time"

	"github.com/bitwombat/gps-tags/model"
	"github.com/bitwombat/gps-tags/notify"
	oshotpkg "github.com/bitwombat/gps-tags/oneshot"
	"github.com/bitwombat/gps-tags/zones"
)

type batteryNotifier struct {
	namedZones []zones.Zone
	oneShot    oshotpkg.OneShot
	notifier   notify.Notifier
}

func (bn batteryNotifier) Notify(ctx context.Context, now func() time.Time, tagData model.TagTx) {
	var latestAnalogue struct {
		ar    *model.AnalogueReading
		seqNo int
	}

	for _, r := range tagData.Records {
		if r.AnalogueReading != nil {
			if r.SeqNo > latestAnalogue.seqNo {
				latestAnalogue.seqNo = r.SeqNo
				latestAnalogue.ar = r.AnalogueReading
			}
		}
	}

	dogName := model.UpperSerNoToName(tagData.SerNo)
	notifyAboutBattery(ctx, now, latestAnalogue.ar, dogName, bn.oneShot, bn.notifier)
}

func notifyAboutBattery(ctx context.Context, now func() time.Time, latestAnalogue *model.AnalogueReading, dogName string, oneShot oshotpkg.OneShot, notifier notify.Notifier) {
	if latestAnalogue == nil {
		debugLogger.Println("No Analogue reading in transmission")

		return
	}

	batteryVoltage := float64(latestAnalogue.InternalBatteryVoltage) / 1000

	// We don't want to hear about low battery in the middle of the night.
	nowIsWakingHours := now().Hour() >= 8 && now().Hour() <= 22

	err := oneShot.SetReset(dogName+"lowBattery",
		oshotpkg.Config{
			SetIf: (batteryVoltage < BatteryLowThreshold) && nowIsWakingHours,
			OnSet: makeNotifier(
				ctx,
				notifier,
				notify.Title(fmt.Sprintf("%s's battery low", dogName)),
				notify.Message(fmt.Sprintf("Battery voltage: %.3f V", batteryVoltage)),
			),
			ResetIf: batteryVoltage > BatteryLowThreshold+BatteryHysteresis,
			OnReset: makeNotifier(
				ctx,
				notifier,
				notify.Title(fmt.Sprintf("New battery for %s detected", dogName)),
				notify.Message(fmt.Sprintf("Battery voltage: %.3f V", batteryVoltage)),
			),
		})
	if err != nil {
		debugLogger.Println("error when setting: ", err) // notifications are not important enough to return an error.

		return
	}

	err = oneShot.SetReset(dogName+"criticalBattery",
		oshotpkg.Config{
			SetIf: (batteryVoltage < BatteryCriticalThreshold) && nowIsWakingHours,
			OnSet: makeNotifier(
				ctx,
				notifier,
				notify.Title(fmt.Sprintf("%s's battery critical", dogName)),
				notify.Message(fmt.Sprintf("Battery voltage: %.3f V",
					batteryVoltage)),
			),
			ResetIf: batteryVoltage > BatteryLowThreshold,
		})
	if err != nil {
		debugLogger.Println("error when setting: ", err) // notifications are not important enough to return an error.

		return
	}
}
