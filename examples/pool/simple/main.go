package main

import (
	"context"
	"fmt"
	"log"

	"github.com/myfantasy/poh"
)

type connection struct {
}

func (conn *connection) DoSomething() {
	fmt.Println("do something from connection")
}

func main() {
	cfg := &poh.PoolConfig[*connection]{}
	cfg.Constructor = func(ctx context.Context) (resource *connection, err error) {
		return &connection{}, nil
	}
	cfg.Destructor = func(resource *connection) {}
	cfg.Verify = func(ctx context.Context, resource *connection) (ok bool) {
		return ok
	}

	pool, err := poh.NewPool(cfg)
	if err != nil {
		log.Fatalf("should create pool BUT error is %v", err)
	}
	defer pool.Close()

	res, err := pool.Acquire(context.Background())
	if err != nil {
		log.Fatalf("should create resource BUT error is %v", err)
	}
	defer res.Release()

	res.Value().DoSomething()
}
