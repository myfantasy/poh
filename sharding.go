package poh

import "time"

// ShardEpochConfig config of one epoch of sharding
type ShardEpochConfig struct {
	FromTime       *time.Time `json:"from_time" db:"from_time"`
	ToTime         *time.Time `json:"to_time" db:"to_time"`
	From           *int       `json:"from" db:"from"`
	To             *int       `json:"to" db:"to"`
	Name           string     `json:"name" db:"name"`
	BucketQuantity int        `json:"bucket_qty" db:"bucket_qty"`
	BucketNames    []string   `json:"bucket_names" db:"bucket_names"`
}

// FindShard find shard info; returns nil when not found and *FoundedShard when found (*FoundedShard.Version is not set)
func (sc *ShardEpochConfig) FindShard(metaID int) *FoundedShard {
	if sc.From != nil && metaID < *sc.From {
		return nil
	}
	if sc.To != nil && metaID > *sc.To {
		return nil
	}

	part64 := metaID % sc.BucketQuantity
	part := int(part64)

	return &FoundedShard{
		EpochName:      sc.Name,
		BucketNumber:   part,
		BucketQuantity: sc.BucketQuantity,
		BucketName:     sc.BucketNames[part],
	}
}

// FindShardD find shard info in setted Date; returns nil when not found and *FoundedShard when found (*FoundedShard.Version is not set)
func (sc *ShardEpochConfig) FindShardD(metaID int, searchDate time.Time) *FoundedShard {
	if sc.FromTime != nil && searchDate.Before(*sc.FromTime) {
		return nil
	}
	if sc.ToTime != nil && searchDate.After(*sc.ToTime) {
		return nil
	}

	return sc.FindShard(metaID)
}

type ShardingConfig struct {
	Epochs  []ShardEpochConfig `json:"epochs" db:"epochs"`
	Version string             `json:"version" db:"version"`
}

// FoundedShard is shard info
type FoundedShard struct {
	EpochName      string `json:"epoch_name" db:"epoch_name"`
	BucketNumber   int    `json:"bucket_number" db:"bucket_number"`
	BucketQuantity int    `json:"bucket_quantity" db:"bucket_quantity"`
	BucketName     string `json:"bucket_name" db:"bucket_name"`
	Version        string `json:"version" db:"version"`
}

// FindShards find shards info in config order;
func (sc *ShardingConfig) FindShards(metaID int) []*FoundedShard {
	res := make([]*FoundedShard, 0, 1)

	for _, sec := range sc.Epochs {
		fs := sec.FindShard(metaID)
		if fs != nil {
			fs.Version = sc.Version
		}
	}

	return res
}

// FindShardsD find shards info in config order in setted Date;
func (sc *ShardingConfig) FindShardsD(metaID int, searchDate time.Time) []*FoundedShard {
	res := make([]*FoundedShard, 0, 1)

	for _, sec := range sc.Epochs {
		fs := sec.FindShardD(metaID, searchDate)
		if fs != nil {
			fs.Version = sc.Version
		}
	}

	return res
}
