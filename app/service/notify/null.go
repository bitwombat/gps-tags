package notify

import (
	"context"
)

type NullNotifier struct{}

func NewNullNotifier() Notifier {
	return NullNotifier{}
}

func (ln NullNotifier) Notify(_ context.Context, _ Title, _ Message) error {
	return nil
}
