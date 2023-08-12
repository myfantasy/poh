# poh
Pool of hosts

If you have any group of sources and you wanna switch between them, you could use poh for simplify this.


## Pool
### Using example
``` golang
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
```
### Resource
`Resource` - a wrapper for your object with additional service fields  
* **`Resource.Release()`** - returns resource into pool
* **`Resource.Value()`** - gets your resource
* `Resource.ID()` - gets unique id for this resource
* `Resource.MarkToClose` - marks an object for deletion
* `Resource.MarkToVerify` - marks an object for verification

### PoolConfig
`PoolConfig` - config for the resource pool  
#### Required fields:
* `Constructor` - constructor for resource
* `Destructor` - destructor for resource
* `Verify` - verifier for resource

Other fields ar optional

### Pool
`Pool` is a pool of resources that manages, creates, destroys, and verifies resources.  
Use **`NewPool`** to create a new pool  
* **`Acquire()`** - gets free resource
* **`Close()`** -  close pool and release in background all resources
* `GetStat()` - gets current statistic
* `Config()` - gets current config  
    If you change the configuration object after creating the pool, the behavior of the pool will also change  
* `MarkAllToClose()` - close all resources



## Poh Pool of hosts