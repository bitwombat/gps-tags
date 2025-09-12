package main

import (
	"context"
	"fmt"

	"github.com/bitwombat/gps-tags/model"
	"github.com/bitwombat/gps-tags/storage"
)

type FakeStorer struct {
	writtenTx         model.TagTx
	fnGetLastStatuses func(context.Context) ([]storage.Status, error)
	fnGetLastNCoords  func(context.Context, int) (storage.Coords, error)
}

func (s *FakeStorer) WriteTx(_ context.Context, tagTx model.TagTx) (string, error) {
	s.writtenTx = tagTx
	return "", nil
}

func (s FakeStorer) GetLastStatuses(ctx context.Context) ([]storage.Status, error) {
	return s.fnGetLastStatuses(ctx)
}

func (s FakeStorer) GetLastNCoords(ctx context.Context, n int) (storage.Coords, error) {
	return s.fnGetLastNCoords(ctx, n)
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
func (n *FakeNotifier) Notify(_ context.Context, title, message string) error {
	fmt.Println("FAKE notification ", title, message)
	n.notifications = append(n.notifications, notification{title: title, message: message})

	return nil
}
