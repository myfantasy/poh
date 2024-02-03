package poh

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestConnection(t *testing.T) {
	cnct := MakeConnection(struct{}{},
		func(ctx context.Context, conn struct{}) error { return nil },
		nil,
		nil,
	)

	if !cnct.CanUse() {
		t.Error("connection should be in can use")
	}

	ok, free := cnct.TryLock(context.Background())
	if !ok {
		t.Error("lock should be start")
	}

	ok2, free2 := cnct.TryLock(context.Background())
	if ok2 {
		t.Error("lock should be faild")
	}

	free2()

	ok2, free2 = cnct.TryLock(context.Background())
	if ok2 {
		t.Error("lock should be faild v2 after free for fail")
	}

	free2()

	free()

	ok2, free2 = cnct.TryLock(context.Background())
	if !ok2 {
		t.Error("lock should be start v2 after free for start")
	}

	if cnct.CanUse() {
		t.Error("connection should be in not can use")
	}

	err := cnct.Close(context.Background())

	if !errors.Is(err, ErrInUse) {
		t.Errorf("close err should be ErrInUse but %v", err)
	}

	free2()

	cnct.TermimateConnection = func(ctx context.Context, conn struct{}) error {
		return fmt.Errorf("test error")
	}

	err = cnct.Close(context.Background())

	if !errors.Is(err, ErrConnTerminate) {
		t.Errorf("close err should be ErrConnTerminate but %v", err)
	}

	if cnct.CheckIsTerminated() {
		t.Errorf("connect should be not terminated")
	}

	cnct.TermimateConnection = func(ctx context.Context, conn struct{}) error { return nil }

	err = cnct.Close(context.Background())

	if err != nil {
		t.Errorf("close err should be nil but %v", err)
	}

	if !cnct.CheckIsTerminated() {
		t.Errorf("connect should be terminated")
	}

	err = cnct.Close(context.Background())

	if err != nil {
		t.Errorf("close second time err should be nil but %v", err)
	}

	ok2, _ = cnct.TryLock(context.Background())
	if ok2 {
		t.Error("lock should be fail v3 after close")
	}

	if !cnct.LastUseTime.After(cnct.StartTime) {
		t.Errorf("Use `%v` should be after star time `%v`", cnct.LastUseTime, cnct.StartTime)
	}

	if cnct.UsedQty != 2 {
		t.Errorf("used qty should be 2 but `%v`", cnct.UsedQty)
	}
}

func TestConnectionCloseByTimeout(t *testing.T) {
	cnct := MakeConnection(struct{}{},
		func(ctx context.Context, conn struct{}) error { return nil },
		nil,
		nil,
	)

	ok := cnct.CheckExpired()

	if ok {
		t.Errorf("connect should be not expired")
	}

	time.Sleep(time.Millisecond)

	cnct.OpenExpire = func() time.Duration { return 0 }

	ok = cnct.CheckExpired()

	if ok {
		t.Errorf("connect should be not expired when OpenExpire = 0")
	}

	doClose, err := cnct.CheckAndClose(context.Background())

	if doClose {
		t.Errorf("connect should be not expired and not closed")
	}

	if err != nil {
		t.Errorf("connect should be not expired and not closed without error but `%v`", err)
	}

	cnct.LockDo(
		func() {
			cnct.OpenExpire = func() time.Duration {
				return time.Nanosecond
			}
		},
	)

	ok = cnct.CheckExpired()

	if !ok {
		t.Errorf("connect should be expired")
	}

	doClose, err = cnct.CheckAndClose(context.Background())

	if !doClose {
		t.Errorf("connect should be expired and closed")
	}

	if err != nil {
		t.Errorf("connect should be expired and closed without error but `%v`", err)
	}
}

func TestConnectionCloseByIdleTimeout(t *testing.T) {
	cnct := MakeConnection(struct{}{},
		func(ctx context.Context, conn struct{}) error { return nil },
		nil,
		nil,
	)

	ok := cnct.CheckIdleExpired()

	if ok {
		t.Errorf("connect should be not expired idle")
	}

	time.Sleep(time.Millisecond)

	cnct.IdleExpire = func() time.Duration { return 0 }

	ok = cnct.CheckIdleExpired()

	if ok {
		t.Errorf("connect should be not expired idle when IdleExpire = 0")
	}

	doClose, err := cnct.CheckAndClose(context.Background())

	if doClose {
		t.Errorf("connect should be not expired idle and not closed")
	}

	if err != nil {
		t.Errorf("connect should be not expired idle and not closed without error but `%v`", err)
	}

	cnct.LockDo(
		func() {
			cnct.IdleExpire = func() time.Duration {
				return time.Nanosecond
			}
		},
	)

	ok = cnct.CheckIdleExpired()

	if !ok {
		t.Errorf("connect should be expired idle")
	}

	doClose, err = cnct.CheckAndClose(context.Background())

	if !doClose {
		t.Errorf("connect should be expired idle and closed")
	}

	if err != nil {
		t.Errorf("connect should be expired idle and closed without error but `%v`", err)
	}
}

func TestConnectionCloseRunOpen(t *testing.T) {
	cnct := MakeConnection(struct{}{},
		func(ctx context.Context, conn struct{}) error { return nil },
		func() time.Duration { return time.Microsecond },
		nil,
	)

	cnct.CloseJobRun(context.Background(), time.Millisecond)

	time.Sleep(2 * time.Millisecond)

	if !cnct.CheckIsTerminated() {
		t.Errorf("connect should be closed")
	}
}

func TestConnectionCloseRunIdle(t *testing.T) {
	cnct := MakeConnection(struct{}{},
		func(ctx context.Context, conn struct{}) error { return nil },
		nil,
		func() time.Duration { return time.Microsecond },
	)

	cnct.CloseJobRun(context.Background(), time.Millisecond)

	time.Sleep(2 * time.Millisecond)

	if !cnct.CheckIsTerminated() {
		t.Errorf("connect should be closed")
	}
}

func TestConnectionTerminate(t *testing.T) {
	cnct := MakeConnection(struct{}{},
		func(ctx context.Context, conn struct{}) error { return nil },
		nil,
		nil,
	)

	ok, free := cnct.TryLock(context.Background())
	if !ok {
		t.Error("lock should be start")
	}

	cnct.TermimateConnection = func(ctx context.Context, conn struct{}) error { return fmt.Errorf("test") }

	err := cnct.Terminate(context.Background())

	if !errors.Is(err, ErrConnTerminate) {
		t.Errorf("terminate err should be ErrConnTerminate but `%v`", err)
	}

	cnct.TermimateConnection = func(ctx context.Context, conn struct{}) error { return nil }

	err = cnct.Terminate(context.Background())

	if err != nil {
		t.Errorf("terminate err should be nil but `%v`", err)
	}

	free()

	if !cnct.CheckIsTerminated() {
		t.Errorf("connect should be terminated")
	}
}
