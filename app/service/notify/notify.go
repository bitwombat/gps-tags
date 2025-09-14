package notify

import "context"

type (
	Title   string
	Message string
)

type Notifier interface {
	Notify(context.Context, Title, Message) error
}
