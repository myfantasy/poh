package poh

import (
	"context"
	"testing"
	"time"

	"github.com/myfantasy/mfctx/jsonify"
)

func TestHub(t *testing.T) {
	hub := MakeHub[string, struct{}](
		context.Background(),
		func(ctx context.Context) (keys []string, err error) {
			return []string{"a", "b", "c"}, nil
		},
		func(ctx context.Context, key string, point struct{}) (err error) {
			return nil
		},
		func(ctx context.Context, key string) (point struct{}, err error) {
			return struct{}{}, nil
		},
		func(ctx context.Context, key string, point struct{}) (err error) {
			return nil
		},
	)

	err := hub.Refresh()
	if err != nil {
		t.Error(err)
	}

	time.Sleep(20 * time.Millisecond)

	_, ok := hub.Get("a")
	if !ok {
		t.Error(jsonify.JsonifySLn(hub))
	}

	_, ok = hub.Get("c")
	if !ok {
		t.Error(jsonify.JsonifySLn(hub))
	}

	_, ok = hub.Get("d")
	if ok {
		t.Error(jsonify.JsonifySLn(hub))
	}

	hub.pointsKeysList = func(ctx context.Context) (keys []string, err error) {
		return []string{"k", "d", "c"}, nil
	}

	err = hub.Refresh()
	if err != nil {
		t.Error(err)
	}

	time.Sleep(20 * time.Millisecond)

	_, ok = hub.Get("a")
	if ok {
		t.Error(jsonify.JsonifySLn(hub))
	}

	_, ok = hub.Get("d")
	if !ok {
		t.Error(jsonify.JsonifySLn(hub))
	}

	_, ok = hub.Get("c")
	if !ok {
		t.Error(jsonify.JsonifySLn(hub))
	}
}
