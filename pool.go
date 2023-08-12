package poh

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/myfantasy/ints"
)

var ErrorConstructorShouldBeSet = fmt.Errorf("constructor should be set")
var ErrorDestructorShouldBeSet = fmt.Errorf("destructor should be set")
var ErrorVerifyShouldBeSet = fmt.Errorf("verify should be set")

var ErrorTheMaximumHasBeenReached = fmt.Errorf("the maximum number has been reached")
var ErrorPoolIsClosed = fmt.Errorf("the pool is closed")

const DefaultVerifyTimeout = time.Second * 5
const DefaultVerifyRestartTimeout = time.Second * 30
const DefaultVerifyTryUntilOkTimeout = time.Microsecond * 100
const DefaultCheckStateTimeout = time.Second * 10

const DefaultCreateResourceTimeout = time.Second * 5

type MarkToCloseReason string

const VerifyFailReason MarkToCloseReason = "verify_fail"
const IdleTimeoutReason MarkToCloseReason = "idle_timeout"
const WorkTimeoutReason MarkToCloseReason = "work_timeout"
const MarkAllToCloseReason MarkToCloseReason = "mark_all_to_close"

type CreateResourceReason string

const AcquireReason CreateResourceReason = "acquire"
const CreateIdleReason CreateResourceReason = "create_idle"

type MarkToVerifyReason string

const VerifyBackgroundReason MarkToVerifyReason = "background"

type Constructor[T any] func(ctx context.Context) (resource T, err error)
type Destructor[T any] func(resource T)
type Clean[T any] func(resource T)
type Verify[T any] func(ctx context.Context, resource T) (ok bool)

type OnMarkToClose[T any] func(r *Resource[T], reason MarkToCloseReason)
type OnMarkToVerify[T any] func(r *Resource[T], reason MarkToVerifyReason)
type OnDestroy[T any] func(id ints.Uuid)
type OnAcquire[T any] func(r *Resource[T], err error)
type OnRelease[T any] func(r *Resource[T])
type OnVerify[T any] func(r *Resource[T], ok bool)
type OnCreate[T any] func(r *Resource[T], err error, reason CreateResourceReason)

// Resource - a wrapper for your object with additional service fields
type Resource[T any] struct {
	resource T
	id       ints.Uuid

	needToClose  bool
	needToVerify bool
	inUse        bool

	pool *Pool[T]

	createTime time.Time
	lastUse    time.Time
	lastVerify time.Time
}

// ID - gets unique id for this resource
func (r *Resource[T]) ID() ints.Uuid {
	return r.id
}

// Value - gets resource
func (r *Resource[T]) Value() T {
	return r.resource
}

// MarkToClose - marks an object for deletion
// You can use any string in reason when call it manual
func (r *Resource[T]) MarkToClose(reason MarkToCloseReason) {
	r.pool.mx.Lock()
	defer r.pool.mx.Unlock()

	r.needToClose = true

	r.pool.config.callOnMarkToClose(r, reason)

	r.putByMarks()
}

// MarkToVerify - marks an object for verification
// You can use any string in reason when call it manual
func (r *Resource[T]) MarkToVerify(reason MarkToVerifyReason) {
	r.pool.mx.Lock()
	defer r.pool.mx.Unlock()

	r.needToVerify = true

	r.putByMarks()

	r.pool.config.callOnMarkToVerify(r, reason)
}

func (r *Resource[T]) putByMarks() {
	if _, ok := r.pool.freeResources[r.id]; ok {
		if r.inUse {
			delete(r.pool.freeResources, r.id)
			r.pool.useResources[r.id] = struct{}{}
			return
		}
		if r.needToClose {
			delete(r.pool.freeResources, r.id)
			r.pool.closeResources[r.id] = struct{}{}
			go r.deleteResource()
			return
		}
		if r.needToVerify {
			delete(r.pool.freeResources, r.id)
			r.pool.verifyResources[r.id] = struct{}{}
			go r.verifyUntilOk()
			return
		}
	}

	if _, ok := r.pool.verifyResources[r.id]; ok {
		if r.inUse {
			delete(r.pool.verifyResources, r.id)
			r.pool.useResources[r.id] = struct{}{}
			return
		}
		if r.needToClose {
			delete(r.pool.verifyResources, r.id)
			r.pool.closeResources[r.id] = struct{}{}
			go r.deleteResource()
			return
		}
		if !r.needToVerify {
			delete(r.pool.verifyResources, r.id)
			r.pool.freeResources[r.id] = struct{}{}
			return
		}
	}

	if _, ok := r.pool.useResources[r.id]; ok {
		if r.inUse {
			return
		}
		if r.needToClose {
			delete(r.pool.useResources, r.id)
			r.pool.closeResources[r.id] = struct{}{}
			go r.deleteResource()
			return
		}
		if r.needToVerify {
			delete(r.pool.useResources, r.id)
			r.pool.verifyResources[r.id] = struct{}{}
			go r.verifyUntilOk()
			return
		}
		delete(r.pool.useResources, r.id)
		r.pool.freeResources[r.id] = struct{}{}
		return
	}
}

func (r *Resource[T]) deleteResource() {
	r.pool.mx.Lock()

	delete(r.pool.resources, r.id)
	delete(r.pool.closeResources, r.id)

	r.pool.mx.Unlock()

	go r.pool.config.Destructor(r.resource)
	r.pool.config.callOnDestroy(r)
}

func (r *Resource[T]) verifyUntilOk() {
	for !r.verify() {
		if r.pool.config.VerifyTryUntilOkTimeout > 0 {
			time.Sleep(r.pool.config.VerifyTryUntilOkTimeout)
		} else {
			time.Sleep(DefaultVerifyTryUntilOkTimeout)
		}
	}
}

func (r *Resource[T]) verify() bool {
	r.pool.mx.Lock()
	if !r.needToVerify {
		r.pool.mx.Unlock()
		return true
	}
	_, needVerify := r.pool.verifyResources[r.id]
	r.pool.mx.Unlock()

	if !needVerify {
		return false
	}

	ctx := context.Background()
	var cancel context.CancelFunc
	if r.pool.config.VerifyTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, r.pool.config.VerifyTimeout)
		defer cancel()
	} else {
		ctx, cancel = context.WithTimeout(ctx, DefaultVerifyTimeout)
		defer cancel()
	}

	result := r.pool.config.Verify(ctx, r.resource)
	r.pool.config.callOnVerify(r, result)
	r.pool.mx.Lock()
	if r.needToClose {

	} else if result {
		r.needToVerify = false
		r.lastVerify = time.Now()
		r.putByMarks()
	}
	r.pool.mx.Unlock()

	if !result {
		r.MarkToClose(VerifyFailReason)
	}
	return true
}

// Release - returns resource into pool
func (r *Resource[T]) Release() {
	if r.pool.config.Clean != nil {
		r.pool.config.Clean(r.resource)
	}

	r.pool.mx.Lock()
	defer r.pool.mx.Unlock()

	r.inUse = false
	r.putByMarks()

	r.pool.config.callOnRelease(r)
}

// PoolConfig - config for the resource pool
type PoolConfig[T any] struct {
	// Constructor - constructor for resource
	Constructor Constructor[T]
	// Destructor - destructor for resource
	Destructor Destructor[T]

	// Clean when set calls when release
	Clean[T]

	// Verify - verifier for resource
	Verify Verify[T]

	// VerifyTimeout thow long can verify do default DefaultVerifyTimeout 5s
	VerifyTimeout time.Duration
	// VerifyRestartTimeout timeout between verifies or DefaultVerifyRestartTimeout 30s
	VerifyRestartTimeout time.Duration
	// VerifyTryUntilOkTimeout retry verify or DefaultVerifyTryUntilOkTimeout = time.Microsecond * 100
	VerifyTryUntilOkTimeout time.Duration

	// Max case when Max <=0 then unlimit
	Max int
	// Min case when Min >0 then sets the number of resources to be created
	Min int

	// Timeout for create resource DefaultCreateResourceTimeout is 5s
	CreateResourceTimeout time.Duration

	// CheckStateTimeout timeout between check minIdle and close timeout and release closed or DefaultCheckStateTimeout 10s
	CheckStateTimeout time.Duration

	// IdleTimeout - timeout to close after last use then > 0
	IdleTimeout time.Duration

	// WorkTimeout - timeout to close after create then > 0
	WorkTimeout time.Duration

	// OnMarkToClose event when resource marks to close
	OnMarkToClose OnMarkToClose[T]
	// OnMarkToVerify event when resource marks to close
	OnMarkToVerify OnMarkToVerify[T]

	OnDestroy OnDestroy[T]
	OnAcquire OnAcquire[T]
	OnRelease OnRelease[T]
	OnVerify  OnVerify[T]
	OnCreate  OnCreate[T]
}

func (c *PoolConfig[T]) callOnMarkToClose(r *Resource[T], reason MarkToCloseReason) {
	if c.OnMarkToClose == nil {
		return
	}

	go c.OnMarkToClose(r, reason)
}
func (c *PoolConfig[T]) callOnMarkToVerify(r *Resource[T], reason MarkToVerifyReason) {
	if c.OnMarkToClose == nil {
		return
	}

	go c.OnMarkToVerify(r, reason)
}
func (c *PoolConfig[T]) callOnDestroy(r *Resource[T]) {
	if c.OnDestroy == nil {
		return
	}

	go c.OnDestroy(r.id)
}
func (c *PoolConfig[T]) callOnAcquire(r *Resource[T], err error) {
	if c.OnAcquire == nil {
		return
	}

	go c.OnAcquire(r, err)
}
func (c *PoolConfig[T]) callOnRelease(r *Resource[T]) {
	if c.OnRelease == nil {
		return
	}

	go c.OnRelease(r)
}
func (c *PoolConfig[T]) callOnVerify(r *Resource[T], ok bool) {
	if c.OnVerify == nil {
		return
	}

	go c.OnVerify(r, ok)
}
func (c *PoolConfig[T]) callOnCreate(r *Resource[T], err error, reason CreateResourceReason) {
	if c.OnCreate == nil {
		return
	}

	go c.OnCreate(r, err, reason)
}

// Pool - resource pool
type Pool[T any] struct {
	resources map[ints.Uuid]*Resource[T]
	mx        sync.Mutex
	mxCreate  sync.Mutex

	isActive bool

	config *PoolConfig[T]

	freeResources   map[ints.Uuid]struct{}
	closeResources  map[ints.Uuid]struct{}
	verifyResources map[ints.Uuid]struct{}
	useResources    map[ints.Uuid]struct{}
}

// NewPool creates a new pool
func NewPool[T any](config *PoolConfig[T]) (*Pool[T], error) {
	if config.Constructor == nil {
		return nil, ErrorConstructorShouldBeSet
	}
	if config.Destructor == nil {
		return nil, ErrorDestructorShouldBeSet
	}
	if config.Verify == nil {
		return nil, ErrorVerifyShouldBeSet
	}

	p := &Pool[T]{
		config:          config,
		isActive:        true,
		resources:       make(map[ints.Uuid]*Resource[T]),
		freeResources:   make(map[ints.Uuid]struct{}),
		closeResources:  make(map[ints.Uuid]struct{}),
		verifyResources: make(map[ints.Uuid]struct{}),
		useResources:    make(map[ints.Uuid]struct{}),
	}

	go p.checkJob()
	return p, nil
}

// Config gets current config
func (p *Pool[T]) Config() *PoolConfig[T] {
	return p.config
}

func (p *Pool[T]) checkJob() {
	for p.isActive || len(p.resources) > 0 {
		p.mx.Lock()
		for _, resource := range p.resources {
			if resource.needToClose {
				continue
			}
			if p.config.WorkTimeout > 0 && time.Now().After(resource.createTime.Add(p.config.WorkTimeout)) {
				go resource.MarkToClose(WorkTimeoutReason)
				continue
			}
			if p.config.IdleTimeout > 0 && time.Now().After(resource.lastUse.Add(p.config.IdleTimeout)) {
				go resource.MarkToClose(IdleTimeoutReason)
				continue
			}

			if p.isActive {
				if p.config.VerifyRestartTimeout > 0 && time.Now().After(resource.lastVerify.Add(p.config.VerifyRestartTimeout)) {
					go resource.MarkToVerify(VerifyBackgroundReason)
					continue
				}
				if p.config.VerifyRestartTimeout == 0 && time.Now().After(resource.lastVerify.Add(DefaultVerifyRestartTimeout)) {
					go resource.MarkToVerify(VerifyBackgroundReason)
					continue
				}
			}
		}
		p.mx.Unlock()

		p.upToMinIdle()

		if p.config.CheckStateTimeout > 0 {
			time.Sleep(p.config.CheckStateTimeout)
		} else {
			time.Sleep(DefaultCheckStateTimeout)
		}
	}
}

func (p *Pool[T]) upToMinIdle() {
	needToAdd := false
	p.mx.Lock()
	needToAdd = p.activeLen() < p.config.Min && (p.config.Max <= 0 || p.activeLen() < p.config.Max)
	p.mx.Unlock()

	for needToAdd {
		ctx := context.Background()
		var cancel context.CancelFunc
		if p.config.CreateResourceTimeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, p.config.CreateResourceTimeout)
		} else {
			ctx, cancel = context.WithTimeout(ctx, DefaultCreateResourceTimeout)
		}

		r, err := p.createNewResource(ctx, false, CreateIdleReason)
		if err != nil {
			// DO NOTHING
		} else {
			r.Release()
		}

		cancel()

		p.mx.Lock()
		needToAdd = p.activeLen() < p.config.Min && (p.config.Max <= 0 || p.activeLen() < p.config.Max)
		p.mx.Unlock()
	}
}

// Close close pool and release in background all resources
func (p *Pool[T]) Close() {
	p.mx.Lock()
	defer p.mx.Unlock()
	p.isActive = false
	go p.MarkAllToClose()
}

func (p *Pool[T]) activeLen() int {
	return len(p.freeResources) + len(p.useResources) + len(p.verifyResources)
}

func (p *Pool[T]) Acquire(ctx context.Context) (res *Resource[T], err error) {
	defer p.config.callOnAcquire(res, err)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	p.mx.Lock()
	var key ints.Uuid
	found := false
	for k := range p.freeResources {
		key = k
		found = true
		break
	}
	if found {
		resource := p.resources[key]
		resource.inUse = true
		resource.lastUse = time.Now()
		resource.putByMarks()

		p.mx.Unlock()
		return resource, nil
	}

	p.mx.Unlock()

	return p.createNewResource(ctx, true, AcquireReason)
}

func (p *Pool[T]) createNewResource(ctx context.Context, useFreeCons bool, reason CreateResourceReason) (*Resource[T], error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	p.mxCreate.Lock()
	p.mx.Lock()
	if !p.isActive {
		p.mx.Unlock()
		p.mxCreate.Unlock()
		return nil, ErrorPoolIsClosed
	}
	if useFreeCons && len(p.freeResources) > 0 {
		var key ints.Uuid
		for k := range p.freeResources {
			key = k
			break
		}
		resource := p.resources[key]
		resource.inUse = true
		resource.putByMarks()

		p.mx.Unlock()
		p.mxCreate.Unlock()
		return resource, nil
	}
	if p.config.Max > 0 && p.activeLen() >= p.config.Max {
		p.mx.Unlock()
		p.mxCreate.Unlock()
		return nil, ErrorTheMaximumHasBeenReached
	}
	p.mx.Unlock()

	res, err := p.config.Constructor(ctx)
	if err != nil {
		p.config.callOnCreate(nil, err, reason)
		p.mxCreate.Unlock()
		return nil, err
	}

	resource := &Resource[T]{
		resource:   res,
		id:         ints.NextUUID(),
		pool:       p,
		lastUse:    time.Now(),
		createTime: time.Now(),
		lastVerify: time.Now(),
	}

	p.mx.Lock()
	p.resources[resource.id] = resource
	p.freeResources[resource.id] = struct{}{}

	resource.inUse = true
	resource.putByMarks()

	p.mx.Unlock()
	p.config.callOnCreate(resource, nil, reason)
	p.mxCreate.Unlock()
	return resource, nil
}

// MarkAllToClose - close all resources
func (p *Pool[T]) MarkAllToClose() {
	p.mx.Lock()
	for _, v := range p.resources {
		go v.MarkToClose(MarkAllToCloseReason)
	}
	p.mx.Unlock()
}

// Stat - statistics for pool
type Stat struct {
	// Total resources includes resourses waiting for destroy
	Total int
	// Total resources Idle(free), inUse and waiting for Validate
	Active int

	// Free resources (idle)
	Free int

	// InUse resources in use
	InUse int

	// Verify resources waiting for verify
	Verify int
}

// GetStat - gets current statistic
func (p *Pool[T]) GetStat() (stat *Stat) {
	stat = &Stat{}
	p.mx.Lock()
	defer p.mx.Unlock()

	stat.Total = len(p.resources)
	stat.Active = p.activeLen()

	stat.Free = len(p.freeResources)
	stat.InUse = len(p.useResources)
	stat.Verify = len(p.verifyResources)

	return stat
}
