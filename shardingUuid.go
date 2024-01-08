package poh

import (
	"sort"
	"time"

	"github.com/myfantasy/ints"
)

type ShardRangeConfigUuid struct {
	From           *ints.Uuid `json:"from" db:"from"`
	To             *ints.Uuid `json:"to" db:"to"`
	Name           string     `json:"name" db:"name"`
	BucketQuantity int        `json:"bucket_qty" db:"bucket_qty"`
	BucketNames    []string   `json:"bucket_names" db:"bucket_names"`
}
type ShardRangeConfigsUuid []ShardRangeConfigUuid

// ShardEpochConfigUuid config of one epoch of sharding
type ShardEpochConfigUuid struct {
	EpochName    string                `json:"epoch_name" db:"epoch_name"`
	FromTime     *time.Time            `json:"from_time" db:"from_time"`
	ToTime       *time.Time            `json:"to_time" db:"to_time"`
	Ranges       ShardRangeConfigsUuid `json:"ranges" db:"ranges"`
	SpecialCases ints.UuidMap[string]  `json:"special_cases" db:"special_cases"`
}

type ShardEpochConfigsUuid []ShardEpochConfigUuid

type ShardingConfigUuid struct {
	Epochs  ShardEpochConfigsUuid `json:"epochs" db:"epochs"`
	Version string                `json:"version" db:"version"`
}

// FoundedShardUuid is shard info
type FoundedShardUuid struct {
	EpochName        string                `json:"epoch_name" db:"epoch_name"`
	Version          string                `json:"version" db:"version"`
	ShardRangeConfig *ShardRangeConfigUuid `json:"sc_cfg" db:"version"`
	IsSpecial        bool                  `json:"is_special" db:"is_special"`
}

func LessOrderedUuid(a, b *ints.Uuid) bool {
	if a == nil && b == nil {
		return false
	}
	if a == nil {
		return true
	}
	if b == nil {
		return false
	}
	return a.Less(*b)
}
func CompareOrderedUuid(from, to *ints.Uuid, val ints.Uuid) int {
	if from == nil && to == nil {
		return 0
	}
	if from == nil {
		if val.Less(*to) {
			return 0
		}
		return 1
	}
	if to == nil {
		if from.Equal(val) || from.Less(val) {
			return 0
		}
		return -1
	}

	if to.Less(val) {
		return 1
	}

	if val.Less(*from) {
		return -1
	}

	return 0
}

func (rsc ShardRangeConfigsUuid) Sort() {
	sort.Slice(rsc, func(i, j int) bool {
		if LessOrderedUuid(rsc[i].From, rsc[j].From) {
			return true
		}

		if LessOrderedUuid(rsc[j].From, rsc[i].From) {
			return false
		}

		return LessOrderedUuid(rsc[i].To, rsc[j].To)
	})
}

func (sec ShardEpochConfigsUuid) Sort() {
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

func (sc ShardRangeConfigUuid) Compare(id ints.Uuid) int {
	return CompareOrderedUuid(sc.From, sc.To, id)
}

func (sec ShardEpochConfigUuid) Compare(shardingTime time.Time) int {
	return CompareTime(sec.FromTime, sec.ToTime, shardingTime)
}

func (rsc ShardRangeConfigsUuid) FindShard(id ints.Uuid) *ShardRangeConfigUuid {
	ix, ok := sort.Find(len(rsc), func(i int) int {
		return rsc[i].Compare(id)
	})
	if !ok {
		return nil
	}

	return &rsc[ix]
}

func (sec ShardEpochConfigsUuid) FindShard(id ints.Uuid, shardingTime time.Time) *FoundedShardUuid {
	ix, ok := sort.Find(len(sec), func(i int) int {
		return sec[i].Compare(shardingTime)
	})
	if !ok {
		return nil
	}

	bName, ok := sec[ix].SpecialCases[id]
	if ok {
		return &FoundedShardUuid{
			EpochName: sec[ix].EpochName,
			IsSpecial: true,
			ShardRangeConfig: &ShardRangeConfigUuid{
				From:           &id,
				To:             &id,
				Name:           "",
				BucketQuantity: 1,
				BucketNames:    []string{bName},
			},
		}
	}

	src := sec[ix].Ranges.FindShard(id)
	if src != nil {
		return &FoundedShardUuid{
			EpochName:        sec[ix].EpochName,
			ShardRangeConfig: src,
		}
	}

	return nil
}

func (sec ShardEpochConfigsUuid) FindShards(id ints.Uuid) (res []*FoundedShardUuid) {
	for i := len(sec) - 1; i >= 0; i-- {
		bName, ok := sec[i].SpecialCases[id]
		if ok {
			res = append(res,
				&FoundedShardUuid{
					EpochName: sec[i].EpochName,
					IsSpecial: true,
					ShardRangeConfig: &ShardRangeConfigUuid{
						From:           &id,
						To:             &id,
						Name:           "",
						BucketQuantity: 1,
						BucketNames:    []string{bName},
					},
				},
			)
			continue
		}

		s := sec[i].Ranges.FindShard(id)
		if s != nil {
			res = append(res,
				&FoundedShardUuid{
					EpochName:        sec[i].EpochName,
					ShardRangeConfig: s,
				},
			)
		}
	}

	return res
}

func (sc ShardingConfigUuid) FindShard(id ints.Uuid, shardingTime time.Time) *FoundedShardUuid {
	fs := sc.Epochs.FindShard(id, shardingTime)
	if fs != nil {
		fs.Version = sc.Version
	}

	return fs
}

func (sc ShardingConfigUuid) FindShards(id ints.Uuid) (res []*FoundedShardUuid) {
	res = sc.Epochs.FindShards(id)

	for _, s := range res {
		s.Version = sc.Version
	}

	return res
}

func (sc ShardingConfigUuid) Sort() {
	sc.Epochs.Sort()
}

func (fs *FoundedShardUuid) Bucket(id ints.Uuid) (bucket string, ok bool) {
	if fs == nil {
		return "", false
	}

	if fs.ShardRangeConfig == nil {
		return "", false
	}

	if fs.ShardRangeConfig.BucketQuantity <= 0 {
		return "", false
	}

	_, rem := id.DivUint64(uint64(fs.ShardRangeConfig.BucketQuantity))

	return fs.ShardRangeConfig.BucketNames[int(rem.UInt64())], true
}
