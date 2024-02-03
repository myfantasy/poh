package poh

import (
	"context"
	"sync"
	"time"

	"github.com/myfantasy/mfctx"
)

const DefaultConnectionPoolCheckTimeout = time.Second

type ConnectionGeneratorFunc[T any] func(ctxBase context.Context) *Connection[T]

type CountFunc func() int

type ConnectionPool[T any] struct {
	ConnectionGenerator ConnectionGeneratorFunc[T]

	MaxCount CountFunc
	MinCount CountFunc

	conns map[string]*Connection[T]
	free  map[string]struct{}

	mx sync.Mutex

	ctxBase context.Context

	CheckTimeout time.Duration
}

func MakeConnectionPool[T any](ctxBase context.Context,
	connGenFunc ConnectionGeneratorFunc[T],
	maxCount CountFunc,
	minCount CountFunc,
) *ConnectionPool[T] {
	return &ConnectionPool[T]{
		conns: make(map[string]*Connection[T]),
		free:  map[string]struct{}{},

		ConnectionGenerator: connGenFunc,

		MaxCount: maxCount,
		MinCount: minCount,

		ctxBase: ctxBase,

		CheckTimeout: DefaultConnectionPoolCheckTimeout,
	}
}

// LockDo - do any with lock from this connection pool
func (cp *ConnectionPool[T]) LockDo(f func()) {
	cp.mx.Lock()
	defer cp.mx.Unlock()
	f()
}

func (cp *ConnectionPool[T]) Get(ctxIn context.Context) (conn *Connection[T], free FreeConnectionFunc, err error) {
	cp.mx.Lock()
	defer cp.mx.Unlock()

	c, f, e := cp.GetInternal(ctxIn)

	return c, func() {
		cp.mx.Lock()
		defer cp.mx.Unlock()
		f()
	}, e
}

func (cp *ConnectionPool[T]) GetInternal(ctxIn context.Context) (conn *Connection[T], free FreeConnectionFunc, err error) {
	ctx := mfctx.FromCtx(ctxIn).Start("poh.ConnectionPool.GetInternal")
	defer func() { ctx.Complete(err) }()

	for k := range cp.free {
		l, freeF := cp.conns[k].TryLock(ctxIn)
		if !l {
			freeF()
			go cp.CheckAndClear(k)
			continue
		}

		delete(cp.free, k)
		return cp.conns[k], func() {
			freeF()
			cp.free[k] = struct{}{}
		}, nil
	}

	return cp.GenerateConnectionInternal(ctxIn)
}

func (cp *ConnectionPool[T]) GenerateConnectionInternal(ctxIn context.Context) (conn *Connection[T], free FreeConnectionFunc, err error) {
	if cp.MaxCount != nil && cp.MaxCount() > 0 && cp.MaxCount() <= len(cp.conns) {
		return nil, freeConnectionFuncEmpty, ErrOverflowCP
	}

	connN := cp.ConnectionGenerator(cp.ctxBase)
	connN.CloseJobRun(cp.ctxBase, cp.CheckTimeout)

	cp.conns[connN.ID] = connN
	l, freeF := connN.TryLock(ctxIn)
	if !l {
		freeF()
		cp.free[connN.ID] = struct{}{}

		return nil, freeConnectionFuncEmpty, ErrInternalLockCP
	}
	return connN, func() {
		freeF()
		cp.free[connN.ID] = struct{}{}
	}, nil
}

func (cp *ConnectionPool[T]) CheckAndClear(id string) {
	cp.mx.Lock()
	defer cp.mx.Unlock()

	cp.CheckAndClearInternal(id)
}

func (cp *ConnectionPool[T]) CheckAndClearInternal(id string) {
	c, ok := cp.conns[id]
	if !ok {
		return
	}

	c.CheckAndClose(cp.ctxBase)

	if !c.CheckIsTerminated() {
		return
	}

	delete(cp.free, id)
	delete(cp.conns, id)
}

func (cp *ConnectionPool[T]) ClearAndOpenJobRun() {
	go func() {
		for cp.ctxBase.Err() == nil {
			time.Sleep(cp.CheckTimeout)
			cp.ClearAndOpenJobStep()
		}
	}()
}

func (cp *ConnectionPool[T]) ClearAndOpenJobStep() {
	cp.mx.Lock()
	defer cp.mx.Unlock()

	cp.ClearAndOpenJobStepInternal()
}
func (cp *ConnectionPool[T]) ClearAndOpenJobStepInternal() {
	for k, v := range cp.conns {
		if v.CheckExpired() || v.CheckIsTerminated() {
			go cp.CheckAndClear(k)
		}
	}

	cp.OpenIdleInternal()
}

func (cp *ConnectionPool[T]) OpenIdleInternal() {
	if cp.MinCount == nil {
		return
	}

	for i := len(cp.conns); i < cp.MinCount(); i++ {
		_, free, _ := cp.GenerateConnectionInternal(cp.ctxBase)
		go free()
	}
}
