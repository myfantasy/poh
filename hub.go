package poh

import (
	"context"
	"sync"

	"github.com/myfantasy/mfctx"
)

type PointsKeysListFunc[K comparable] func(ctx context.Context) (keys []K, err error)
type PointDestroyFunc[K comparable, T any] func(ctx context.Context, key K, point T) (err error)
type PointRefreshFunc[K comparable, T any] func(ctx context.Context, key K, point T) (err error)
type PointGenerateFunc[K comparable, T any] func(ctx context.Context, key K) (point T, err error)

type Hub[K comparable, T any] struct {
	points map[K]T

	ctxBase context.Context

	pointsKeysList PointsKeysListFunc[K]
	pointDestroy   PointDestroyFunc[K, T]
	pointGenerate  PointGenerateFunc[K, T]
	pointRefresh   PointRefreshFunc[K, T]

	mx sync.Mutex
}

func MakeHub[K comparable, T any](
	ctxBase context.Context,
	pointsKeysList PointsKeysListFunc[K],
	pointDestroy PointDestroyFunc[K, T],
	pointGenerate PointGenerateFunc[K, T],
	pointRefresh PointRefreshFunc[K, T],
) *Hub[K, T] {
	return &Hub[K, T]{
		points: make(map[K]T),

		ctxBase: ctxBase,

		pointsKeysList: pointsKeysList,
		pointDestroy:   pointDestroy,
		pointGenerate:  pointGenerate,
		pointRefresh:   pointRefresh,
	}
}

func (hub *Hub[K, T]) Get(key K) (point T, ok bool) {
	hub.mx.Lock()
	defer hub.mx.Unlock()

	return hub.GetInternal(key)
}

func (hub *Hub[K, T]) GetInternal(key K) (point T, ok bool) {
	point, ok = hub.points[key]

	return point, ok
}

func (hub *Hub[K, T]) Refresh() (err error) {
	hub.mx.Lock()
	defer hub.mx.Unlock()

	return hub.refreshLoadInternal()
}

func (hub *Hub[K, T]) refreshLoadInternal() (err error) {
	ctx := mfctx.FromCtx(hub.ctxBase).Start("Hub.refreshInternal")
	defer func() { ctx.Complete(err) }()

	keys, err := hub.pointsKeysList(hub.ctxBase)
	if err != nil {
		return err
	}

	mkey := make(map[K]bool, len(keys))
	for _, key := range keys {
		mkey[key] = true
		go hub.refreshPoint(ctx.Copy(), key)
	}

	for key := range hub.points {
		if !mkey[key] {
			go hub.destroy(ctx.Copy(), key)
		}
	}

	return nil
}

func (hub *Hub[K, T]) refreshPoint(ctx *mfctx.Crumps, key K) (err error) {
	hub.mx.Lock()
	defer hub.mx.Unlock()

	ctx = ctx.With("key", key).Start("hub.refreshPointInternal")
	defer func() { ctx.Complete(err) }()

	point, ok := hub.points[key]
	if !ok {
		point, err := hub.pointGenerate(ctx, key)
		if err != nil {
			return err
		}

		hub.points[key] = point
		return nil
	}

	err = hub.pointRefresh(ctx, key, point)
	if err != nil {
		return err
	}

	return nil
}

func (hub *Hub[K, T]) destroy(ctx *mfctx.Crumps, key K) {
	hub.mx.Lock()
	point, ok := hub.points[key]
	if ok {
		delete(hub.points, key)
	}
	hub.mx.Unlock()

	if ok {
		go hub.destroyJob(ctx.Copy(), key, point)
	}
}

func (hub *Hub[K, T]) destroyJob(ctx *mfctx.Crumps, key K, point T) {
	err := hub.pointDestroy(ctx, key, point)
	for err != nil {
		err = hub.pointDestroy(ctx, key, point)
	}
}
