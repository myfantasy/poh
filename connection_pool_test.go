package poh

import (
	"context"
	"testing"
	"time"
)

func TestConnectionPoolMinCount(t *testing.T) {
	cp := MakeConnectionPool(
		context.Background(),
		func(cause error) {},
		func(ctxBase context.Context) (*Connection[struct{}], error) {
			return MakeConnection(struct{}{},
				func(ctx context.Context, conn struct{}) error { return nil },
				func() time.Duration { return 10 * time.Second },
				nil,
			), nil
		},
		func() int { return 10 },
		func() int { return 5 },
	)

	cp.ClearAndOpenJobStep()

	if len(cp.conns) != 5 {
		t.Errorf("conns should be 5 but `%v`", len(cp.conns))
	}
}

func TestConnectionPoolClear(t *testing.T) {
	cp := MakeConnectionPool(
		context.Background(),
		func(cause error) {},
		func(ctxBase context.Context) (*Connection[struct{}], error) {
			return MakeConnection(struct{}{},
				func(ctx context.Context, conn struct{}) error { return nil },
				nil,
				nil,
			), nil
		},
		nil,
		nil,
	)

	for i := 0; i < 5; i++ {
		conn, free, _ := cp.GenerateConnectionInternal(context.Background())
		conn.Terminate(context.Background())
		free()
	}

	if len(cp.conns) != 5 {
		t.Errorf("conns should be 5 but `%v`", len(cp.conns))
	}

	cp.ClearAndOpenJobStep()

	// wait for background process
	time.Sleep(time.Millisecond)

	if len(cp.conns) != 0 {
		t.Errorf("conns should be 0 after clear but `%v`", len(cp.conns))
	}
}

func TestConnectionPoolGet(t *testing.T) {

}
