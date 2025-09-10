package main

import (
	"context"
	"fmt"

	"github.com/bitwombat/gps-tags/device"
	"github.com/bitwombat/gps-tags/storage"
)

type FakeStorer struct {
	writtenTx           device.TagTx
	fnGetLastPositions  func(context.Context) ([]storage.PositionRecord, error)
	fnGetLastNPositions func(context.Context, int) ([]storage.PathPointRecord, error)
}

func (s *FakeStorer) WriteTx(ctx context.Context, tagTx device.TagTx) (string, error) {
	s.writtenTx = tagTx
	return "", nil
}

func (s FakeStorer) GetLastPositions(ctx context.Context) ([]storage.PositionRecord, error) {
	return s.fnGetLastPositions(ctx)
}

func (s FakeStorer) GetLastNPositions(ctx context.Context, n int) ([]storage.PathPointRecord, error) {
	return s.fnGetLastNPositions(ctx, n)
}

type notification struct {
	title   string
	message string
}

type FakeNotifier struct {
	notifications []notification
}

// TODO: title and message should be types. "string, string" in interface not
// very explicative.
func (n *FakeNotifier) Notify(_ context.Context, title string, message string) error {
	fmt.Println("FAKE notification ", title, message)
	n.notifications = append(n.notifications, notification{title: title, message: message})

	return nil
}
