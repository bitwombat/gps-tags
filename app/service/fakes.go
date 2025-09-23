package main

import (
	"context"
	"fmt"

	"github.com/bitwombat/gps-tags/model"
	"github.com/bitwombat/gps-tags/notify"
	"github.com/bitwombat/gps-tags/storage"
)

type FakeStorer struct {
	writtenTx         model.TagTx
	fnGetLastStatuses func(context.Context) (storage.Statuses, error)
	fnGetLastNCoords  func(context.Context, int) (storage.Coords, error)
}

func (s *FakeStorer) WriteTx(_ context.Context, tagTx model.TagTx) (string, error) {
	s.writtenTx = tagTx
	return "", nil
}

func (s FakeStorer) GetLastStatuses(ctx context.Context) (storage.Statuses, error) {
	return s.fnGetLastStatuses(ctx)
}

func (s FakeStorer) GetLastNCoords(ctx context.Context, n int) (storage.Coords, error) {
	return s.fnGetLastNCoords(ctx, n)
}

type notification struct {
	title   notify.Title
	message notify.Message
}

type FakeNotifier struct {
	notifications []notification
}

func (n *FakeNotifier) Notify(_ context.Context, title notify.Title, message notify.Message) error {
	fmt.Println("FAKE notification ", title, message)
	n.notifications = append(n.notifications, notification{title: title, message: message})

	return nil
}
