package poh

import (
	"sort"
	"time"

	"github.com/myfantasy/ints"
)

type ShardRangeConfigInt128 struct {
	From           *ints.UInt128 `json:"from" db:"from"`
	To             *ints.UInt128 `json:"to" db:"to"`
	Name           string        `json:"name" db:"name"`
	BucketQuantity int           `json:"bucket_qty" db:"bucket_qty"`
	BucketNames    []string      `json:"bucket_names" db:"bucket_names"`
}
type ShardRangeConfigsInt128 []ShardRangeConfigInt128

// ShardEpochConfigInt128 config of one epoch of sharding
type ShardEpochConfigInt128 struct {
	EpochName string                  `json:"epoch_name" db:"epoch_name"`
	FromTime  *time.Time              `json:"from_time" db:"from_time"`
	ToTime    *time.Time              `json:"to_time" db:"to_time"`
	Ranges    ShardRangeConfigsInt128 `json:"ranges" db:"ranges"`
}

type ShardEpochConfigsInt128 []ShardEpochConfigInt128

type ShardingConfigInt128 struct {
	Epochs  ShardEpochConfigsInt128 `json:"epochs" db:"epochs"`
	Version string                  `json:"version" db:"version"`
}

// FoundedShardInt128 is shard info
type FoundedShardInt128 struct {
	EpochName        string                  `json:"epoch_name" db:"epoch_name"`
	Version          string                  `json:"version" db:"version"`
	ShardRangeConfig *ShardRangeConfigInt128 `json:"sc_cfg" db:"version"`
}

func LessOrderedInt128(a, b *ints.UInt128) bool {
	if a == nil && b == nil {
		return false
	}
	if a == nil {
		return true
	}
	if b == nil {
		return false
	}
	return a.Less(b)
}
func CompareOrderedInt128(from, to *ints.UInt128, val ints.UInt128) int {
	if from == nil && to == nil {
		return 0
	}
	if from == nil {
		if val.Less(to) {
			return 0
		}
		return 1
	}
	if to == nil {
		if from.Equal(&val) || from.Less(&val) {
			return 0
		}
		return -1
	}

	if to.Less(&val) {
		return 1
	}

	if val.Less(from) {
		return -1
	}

	return 0
}

func (rsc ShardRangeConfigsInt128) Sort() {
	sort.Slice(rsc, func(i, j int) bool {
		if LessOrderedInt128(rsc[i].From, rsc[j].From) {
			return true
		}

		if LessOrderedInt128(rsc[j].From, rsc[i].From) {
			return false
		}

		return LessOrderedInt128(rsc[i].To, rsc[j].To)
	})
}

func (sec ShardEpochConfigsInt128) Sort() {
	sort.Slice(sec, func(i, j int) bool {
		if LessTime(sec[i].FromTime, sec[j].FromTime) {
			return true
		}
		if LessTime(sec[j].FromTime, sec[i].FromTime) {
			return false
		}

		return LessTime(sec[i].ToTime, sec[j].ToTime)
	})

	for _, s := range sec {
		s.Ranges.Sort()
	}
}

func (sc ShardRangeConfigInt128) Compare(id ints.UInt128) int {
	return CompareOrderedInt128(sc.From, sc.To, id)
}

func (sec ShardEpochConfigInt128) Compare(shardingTime time.Time) int {
	return CompareTime(sec.FromTime, sec.ToTime, shardingTime)
}

func (rsc ShardRangeConfigsInt128) FindShard(id ints.UInt128) *ShardRangeConfigInt128 {
	ix, ok := sort.Find(len(rsc), func(i int) int {
		return rsc[i].Compare(id)
	})
	if !ok {
		return nil
	}

	return &rsc[ix]
}

func (sec ShardEpochConfigsInt128) FindShard(id ints.UInt128, shardingTime time.Time) *FoundedShardInt128 {
	ix, ok := sort.Find(len(sec), func(i int) int {
		return sec[i].Compare(shardingTime)
	})
	if !ok {
		return nil
	}

	return &FoundedShardInt128{
		EpochName:        sec[ix].EpochName,
		ShardRangeConfig: sec[ix].Ranges.FindShard(id),
	}
}

func (sec ShardEpochConfigsInt128) FindShards(id ints.UInt128) (res []*FoundedShardInt128) {
	for i := len(sec) - 1; i >= 0; i-- {
		s := sec[i].Ranges.FindShard(id)
		if s != nil {
			res = append(res,
				&FoundedShardInt128{
					EpochName:        sec[i].EpochName,
					ShardRangeConfig: s,
				},
			)
		}
	}

	return res
}

func (sc ShardingConfigInt128) FindShard(id ints.UInt128, shardingTime time.Time) *FoundedShardInt128 {
	fs := sc.Epochs.FindShard(id, shardingTime)
	if fs != nil {
		fs.Version = sc.Version
	}

	return fs
}

func (sc ShardingConfigInt128) FindShards(id ints.UInt128) (res []*FoundedShardInt128) {
	res = sc.Epochs.FindShards(id)

	for _, s := range res {
		s.Version = sc.Version
	}

	return res
}

func (sc ShardingConfigInt128) Sort() {
	sc.Epochs.Sort()
}
