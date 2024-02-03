package poh

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/myfantasy/ints"
	"github.com/myfantasy/mfctx"
)

const ConnectionIDLogParam = "conn_id"
const ConnectionLockedLogParam = "locked"
const ConnectionCheckCloseLogParam = "do_cls"

type FreeConnectionFunc func()
type ExpireDurationFunc func() time.Duration

type TermimateConnectionFunc[T any] func(ctx context.Context, conn T) error

func freeConnectionFuncEmpty() {}

type Connection[T any] struct {
	Conn T

	InUse        bool
	IsTerminated bool
	StartTime    time.Time
	LastUseTime  time.Time
	ID           string
	UsedQty      int

	OpenExpire          ExpireDurationFunc
	TermimateConnection TermimateConnectionFunc[T]

	IdleExpire ExpireDurationFunc

	mx sync.Mutex
}

func MakeConnection[T any](conn T,
	terminateConn TermimateConnectionFunc[T],
	openExpire ExpireDurationFunc,
	idleExpire ExpireDurationFunc,
) *Connection[T] {
	return &Connection[T]{
		Conn:                conn,
		StartTime:           time.Now(),
		LastUseTime:         time.Now(),
		ID:                  ints.NextUUID().Text(62),
		OpenExpire:          openExpire,
		TermimateConnection: terminateConn,
		IdleExpire:          idleExpire,
	}
}

// LockDo - do any with lock from this connection
func (c *Connection[T]) LockDo(f func()) {
	c.mx.Lock()
	defer c.mx.Unlock()
	f()
}

// Close - close connection with lock
func (c *Connection[T]) Close(ctxIn context.Context) (err error) {
	ctx := mfctx.FromCtx(ctxIn).Start("poh.Connection.Close")
	ctx.With(ConnectionIDLogParam, c.ID)
	defer func() { ctx.Complete(err) }()
	c.mx.Lock()
	defer c.mx.Unlock()
	return c.CloseInternal(ctx)
}

// CloseInternal - close connection without lock (use Close)
func (c *Connection[T]) CloseInternal(ctxIn context.Context) (err error) {
	ctx := mfctx.FromCtx(ctxIn).Start("poh.Connection.CloseInternal")
	ctx.With(ConnectionIDLogParam, c.ID)
	defer func() { ctx.Complete(err) }()
	if c.InUse {
		return ErrInUse
	}

	if c.IsTerminated {
		return nil
	}

	err = c.TermimateConnection(ctx, c.Conn)

	if err != nil {
		return errors.Join(ErrConnTerminate, err)
	}

	c.IsTerminated = true

	return nil
}

// Terminate - close connection even its InUse with lock
func (c *Connection[T]) Terminate(ctxIn context.Context) (err error) {
	ctx := mfctx.FromCtx(ctxIn).Start("poh.Connection.Terminate")
	ctx.With(ConnectionIDLogParam, c.ID)
	defer func() { ctx.Complete(err) }()
	c.mx.Lock()
	defer c.mx.Unlock()
	return c.TerminateInternal(ctx)
}

// TerminateInternal - close connection even its InUse without lock (use Close)
func (c *Connection[T]) TerminateInternal(ctxIn context.Context) (err error) {
	ctx := mfctx.FromCtx(ctxIn).Start("poh.Connection.TerminateInternal")
	ctx.With(ConnectionIDLogParam, c.ID)
	defer func() { ctx.Complete(err) }()

	if c.IsTerminated {
		return nil
	}

	err = c.TermimateConnection(ctx, c.Conn)

	if err != nil {
		return errors.Join(ErrConnTerminate, err)
	}

	c.IsTerminated = true

	return nil
}

// CheckIsTerminated - check connection in terminated state with lock
func (c *Connection[T]) CheckIsTerminated() bool {
	c.mx.Lock()
	defer c.mx.Unlock()
	return c.CheckIsTerminatedInternal()
}

// CheckIsTerminatedInternal - check connection in terminated state without lock (use CheckIsTerminated)
func (c *Connection[T]) CheckIsTerminatedInternal() bool {
	return c.IsTerminated
}

// CheckInUse - check connection in use with lock
func (c *Connection[T]) CheckInUse() bool {
	c.mx.Lock()
	defer c.mx.Unlock()
	return c.CheckInUseInternal()
}

// CheckInUseInternal - check connection in use without lock (use CheckInUse)
func (c *Connection[T]) CheckInUseInternal() bool {
	return c.InUse
}

// CheckExpired - checks expired open timeout
func (c *Connection[T]) CheckExpired() bool {
	c.mx.Lock()
	defer c.mx.Unlock()
	return c.CheckExpiredInternal()
}

// CheckExpiredInternal - checks expired open timeout  (use CheckExpired)
func (c *Connection[T]) CheckExpiredInternal() bool {
	return c.OpenExpire != nil && c.OpenExpire() > 0 && time.Now().After(c.StartTime.Add(c.OpenExpire()))
}

// CheckIdleExpired - checks expired idle timeout
func (c *Connection[T]) CheckIdleExpired() bool {
	c.mx.Lock()
	defer c.mx.Unlock()
	return c.CheckIdleExpiredInternal()
}

// CheckIdleExpiredInternal - checks expired idle timeout (use CheckIdleExpired)
func (c *Connection[T]) CheckIdleExpiredInternal() bool {
	return c.IdleExpire != nil && c.IdleExpire() > 0 && time.Now().After(c.LastUseTime.Add(c.IdleExpire()))
}

// CanUse - check connection may be used with lock
func (c *Connection[T]) CanUse() bool {
	c.mx.Lock()
	defer c.mx.Unlock()
	return c.CanUseInternal()
}

// CanUseInternal - check connection in use without lock (use CanUse)
func (c *Connection[T]) CanUseInternal() bool {
	expired := c.CheckExpiredInternal()
	return !c.InUse && !c.IsTerminated && !expired
}

// TryLock - try locks for use with lock
// free is never nil
// use l, f := c.TryLock(ctx)
// defer free
func (c *Connection[T]) TryLock(ctxIn context.Context) (locked bool, free FreeConnectionFunc) {
	c.mx.Lock()
	defer c.mx.Unlock()
	locked, freeRes := c.TryLockInternal(ctxIn)

	if locked {
		freeN := func() {
			c.mx.Lock()
			defer c.mx.Unlock()
			freeRes()
		}

		return locked, freeN
	}

	return locked, freeRes
}

// TryLockInternal - try locks for use without lock (use TryLock)
func (c *Connection[T]) TryLockInternal(ctxIn context.Context) (locked bool, free FreeConnectionFunc) {
	ctx := mfctx.FromCtx(ctxIn).Start("poh.Connection.TryLockInternal")
	ctx.With(ConnectionIDLogParam, c.ID)
	defer func() { ctx.Complete(nil) }()
	defer func() { ctx.With(ConnectionLockedLogParam, locked) }()

	if !c.CanUseInternal() {
		return false, freeConnectionFuncEmpty
	}

	c.InUse = true
	c.UsedQty++
	c.LastUseTime = time.Now()

	fn := func() {
		ctx := mfctx.FromCtx(ctxIn).Start("poh.Connection.TryLockInternal.free")
		ctx.With(ConnectionIDLogParam, c.ID)
		defer func() { ctx.Complete(nil) }()
		c.InUse = false
		c.LastUseTime = time.Now()
	}

	return true, fn
}

// CheckAndClose - close connection when Cfg is set; doClose - close was tryed; with lock
func (c *Connection[T]) CheckAndClose(ctxIn context.Context) (doClose bool, err error) {
	c.mx.Lock()
	defer c.mx.Unlock()
	return c.CheckAndCloseInternal(ctxIn)
}

// CheckAndCloseInternal - close connection when Cfg is set; doClose - close was tryed; without lock so use (CheckAndClose)
func (c *Connection[T]) CheckAndCloseInternal(ctxIn context.Context) (doClose bool, err error) {
	ctx := mfctx.FromCtx(ctxIn).Start("poh.Connection.CheckAndCloseInternal")
	ctx.With(ConnectionIDLogParam, c.ID)
	defer func() { ctx.Complete(err) }()
	defer func() { ctx.With(ConnectionCheckCloseLogParam, doClose) }()

	if c.InUse {
		return false, nil
	}
	if !c.CheckExpiredInternal() && !c.CheckIdleExpiredInternal() {
		return false, nil
	}

	err = c.CloseInternal(ctx)

	return true, err
}

// CloseJobRun runs close process
func (c *Connection[T]) CloseJobRun(ctxBase context.Context, checkTimeout time.Duration) {
	go func() {
		for !c.CheckIsTerminated() {
			time.Sleep(checkTimeout)
			c.CheckAndClose(ctxBase)
		}
	}()
}
