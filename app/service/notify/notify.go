package notify

import "context"

type Notifier interface {
	Notify(context.Context, string, string) error
}
