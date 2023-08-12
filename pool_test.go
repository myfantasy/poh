package poh

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

type TestConnection struct {
}

func TestHappyPathPool(t *testing.T) {
	cfg := &PoolConfig[*TestConnection]{}
	cfg.Constructor = func(ctx context.Context) (resource *TestConnection, err error) {
		return &TestConnection{}, nil
	}
	cfg.Destructor = func(resource *TestConnection) {}
	cfg.Verify = func(ctx context.Context, resource *TestConnection) (ok bool) {
		return ok
	}

	pool, err := NewPool(cfg)
	if err != nil {
		t.Fatalf("should create pool BUT error is %v", err)
	}
	defer pool.Close()

	res, err := pool.Acquire(context.Background())
	if err != nil {
		t.Fatalf("should create resource BUT error is %v", err)
	}
	defer res.Release()
}

func TestInitPoolErrors(t *testing.T) {
	cfg := &PoolConfig[*TestConnection]{}
	_, err := NewPool(cfg)

	if err == nil {
		t.Error("pool should be not created BUT its created")
	}

	if !errors.Is(err, ErrorConstructorShouldBeSet) {
		t.Errorf("should %v BUT value is %v", ErrorConstructorShouldBeSet, err)
	}

	cfg.Constructor = func(ctx context.Context) (resource *TestConnection, err error) {
		return &TestConnection{}, nil
	}
	_, err = NewPool(cfg)
	if !errors.Is(err, ErrorDestructorShouldBeSet) {
		t.Errorf("should %v BUT value is %v", ErrorDestructorShouldBeSet, err)
	}

	cfg.Destructor = func(resource *TestConnection) {}
	_, err = NewPool(cfg)
	if !errors.Is(err, ErrorVerifyShouldBeSet) {
		t.Errorf("should %v BUT value is %v", ErrorVerifyShouldBeSet, err)
	}

	cfg.Verify = func(ctx context.Context, resource *TestConnection) (ok bool) {
		return ok
	}
}

func statCheck(a Stat, b Stat) bool {
	return a.Total == b.Total &&
		a.Active == b.Active &&
		a.Free == b.Free &&
		a.InUse == b.InUse &&
		a.Verify == b.Verify
}

func toStringStat(s Stat) string {
	b, _ := json.MarshalIndent(s, "", " ")
	return string(b)
}

func TestHappyPathPoolAcquireAndReleaseAndVerify(t *testing.T) {
	cfg := &PoolConfig[*TestConnection]{}
	cfg.Constructor = func(ctx context.Context) (resource *TestConnection, err error) {
		return &TestConnection{}, nil
	}
	cfg.Destructor = func(resource *TestConnection) {}
	cfg.Verify = func(ctx context.Context, resource *TestConnection) (ok bool) {
		return true
	}

	verifyCntOK := 0
	verifyCntFail := 0
	cfg.OnVerify = func(r *Resource[*TestConnection], ok bool) {
		if ok {
			verifyCntOK++
		} else {
			verifyCntFail++
		}
	}

	pool, err := NewPool(cfg)
	if err != nil {
		t.Fatalf("should create pool BUT error is %v", err)
	}
	defer pool.Close()

	n := 5

	var ress []*Resource[*TestConnection]
	for i := 0; i < n; i++ {
		res, err := pool.Acquire(context.Background())
		if err != nil {
			t.Fatalf("should create resource BUT error is %v", err)
		}
		ress = append(ress, res)
	}

	stat := pool.GetStat()
	checkStat := Stat{Total: 5, Active: 5, Free: 0, InUse: 5, Verify: 0}

	if !statCheck(*stat, checkStat) {
		t.Fatalf("should stat after Acquire be %v BUT is %v", toStringStat(checkStat), toStringStat(*stat))
	}

	for _, res := range ress {
		res.Release()
	}

	stat = pool.GetStat()
	checkStat = Stat{Total: 5, Active: 5, Free: 5, InUse: 0, Verify: 0}

	if !statCheck(*stat, checkStat) {
		t.Fatalf("should stat after release be %v BUT is %v", toStringStat(checkStat), toStringStat(*stat))
	}

	ress = make([]*Resource[*TestConnection], 0)
	for i := 0; i < n; i++ {
		res, err := pool.Acquire(context.Background())
		if err != nil {
			t.Fatalf("should create resource BUT error is %v", err)
		}
		ress = append(ress, res)
	}

	stat = pool.GetStat()
	checkStat = Stat{Total: 5, Active: 5, Free: 0, InUse: 5, Verify: 0}

	if !statCheck(*stat, checkStat) {
		t.Fatalf("should stat after second Acquire be %v BUT is %v", toStringStat(checkStat), toStringStat(*stat))
	}

	for _, res := range ress {
		res.MarkToVerify("Hands")
		res.Release()
	}

	time.Sleep(1000 * time.Millisecond)

	stat = pool.GetStat()
	checkStat = Stat{Total: 5, Active: 5, Free: 5, InUse: 0, Verify: 0}

	if !statCheck(*stat, checkStat) {
		t.Fatalf("should stat after second release and verify be %v BUT is %v (verify OK|FAIL %v | %v)",
			toStringStat(checkStat), toStringStat(*stat), verifyCntOK, verifyCntFail)
	}

	for _, res := range ress {
		res.MarkToVerify("Hands2")
	}

	time.Sleep(100 * time.Millisecond)

	stat = pool.GetStat()
	checkStat = Stat{Total: 5, Active: 5, Free: 5, InUse: 0, Verify: 0}

	if !statCheck(*stat, checkStat) {
		t.Fatalf("should stat after second verify be %v BUT is %v (verify OK|FAIL %v | %v)",
			toStringStat(checkStat), toStringStat(*stat), verifyCntOK, verifyCntFail)
	}

	pool.MarkAllToClose()

	time.Sleep(100 * time.Millisecond)

	stat = pool.GetStat()
	checkStat = Stat{Total: 0, Active: 0, Free: 0, InUse: 0, Verify: 0}

	if !statCheck(*stat, checkStat) {
		t.Fatalf("should stat after MarkAllToClose be %v BUT is %v", toStringStat(checkStat), toStringStat(*stat))
	}

}

func TestHappyPathPoolAcquireAndReleaseAndLimits(t *testing.T) {
	cfg := &PoolConfig[*TestConnection]{}
	cfg.Constructor = func(ctx context.Context) (resource *TestConnection, err error) {
		return &TestConnection{}, nil
	}
	cfg.Destructor = func(resource *TestConnection) {}
	cfg.Verify = func(ctx context.Context, resource *TestConnection) (ok bool) {
		return true
	}
	cfg.Min = 2
	cfg.Max = 4

	verifyCntOK := 0
	verifyCntFail := 0
	cfg.OnVerify = func(r *Resource[*TestConnection], ok bool) {
		if ok {
			verifyCntOK++
		} else {
			verifyCntFail++
		}
	}

	pool, err := NewPool(cfg)
	if err != nil {
		t.Fatalf("should create pool BUT error is %v", err)
	}
	defer pool.Close()

	time.Sleep(100 * time.Millisecond)

	stat := pool.GetStat()
	checkStat := Stat{Total: 2, Active: 2, Free: 2, InUse: 0, Verify: 0}

	if !statCheck(*stat, checkStat) {
		t.Fatalf("should stat after second Acquire be %v BUT is %v", toStringStat(checkStat), toStringStat(*stat))
	}

	n := 4
	var ress []*Resource[*TestConnection]
	for i := 0; i < n; i++ {
		res, err := pool.Acquire(context.Background())
		if err != nil {
			t.Fatalf("should create resource BUT error is %v", err)
		}
		ress = append(ress, res)
	}

	stat = pool.GetStat()
	checkStat = Stat{Total: 4, Active: 4, Free: 0, InUse: 4, Verify: 0}

	if !statCheck(*stat, checkStat) {
		t.Fatalf("should stat after Acquire be %v BUT is %v", toStringStat(checkStat), toStringStat(*stat))
	}

	_, err = pool.Acquire(context.Background())

	if err == nil {
		t.Fatalf("should be fail BUT the resource has been created")
	}
	if !errors.Is(err, ErrorTheMaximumHasBeenReached) {
		t.Errorf("should %v BUT value is %v", ErrorTheMaximumHasBeenReached, err)
	}

	for _, res := range ress {
		res.Release()
	}

	stat = pool.GetStat()
	checkStat = Stat{Total: 4, Active: 4, Free: 4, InUse: 0, Verify: 0}

	if !statCheck(*stat, checkStat) {
		t.Fatalf("should stat after Release be %v BUT is %v", toStringStat(checkStat), toStringStat(*stat))
	}

}

func TestBadHealth(t *testing.T) {
	cfg := &PoolConfig[*TestConnection]{}
	cfg.Constructor = func(ctx context.Context) (resource *TestConnection, err error) {
		return &TestConnection{}, nil
	}
	cfg.Destructor = func(resource *TestConnection) {}
	cfg.Verify = func(ctx context.Context, resource *TestConnection) (ok bool) {
		return false
	}
	cfg.Min = 2
	cfg.Max = 4
	cfg.VerifyRestartTimeout = time.Second * 1
	cfg.CheckStateTimeout = time.Millisecond * 1500

	verifyCntOK := 0
	verifyCntFail := 0
	cfg.OnVerify = func(r *Resource[*TestConnection], ok bool) {
		if ok {
			verifyCntOK++
		} else {
			verifyCntFail++
		}
	}

	pool, err := NewPool(cfg)
	if err != nil {
		t.Fatalf("should create pool BUT error is %v", err)
	}
	defer pool.Close()

	time.Sleep(100 * time.Millisecond)

	stat := pool.GetStat()
	checkStat := Stat{Total: 2, Active: 2, Free: 2, InUse: 0, Verify: 0}

	if !statCheck(*stat, checkStat) {
		t.Fatalf("should stat after create pool with min be %v BUT is %v", toStringStat(checkStat), toStringStat(*stat))
	}

	time.Sleep(time.Second * 2)

	stat = pool.GetStat()
	// should clase all connections by verify
	checkStat = Stat{Total: 0, Active: 0, Free: 0, InUse: 0, Verify: 0}

	if !statCheck(*stat, checkStat) {
		t.Fatalf("should stat after first check be %v BUT is %v", toStringStat(checkStat), toStringStat(*stat))
	}

	if verifyCntFail != 2 {
		t.Fatalf("should stat after first check verifyCntFail be %v BUT is %v", 2, verifyCntFail)
	}

	time.Sleep(cfg.CheckStateTimeout)

	stat = pool.GetStat()
	// should reopen connect
	checkStat = Stat{Total: 2, Active: 2, Free: 2, InUse: 0, Verify: 0}

	if !statCheck(*stat, checkStat) {
		t.Fatalf("should stat after second check be %v BUT is %v", toStringStat(checkStat), toStringStat(*stat))
	}
}
