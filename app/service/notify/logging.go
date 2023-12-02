package notify

import (
	"context"
	"log"
)

type LoggingNotifier struct {
	Notifier Notifier
	Logger   *log.Logger
}

func NewLoggingNotifier(n Notifier, l *log.Logger) Notifier {
	return LoggingNotifier{
		Notifier: n,
		Logger:   l,
	}
}

func (ln LoggingNotifier) Notify(ctx context.Context, title, message string) error {
	ln.Logger.Print("Sending notification: \"" + title + "\" \"" + message + "\"")
	err := ln.Notifier.Notify(ctx, title, message)
	return err
}
