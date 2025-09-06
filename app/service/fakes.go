package main

import (
	"context"

	"github.com/bitwombat/gps-tags/device"
	"github.com/bitwombat/gps-tags/storage"
)

type FakeStorer struct {
	fnWriteTx           func(context.Context, device.TagTx) (string, error)
	fnGetLastPositions  func(context.Context) ([]storage.PositionRecord, error)
	fnGetLastNPositions func(context.Context, int) ([]storage.PathPointRecord, error)
}

func (s FakeStorer) WriteTx(ctx context.Context, tagTx device.TagTx) (string, error) {
	return s.fnWriteTx(ctx, tagTx)
}

func (s FakeStorer) GetLastPositions(ctx context.Context) ([]storage.PositionRecord, error) {
	return s.fnGetLastPositions(ctx)
}

func (s FakeStorer) GetLastNPositions(ctx context.Context, n int) ([]storage.PathPointRecord, error) {
	return s.fnGetLastNPositions(ctx, n)
}
