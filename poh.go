package poh

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

var ErrorThereAreNoActivePools = fmt.Errorf("there are no active pools")
var ErrorCannotFoundPoolForAcquire = fmt.Errorf("cannot found pool for acquire")
var ErrorCannotFoundPoolSkipedForAcquire = fmt.Errorf("cannot found pool was skipped for acquire")

type MarkToVerifyPoolReason string

const VerifyPoolBackgroundReason MarkToVerifyPoolReason = "background"

type VerifyPool[T any] func(ctx context.Context, p *Pool[T]) (ok bool)

type OnMarkToVerifyPoolFunc[T any] func(name string, reason MarkToVerifyPoolReason)
type OnVerifyPoolFunc[T any] func(name string, pool *Pool[T], res bool)
type OnCreatePoolFunc[T any] func(name string, pool *Pool[T], err error)
type OnDestroyPoolFunc[T any] func(name string, pool *Pool[T])
type OnRecalcStartFunc[T any] func()
type OnRecalcFinishFunc[T any] func()
type OnAcquireFromPoolFunc[T any] func(res *Resource[T], name string, err error)
type OnAcquirePoolFromPoolFunc[T any] func(name string, pool *Pool[T], err error)

// PohConfig - config for Poh (pool of hosts) [Not necessarily hosts, you can say services, points, etc.]
type PohHostConfig[T any] struct {
	// PoolConfig pool config
	PoolConfig *PoolConfig[T]

	// Priority for host
	Priority int
}

// PohConfig - config for Poh (pool of hosts)
type PohConfig[T any] struct {
	// Hosts - map of hosts config
	Hosts map[string]*PohHostConfig[T]

	// Verify - one pool from poh
	Verify VerifyPool[T]

	// VerifyTryCount count of verifying try when it's returning error
	VerifyTryCount int

	// VerifyTimeout thow long can verify do default DefaultVerifyTimeout 5s
	VerifyTimeout time.Duration
	// VerifyRestartTimeout timeout between verifies or DefaultVerifyRestartTimeout 30s
	VerifyRestartTimeout time.Duration
	// VerifyTryUntilOkTimeout retry verify or DefaultVerifyTryUntilOkTimeout = time.Microsecond * 100
	VerifyTryUntilOkTimeout time.Duration

	// CheckStateTimeout timeout between verify and reorder DefaultCheckStateTimeout 10s
	CheckStateTimeout time.Duration

	OnMarkToVerifyPool    OnMarkToVerifyPoolFunc[T]
	OnVerifyPool          OnVerifyPoolFunc[T]
	OnCreatePool          OnCreatePoolFunc[T]
	OnDestroyPool         OnDestroyPoolFunc[T]
	OnRecalcStart         OnRecalcStartFunc[T]
	OnRecalcFinish        OnRecalcFinishFunc[T]
	OnAcquireFromPool     OnAcquireFromPoolFunc[T]
	OnAcquirePoolFromPool OnAcquirePoolFromPoolFunc[T]
}

func (pc *PohConfig[T]) onMarkToVerifyPoolCall(name string, reason MarkToVerifyPoolReason) {
	if pc.OnMarkToVerifyPool == nil {
		return
	}

	go pc.OnMarkToVerifyPool(name, reason)
}
func (pc *PohConfig[T]) onVerifyPoolCall(name string, pool *Pool[T], res bool) {
	if pc.OnVerifyPool == nil {
		return
	}

	go pc.OnVerifyPool(name, pool, res)
}
func (pc *PohConfig[T]) onCreatePoolCall(name string, pool *Pool[T], err error) {
	if pc.OnCreatePool == nil {
		return
	}

	go pc.OnCreatePool(name, pool, err)
}
func (pc *PohConfig[T]) onDestroyPoolCall(name string, pool *Pool[T]) {
	if pc.OnDestroyPool == nil {
		return
	}

	go pc.OnDestroyPool(name, pool)
}
func (pc *PohConfig[T]) onRecalcStartCall() {
	if pc.OnRecalcStart == nil {
		return
	}

	go pc.OnRecalcStart()
}
func (pc *PohConfig[T]) onRecalcFinishCall() {
	if pc.OnRecalcFinish == nil {
		return
	}

	go pc.OnRecalcFinish()
}
func (pc *PohConfig[T]) onAcquireFromPoolCall(res *Resource[T], name string, err error) {
	if pc.OnAcquireFromPool == nil {
		return
	}

	go pc.OnAcquireFromPool(res, name, err)
}
func (pc *PohConfig[T]) onAcquirePoolFromPoolCall(name string, pool *Pool[T], err error) {
	if pc.OnAcquirePoolFromPool == nil {
		return
	}

	go pc.OnAcquirePoolFromPool(name, pool, err)
}

type OrderedNames struct {
	// Priority host
	Priority int

	// RRPlace Round-robin place
	RRPlace int

	// names of pools
	Names []string
}

type checkState struct {
	lastCheck time.Time
	isVrified bool
	needCheck bool
}

// Poh - pool of hosts
type Poh[T any] struct {
	pools      map[string]*Pool[T]
	poolsState map[string]*checkState
	config     PohConfig[T]

	mx sync.Mutex

	order []*OrderedNames

	isActive bool
}

// NewPoh - generate new pool of hosts (poh)
func NewPoh[T any](config *PohConfig[T]) (*Poh[T], error) {
	p := &Poh[T]{
		pools:      make(map[string]*Pool[T]),
		poolsState: make(map[string]*checkState),
		config:     *config,
		order:      make([]*OrderedNames, 0),
		isActive:   true,
	}

	p.recalc()
	go p.jobRecalc()
	return p, nil
}

// Close close poh and all pools and release in background all resources
func (p *Poh[T]) Close() {
	p.mx.Lock()
	p.isActive = false
	for _, pool := range p.pools {
		go pool.Close()
	}
	p.mx.Unlock()
}

// DoLock - do func in lock poh, for change settings for example
func (p *Poh[T]) DoLock(f func()) {
	p.mx.Lock()
	f()
	p.mx.Unlock()
}

func (p *Poh[T]) jobRecalc() {
	for p.isActive {
		p.mx.Lock()
		for name, pool := range p.poolsState {
			if p.isActive {
				if p.config.VerifyRestartTimeout > 0 && time.Now().After(pool.lastCheck.Add(p.config.VerifyRestartTimeout)) {
					go p.MarkToVerifyPool(name, VerifyPoolBackgroundReason)
					continue
				}
				if p.config.VerifyRestartTimeout == 0 && time.Now().After(pool.lastCheck.Add(DefaultVerifyRestartTimeout)) {
					go p.MarkToVerifyPool(name, VerifyPoolBackgroundReason)
					continue
				}
			}
		}
		p.mx.Unlock()

		p.recalc()
		if p.config.CheckStateTimeout > 0 {
			time.Sleep(p.config.CheckStateTimeout)
		} else {
			time.Sleep(DefaultCheckStateTimeout)
		}
	}
}

func (p *Poh[T]) recalc() {
	p.mx.Lock()

	p.config.onRecalcStartCall()

	for n, pool := range p.pools {
		pool := pool
		_, ok := p.config.Hosts[n]

		if !ok {
			delete(p.pools, n)
			delete(p.poolsState, n)
			p.config.onDestroyPoolCall(n, pool)
			go pool.Close()
		}
	}

	for n, cfg := range p.config.Hosts {
		pool, ok := p.pools[n]

		if ok {
			pool.config = cfg.PoolConfig
		} else {
			pn, err := NewPool[T](cfg.PoolConfig)
			if err != nil {
				p.config.onCreatePoolCall(n, pn, err)
			} else {
				p.config.onCreatePoolCall(n, pn, nil)
				p.pools[n] = pn
			}
		}
	}

	type nameOrder struct {
		name  string
		order int
	}

	orderedNames := make([]nameOrder, 0, len(p.pools))

	for n := range p.pools {
		state, ok := p.poolsState[n]
		if !ok {
			state = &checkState{
				lastCheck: time.Now(),
				isVrified: true,
			}
			p.poolsState[n] = state
		}

		cfg, ok := p.config.Hosts[n]

		if state.isVrified && ok {
			orderedNames = append(orderedNames, nameOrder{name: n, order: cfg.Priority})
		}
	}

	sort.Slice(orderedNames, func(i, j int) bool {
		return orderedNames[i].order < orderedNames[j].order
	})

	order := []*OrderedNames{}

	if len(orderedNames) > 0 {
		lastOrder := orderedNames[0].order - 1
		var currentON *OrderedNames
		for _, onms := range orderedNames {
			if lastOrder != onms.order {
				currentON = &OrderedNames{
					Priority: onms.order,
					Names:    make([]string, 0, 1),
				}
				order = append(order, currentON)
			}
			currentON.Names = append(currentON.Names, onms.name)
		}
	}

	p.order = order

	p.config.onRecalcFinishCall()

	p.mx.Unlock()
}

// Acquire - gets one free resource from one of pool and pool name from poh
// How to use:
// res, name, err := p.Acquire(ctx)
// _ = name
// if err != nil { panic(err) }
// defer res.Release()
// resource := res.Value()
func (p *Poh[T]) Acquire(ctx context.Context) (res *Resource[T], name string, err error) {
	defer func() {
		p.config.onAcquireFromPoolCall(res, name, err)
	}()
	p.mx.Lock()
	order := p.order
	p.mx.Unlock()

	var outError error

	if len(order) == 0 {
		return nil, "", ErrorThereAreNoActivePools
	}

	for _, o := range order {
		for i := 0; i < len(o.Names); i++ {
			p.mx.Lock()
			o.RRPlace++
			if o.RRPlace >= len(o.Names) {
				o.RRPlace = 0
			}
			place := o.RRPlace

			poolName := o.Names[place]

			pool, ok := p.pools[poolName]
			p.mx.Unlock()

			if !ok {
				outError = ErrorCannotFoundPoolForAcquire
				continue
			}

			res, err := pool.Acquire(ctx)
			if err != nil {
				outError = err
				continue
			}

			return res, poolName, nil
		}
	}

	return nil, "", outError
}

// AcquirePool - gets one most actual pool and pool name from poh
// exclude - excluded pools names
func (p *Poh[T]) AcquirePool(ctx context.Context, exclude ...string) (res *Pool[T], name string, err error) {
	defer func() {
		p.config.onAcquirePoolFromPoolCall(name, res, err)
	}()
	exSet := ToSet(exclude)
	p.mx.Lock()
	order := p.order
	p.mx.Unlock()

	var outError error

	if len(order) == 0 {
		return nil, "", ErrorThereAreNoActivePools
	}

	for _, o := range order {
		for i := 0; i < len(o.Names); i++ {
			p.mx.Lock()
			o.RRPlace++
			if o.RRPlace >= len(o.Names) {
				o.RRPlace = 0
			}
			place := o.RRPlace

			poolName := o.Names[place]

			if _, ok := exSet[poolName]; ok {
				if outError == nil {
					outError = ErrorCannotFoundPoolSkipedForAcquire
				}
				continue
			}

			pool, ok := p.pools[poolName]
			p.mx.Unlock()

			if !ok {
				outError = ErrorCannotFoundPoolForAcquire
				continue
			}

			return pool, poolName, nil
		}
	}

	return nil, "", outError
}

// GetPool gets pool by name. Use Acquire for get resource
func (p *Poh[T]) GetPool(poolName string) (pool *Pool[T], ok bool) {
	p.mx.Lock()
	defer p.mx.Unlock()

	pool, ok = p.pools[poolName]

	return pool, ok
}

// MarkToVerifyPool - marks the pool for verification
// You can use any string in reason when call it manual
func (p *Poh[T]) MarkToVerifyPool(name string, reason MarkToVerifyPoolReason) {
	p.mx.Lock()
	defer p.mx.Unlock()

	state, ok := p.poolsState[name]

	if !ok {
		return
	}

	p.config.onMarkToVerifyPoolCall(name, reason)

	state.needCheck = true
	go p.verifyPool(name)
}

func (p *Poh[T]) verifyPool(name string) {
	p.mx.Lock()
	state, okState := p.poolsState[name]
	pool, okPool := p.pools[name]
	p.mx.Unlock()

	if !okState || !okPool {
		return
	}

	if !state.needCheck {
		return
	}

	ctx := context.Background()
	var cancel context.CancelFunc
	if p.config.VerifyTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, p.config.VerifyTimeout)
		defer cancel()
	} else {
		ctx, cancel = context.WithTimeout(ctx, DefaultVerifyTimeout)
		defer cancel()
	}

	res := p.config.Verify(ctx, pool)

	p.mx.Lock()
	state.needCheck = false
	state.isVrified = res
	state.lastCheck = time.Now()

	p.config.onVerifyPoolCall(name, pool, res)

	p.mx.Unlock()

	if !res {
		go p.recalc()
	}
}
