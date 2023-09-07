package poh

import (
	"context"
	"testing"
)

func TestHappyPathPoh(t *testing.T) {
	cfgPool := &PoolConfig[*TestConnection]{}
	cfgPool.Constructor = func(ctx context.Context) (resource *TestConnection, err error) {
		return &TestConnection{}, nil
	}
	cfgPool.Destructor = func(resource *TestConnection) {}
	cfgPool.Verify = func(ctx context.Context, resource *TestConnection) (ok bool) {
		return true
	}
	cfgPoh := &PohConfig[*TestConnection]{
		Hosts: map[string]*PohHostConfig[*TestConnection]{
			"test2": {PoolConfig: cfgPool, Priority: 1},
			"test":  {PoolConfig: cfgPool, Priority: 0},
		},
		Verify: func(ctx context.Context, p *Pool[*TestConnection]) (ok bool) {
			return true
		},
	}

	p, err := NewPoh(cfgPoh)
	if err != nil {
		t.Fatalf("pool should be created BUT error is %v", err)
	}

	_, name, err := p.Acquire(context.Background())
	if err != nil {
		t.Fatalf("pool Acquire should be complete BUT error is %v", err)
	}

	if name != "test" {
		t.Fatalf("pool Acquire should be complete from 'test' BUT name is %v", name)
	}
}
