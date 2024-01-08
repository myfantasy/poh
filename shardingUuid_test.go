package poh

import (
	"testing"
	"time"

	"github.com/myfantasy/ints"
)

func TestShardingConfigUuidFindShard(t *testing.T) {
	sc := SharedConfigForTest()
	//sc := SharedConfigFromStringForTest()

	s := sc.FindShard(
		ints.LimitSerialUUID(
			time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC),
		),
		time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC),
	)

	if s == nil || s.ShardRangeConfig.Name != "s0" ||
		s.ShardRangeConfig.BucketQuantity != 3 || s.ShardRangeConfig.BucketNames[2] != "c" ||
		s.IsSpecial == true {
		t.Errorf(ToJson(s))
	}

	s = sc.FindShard(
		ints.UuidFromTextMust("0188bc5c-1400-4000-8000-000000000000", 16, true),
		time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC),
	)

	if s == nil || s.ShardRangeConfig.Name != "" ||
		s.ShardRangeConfig.BucketQuantity != 1 || s.ShardRangeConfig.BucketNames[0] != "s2" ||
		s.IsSpecial == false {
		t.Errorf(ToJson(s))
	}

	s = sc.FindShard(
		ints.UuidFromTextMust("0199bc5b-1400-4000-8000-000000000000", 16, true),
		time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC),
	)

	if s == nil || s.ShardRangeConfig.Name != "" ||
		s.ShardRangeConfig.BucketQuantity != 1 || s.ShardRangeConfig.BucketNames[0] != "s1" ||
		s.IsSpecial == false {
		t.Errorf(ToJson(s))
	}

	s = sc.FindShard(
		ints.UuidFromTextMust("0199bc5a-1400-4000-8000-000000000000", 16, true),
		time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC),
	)

	if s != nil {
		t.Errorf(ToJson(s))
	}

}

func TestShardingConfigUuidFindShards(t *testing.T) {
	sc := SharedConfigForTest()

	fss := sc.FindShards(ints.UuidFromTextMust("0188bc5c-1400-4000-8000-000000000000", 16, true))

	if ToJson(fss) != TestShardingConfigUuidFindShards_RA {
		t.Errorf("Should `%v` but `%v`", TestShardingConfigUuidFindShards_RA, ToJson(fss))
	}

	fss = sc.FindShards(ints.LimitSerialUUID(
		time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC),
	))

	if ToJson(fss) != TestShardingConfigUuidFindShards_RB {
		t.Errorf("Should `%v` but `%v`", TestShardingConfigUuidFindShards_RD, ToJson(fss))
	}

	fss = sc.FindShards(ints.UuidFromTextMust("0199bc5b-1400-4000-8000-000000000000", 16, true))

	if ToJson(fss) != TestShardingConfigUuidFindShards_RC {
		t.Errorf("Should `%v` but `%v`", TestShardingConfigUuidFindShards_RD, ToJson(fss))
	}

	fss = sc.FindShards(ints.UuidFromTextMust("0199bc5a-1400-4000-8000-000000000000", 16, true))

	if ToJson(fss) != TestShardingConfigUuidFindShards_RD {
		t.Errorf("Should `%v` but `%v`", TestShardingConfigUuidFindShards_RD, ToJson(fss))
	}
}

func TestShardingConfigUuidFindedBucket(t *testing.T) {
	sc := SharedConfigForTest()
	//sc := SharedConfigFromStringForTest()

	s := sc.FindShard(
		ints.LimitSerialUUID(
			time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC),
		),
		time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC),
	)

	id := ints.LimitSerialUUID(
		time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC),
	)
	b, ok := s.Bucket(id)

	if !ok || b != "a" {
		t.Errorf("Should be `a` and true but `%v` %v", b, ok)
	}

	id.UInt128 = id.Add(ints.UInt128FromInt(1))

	b, ok = s.Bucket(id)

	if !ok || b != "b" {
		t.Errorf("Should be `b` and true but `%v` %v", b, ok)
	}

	id.UInt128 = id.Add(ints.UInt128FromInt(1))

	b, ok = s.Bucket(id)

	if !ok || b != "c" {
		t.Errorf("Should be `c` and true but `%v` %v", b, ok)
	}

	id.UInt128 = id.Add(ints.UInt128FromInt(1))

	b, ok = s.Bucket(id)

	if !ok || b != "a" {
		t.Errorf("Should be `a` and true but `%v` %v", b, ok)
	}

	s = sc.FindShard(
		ints.UuidFromTextMust("0199bc5a-1400-4000-8000-000000000000", 16, true),
		time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC),
	)

	b, ok = s.Bucket(id)

	if ok || b != "" {
		t.Errorf("Should be ``(empty) and false but `%v` %v", b, ok)
	}
}

func BenchmarkShardingConfigUuidGlobalS1(b *testing.B) {
	sc := SharedConfigFromStringForTest()
	uuid := ints.LimitSerialUUID(
		time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC),
	)
	t := time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC)

	for i := 0; i < b.N; i++ {
		sc.FindShard(
			uuid,
			t,
		)
	}
}

func BenchmarkShardingConfigUuidGlobalS2(b *testing.B) {
	sc := SharedConfigFromStringForTest()
	uuid := ints.UuidFromTextMust("0188bc5c-1400-4000-8000-000000000000", 16, true)
	t := time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC)

	for i := 0; i < b.N; i++ {
		sc.FindShard(
			uuid,
			t,
		)
	}
}
