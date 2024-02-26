package poh

import (
	"encoding/json"
	"sort"
	"time"

	"golang.org/x/exp/constraints"
)

type ShardRangeConfig[T constraints.Ordered] struct {
	From           *T       `json:"from" db:"from"`
	To             *T       `json:"to" db:"to"`
	Name           string   `json:"name" db:"name"`
	BucketQuantity int      `json:"bucket_qty" db:"bucket_qty"`
	BucketNames    []string `json:"bucket_names" db:"bucket_names"`
}
type ShardRangeConfigs[T constraints.Ordered] []ShardRangeConfig[T]

// ShardEpochConfig config of one epoch of sharding
type ShardEpochConfig[T constraints.Ordered] struct {
	EpochName    string               `json:"epoch_name" db:"epoch_name"`
	EpochInfo    json.RawMessage      `json:"epoch_info" db:"epoch_info"`
	FromTime     *time.Time           `json:"from_time" db:"from_time"`
	ToTime       *time.Time           `json:"to_time" db:"to_time"`
	Ranges       ShardRangeConfigs[T] `json:"ranges" db:"ranges"`
	SpecialCases map[T]string         `json:"special_cases" db:"special_cases"`
}

type ShardEpochConfigs[T constraints.Ordered] []ShardEpochConfig[T]

type ShardingConfig[T constraints.Ordered] struct {
	Epochs  ShardEpochConfigs[T] `json:"epochs" db:"epochs"`
	Version string               `json:"version" db:"version"`
}

// FoundedShard is shard info
type FoundedShard[T constraints.Ordered] struct {
	EpochName        string               `json:"epoch_name" db:"epoch_name"`
	EpochInfo        json.RawMessage      `json:"epoch_info" db:"epoch_info"`
	Version          string               `json:"version" db:"version"`
	ShardRangeConfig *ShardRangeConfig[T] `json:"sc_cfg" db:"version"`
	IsSpecial        bool                 `json:"is_special" db:"is_special"`
}

func LessOrdered[T constraints.Ordered](a, b *T) bool {
	if a == nil && b == nil {
		return false
	}
	if a == nil {
		return true
	}
	if b == nil {
		return false
	}
	return *a < *b
}
func CompareOrdered[T constraints.Ordered](from, to *T, val T) int {
	if from == nil && to == nil {
		return 0
	}
	if from == nil {
		if *to >= val {
			return 0
		}
		return 1
	}
	if to == nil {
		if *from <= val {
			return 0
		}
		return -1
	}

	if *to < val {
		return 1
	}

	if *from > val {
		return -1
	}

	return 0
}

func LessTime(a, b *time.Time) bool {
	if a == nil && b == nil {
		return false
	}
	if a == nil {
		return true
	}
	if b == nil {
		return false
	}
	return (*a).Before(*b)
}
func CompareTime(from, to *time.Time, val time.Time) int {
	if from == nil && to == nil {
		return 0
	}
	if from == nil {
		if val.Before(*to) || val.Equal(*to) {
			return 0
		}
		return 1
	}
	if to == nil {
		if (*from).Before(val) || (*from).Equal(val) {
			return 0
		}
		return -1
	}

	if (*to).Before(val) {
		return 1
	}

	if val.Before(*from) {
		return -1
	}

	return 0
}

// type Lessable[T any] interface {
// 	Less(a T) bool
// }

// func LessLessable[T any](a *Lessable[T], b *T) bool {
// 	if a == nil && b == nil {
// 		return false
// 	}
// 	if a == nil {
// 		return true
// 	}
// 	if b == nil {
// 		return false
// 	}
// 	return (*a).Less(*b)
// }
// func LessLessablePoint[T any](a *Lessable[*T], b *T) bool {
// 	if a == nil && b == nil {
// 		return false
// 	}
// 	if a == nil {
// 		return true
// 	}
// 	if b == nil {
// 		return false
// 	}
// 	return (*a).Less(b)
// }

func (rsc ShardRangeConfigs[T]) Sort() {
	sort.Slice(rsc, func(i, j int) bool {
		if LessOrdered(rsc[i].From, rsc[j].From) {
			return true
		}

		if LessOrdered(rsc[j].From, rsc[i].From) {
			return false
		}

		return LessOrdered(rsc[i].To, rsc[j].To)
	})
}

func (sec ShardEpochConfigs[T]) Sort() {
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

func (sc ShardRangeConfig[T]) Compare(id T) int {
	return CompareOrdered(sc.From, sc.To, id)
}

func (sec ShardEpochConfig[T]) Compare(shardingTime time.Time) int {
	return CompareTime(sec.FromTime, sec.ToTime, shardingTime)
}

func (rsc ShardRangeConfigs[T]) FindShard(id T) *ShardRangeConfig[T] {
	ix, ok := sort.Find(len(rsc), func(i int) int {
		return rsc[i].Compare(id)
	})
	if !ok {
		return nil
	}

	return &rsc[ix]
}

func (sec ShardEpochConfigs[T]) FindShard(id T, shardingTime time.Time) *FoundedShard[T] {
	ix, ok := sort.Find(len(sec), func(i int) int {
		return sec[i].Compare(shardingTime)
	})
	if !ok {
		return nil
	}

	bName, ok := sec[ix].SpecialCases[id]
	if ok {
		return &FoundedShard[T]{
			EpochName: sec[ix].EpochName,
			EpochInfo: sec[ix].EpochInfo,
			IsSpecial: true,
			ShardRangeConfig: &ShardRangeConfig[T]{
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
		return &FoundedShard[T]{
			EpochName:        sec[ix].EpochName,
			EpochInfo:        sec[ix].EpochInfo,
			ShardRangeConfig: src,
		}
	}

	return nil
}

func (sec ShardEpochConfigs[T]) FindShards(id T) (res []*FoundedShard[T]) {
	for i := len(sec) - 1; i >= 0; i-- {
		bName, ok := sec[i].SpecialCases[id]
		if ok {
			res = append(res,
				&FoundedShard[T]{
					EpochName: sec[i].EpochName,
					EpochInfo: sec[i].EpochInfo,
					IsSpecial: true,
					ShardRangeConfig: &ShardRangeConfig[T]{
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
				&FoundedShard[T]{
					EpochName:        sec[i].EpochName,
					EpochInfo:        sec[i].EpochInfo,
					ShardRangeConfig: s,
				},
			)
		}
	}

	return res
}

func (sc ShardingConfig[T]) FindShard(id T, shardingTime time.Time) *FoundedShard[T] {
	fs := sc.Epochs.FindShard(id, shardingTime)
	if fs != nil {
		fs.Version = sc.Version
	}

	return fs
}

func (sc ShardingConfig[T]) FindShards(id T) (res []*FoundedShard[T]) {
	res = sc.Epochs.FindShards(id)

	for _, s := range res {
		s.Version = sc.Version
	}

	return res
}

func (sc ShardingConfig[T]) Sort() {
	sc.Epochs.Sort()
}

func (fs *FoundedShard[T]) Bucket(id T, conv func(id T) (hash int)) (bucket string, ok bool) {
	if fs == nil {
		return "", false
	}

	if fs.ShardRangeConfig == nil {
		return "", false
	}

	if fs.ShardRangeConfig.BucketQuantity <= 0 {
		return "", false
	}

	rem := conv(id) % fs.ShardRangeConfig.BucketQuantity

	return fs.ShardRangeConfig.BucketNames[rem], true
}

func BucketInt(id int, fs *FoundedShard[int]) (bucket string, ok bool) {
	if fs == nil {
		return "", false
	}

	if fs.ShardRangeConfig == nil {
		return "", false
	}

	if fs.ShardRangeConfig.BucketQuantity <= 0 {
		return "", false
	}

	rem := id % fs.ShardRangeConfig.BucketQuantity

	return fs.ShardRangeConfig.BucketNames[rem], true
}

func BucketString(id string, fs *FoundedShard[string]) (bucket string, ok bool) {
	if fs == nil {
		return "", false
	}

	if fs.ShardRangeConfig == nil {
		return "", false
	}

	if fs.ShardRangeConfig.BucketQuantity <= 0 {
		return "", false
	}

	rem := StringToIntHashXX(id) % fs.ShardRangeConfig.BucketQuantity

	return fs.ShardRangeConfig.BucketNames[rem], true
}
