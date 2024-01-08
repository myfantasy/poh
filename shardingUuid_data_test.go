package poh

import (
	"encoding/json"
	"time"

	"github.com/myfantasy/ints"
)

func SharedConfigFromStringForTest() *ShardingConfigUuid {
	s := `
{
	"epochs": [
	  {
		"epoch_name": "ep0",
		"from_time": null,
		"to_time": "2023-06-01T00:00:00Z",
		"ranges": [
		  {
			"from": null,
			"to": "0188bc5a-1400-4000-8000-000000000000",
			"name": "s0",
			"bucket_qty": 3,
			"bucket_names": [
			  "a",
			  "b",
			  "c"
			]
		  },
		  {
			"from": "0188bc5a-1400-4000-8000-000000000000",
			"to": "018a4e0a-1c00-4000-8000-000000000000",
			"name": "s1",
			"bucket_qty": 1,
			"bucket_names": [
			  "d"
			]
		  }
		],
		"special_cases": {
		  "0188bc5c-1400-4000-8000-000000000000": "s2",
		  "0199bc5b-1400-4000-8000-000000000000": "s1"
		}
	  },
	  {
		"epoch_name": "ep1",
		"from_time": "2023-06-01T00:00:00Z",
		"to_time": "2023-09-01T00:00:00Z",
		"ranges": [
		  {
			"from": null,
			"to": "0188bc5a-1400-4000-8000-000000000000",
			"name": "s0",
			"bucket_qty": 3,
			"bucket_names": [
			  "a",
			  "b",
			  "c"
			]
		  },
		  {
			"from": "0188bc5a-1400-4000-8000-000000000000",
			"to": "018a4e0a-1c00-4000-8000-000000000000",
			"name": "s1",
			"bucket_qty": 1,
			"bucket_names": [
			  "d"
			]
		  }
		],
		"special_cases": {
		  "0188bc5c-1400-4000-8000-000000000000": "s2",
		  "0199bc5b-1400-4000-8000-000000000000": "s1"
		}
	  },
	  {
		"epoch_name": "ep2",
		"from_time": "2023-09-01T00:00:00Z",
		"to_time": "2024-01-01T00:00:00Z",
		"ranges": [
		  {
			"from": null,
			"to": "0188bc5a-1400-4000-8000-000000000000",
			"name": "s0",
			"bucket_qty": 3,
			"bucket_names": [
			  "a",
			  "b",
			  "c"
			]
		  },
		  {
			"from": "0188bc5a-1400-4000-8000-000000000000",
			"to": "018a4e0a-1c00-4000-8000-000000000000",
			"name": "s1",
			"bucket_qty": 1,
			"bucket_names": [
			  "d"
			]
		  },
		  {
			"from": "018a4e0a-1c00-4000-8000-000000000000",
			"to": "018cc251-f400-4000-8000-000000000000",
			"name": "s2",
			"bucket_qty": 2,
			"bucket_names": [
			  "c",
			  "d"
			]
		  }
		],
		"special_cases": {
		  "0188bc5c-1400-4000-8000-000000000000": "s2",
		  "0199bc5b-1400-4000-8000-000000000000": "s1"
		}
	  },
	  {
		"epoch_name": "ep3",
		"from_time": "2024-01-01T00:00:00Z",
		"to_time": "2024-03-01T00:00:00Z",
		"ranges": [
		  {
			"from": null,
			"to": "0188bc5a-1400-4000-8000-000000000000",
			"name": "s0",
			"bucket_qty": 3,
			"bucket_names": [
			  "a",
			  "b",
			  "c"
			]
		  },
		  {
			"from": "0188bc5a-1400-4000-8000-000000000000",
			"to": "018a4e0a-1c00-4000-8000-000000000000",
			"name": "s1",
			"bucket_qty": 1,
			"bucket_names": [
			  "d"
			]
		  },
		  {
			"from": "018a4e0a-1c00-4000-8000-000000000000",
			"to": "018cc251-f400-4000-8000-000000000000",
			"name": "s2",
			"bucket_qty": 2,
			"bucket_names": [
			  "c",
			  "d"
			]
		  },
		  {
			"from": "018cc251-f400-4000-8000-000000000000",
			"to": "018df74f-8400-4000-8000-000000000000",
			"name": "s3",
			"bucket_qty": 2,
			"bucket_names": [
			  "b",
			  "e"
			]
		  },
		  {
			"from": "018df74f-8400-4000-8000-000000000000",
			"to": null,
			"name": "s4",
			"bucket_qty": 2,
			"bucket_names": [
			  "a",
			  "d"
			]
		  }
		],
		"special_cases": {
		  "0188bc5c-1400-4000-8000-000000000000": "s2",
		  "0199bc5b-1400-4000-8000-000000000000": "s1"
		}
	  }
	],
	"version": "v1"
}
`

	var sc ShardingConfigUuid

	err := json.Unmarshal([]byte(s), &sc)
	if err != nil {
		panic(err)
	}

	return &sc
}

func SharedConfigForTest() *ShardingConfigUuid {
	sc := &ShardingConfigUuid{
		Version: "v1",
		Epochs: ShardEpochConfigsUuid{
			ShardEpochConfigUuid{
				EpochName: "ep0",
				FromTime:  nil,
				ToTime:    ToP(time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC)),
				Ranges: ShardRangeConfigsUuid{
					ShardRangeConfigUuid{
						From: nil,
						To: ToP(ints.LimitSerialUUID(
							time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC),
						)),
						Name:           "s0",
						BucketQuantity: 3,
						BucketNames:    []string{"a", "b", "c"},
					},
					ShardRangeConfigUuid{
						From: ToP(ints.LimitSerialUUID(
							time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC),
						)),
						To: ToP(ints.LimitSerialUUID(
							time.Date(2023, 9, 1, 0, 0, 0, 0, time.UTC),
						)),
						Name:           "s1",
						BucketQuantity: 1,
						BucketNames:    []string{"d"},
					},
				},
				SpecialCases: ints.UuidMap[string]{
					ints.UuidFromTextMust("0199bc5b-1400-4000-8000-000000000000", 16, true): "s1",
					ints.UuidFromTextMust("0188bc5c-1400-4000-8000-000000000000", 16, true): "s2",
				},
			},
			ShardEpochConfigUuid{
				EpochName: "ep1",
				FromTime:  ToP(time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC)),
				ToTime:    ToP(time.Date(2023, 9, 1, 0, 0, 0, 0, time.UTC)),
				Ranges: ShardRangeConfigsUuid{
					ShardRangeConfigUuid{
						From: nil,
						To: ToP(ints.LimitSerialUUID(
							time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC),
						)),
						Name:           "s0",
						BucketQuantity: 3,
						BucketNames:    []string{"a", "b", "c"},
					},
					ShardRangeConfigUuid{
						From: ToP(ints.LimitSerialUUID(
							time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC),
						)),
						To: ToP(ints.LimitSerialUUID(
							time.Date(2023, 9, 1, 0, 0, 0, 0, time.UTC),
						)),
						Name:           "s1",
						BucketQuantity: 1,
						BucketNames:    []string{"d"},
					},
				},
				SpecialCases: ints.UuidMap[string]{
					ints.UuidFromTextMust("0199bc5b-1400-4000-8000-000000000000", 16, true): "s1",
					ints.UuidFromTextMust("0188bc5c-1400-4000-8000-000000000000", 16, true): "s2",
				},
			},
			ShardEpochConfigUuid{
				EpochName: "ep2",
				FromTime:  ToP(time.Date(2023, 9, 1, 0, 0, 0, 0, time.UTC)),
				ToTime:    ToP(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
				Ranges: ShardRangeConfigsUuid{
					ShardRangeConfigUuid{
						From: nil,
						To: ToP(ints.LimitSerialUUID(
							time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC),
						)),
						Name:           "s0",
						BucketQuantity: 3,
						BucketNames:    []string{"a", "b", "c"},
					},
					ShardRangeConfigUuid{
						From: ToP(ints.LimitSerialUUID(
							time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC),
						)),
						To: ToP(ints.LimitSerialUUID(
							time.Date(2023, 9, 1, 0, 0, 0, 0, time.UTC),
						)),
						Name:           "s1",
						BucketQuantity: 1,
						BucketNames:    []string{"d"},
					},
					ShardRangeConfigUuid{
						From: ToP(ints.LimitSerialUUID(
							time.Date(2023, 9, 1, 0, 0, 0, 0, time.UTC),
						)),
						To: ToP(ints.LimitSerialUUID(
							time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						)),
						Name:           "s2",
						BucketQuantity: 2,
						BucketNames:    []string{"c", "d"},
					},
				},
				SpecialCases: ints.UuidMap[string]{
					ints.UuidFromTextMust("0199bc5b-1400-4000-8000-000000000000", 16, true): "s1",
					ints.UuidFromTextMust("0188bc5c-1400-4000-8000-000000000000", 16, true): "s2",
				},
			},
			ShardEpochConfigUuid{
				EpochName: "ep3",
				FromTime:  ToP(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
				ToTime:    ToP(time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)),
				Ranges: ShardRangeConfigsUuid{
					ShardRangeConfigUuid{
						From: nil,
						To: ToP(ints.LimitSerialUUID(
							time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC),
						)),
						Name:           "s0",
						BucketQuantity: 3,
						BucketNames:    []string{"a", "b", "c"},
					},
					ShardRangeConfigUuid{
						From: ToP(ints.LimitSerialUUID(
							time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC),
						)),
						To: ToP(ints.LimitSerialUUID(
							time.Date(2023, 9, 1, 0, 0, 0, 0, time.UTC),
						)),
						Name:           "s1",
						BucketQuantity: 1,
						BucketNames:    []string{"d"},
					},
					ShardRangeConfigUuid{
						From: ToP(ints.LimitSerialUUID(
							time.Date(2023, 9, 1, 0, 0, 0, 0, time.UTC),
						)),
						To: ToP(ints.LimitSerialUUID(
							time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						)),
						Name:           "s2",
						BucketQuantity: 2,
						BucketNames:    []string{"c", "d"},
					},
					ShardRangeConfigUuid{
						From: ToP(ints.LimitSerialUUID(
							time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						)),
						To: ToP(ints.LimitSerialUUID(
							time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
						)),
						Name:           "s3",
						BucketQuantity: 2,
						BucketNames:    []string{"b", "e"},
					},
					ShardRangeConfigUuid{
						From: ToP(ints.LimitSerialUUID(
							time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
						)),
						To:             nil,
						Name:           "s4",
						BucketQuantity: 2,
						BucketNames:    []string{"a", "d"},
					},
				},
				SpecialCases: ints.UuidMap[string]{
					ints.UuidFromTextMust("0199bc5b-1400-4000-8000-000000000000", 16, true): "s1",
					ints.UuidFromTextMust("0188bc5c-1400-4000-8000-000000000000", 16, true): "s2",
				},
			},
		},
	}
	sc.Sort()

	return sc
}
