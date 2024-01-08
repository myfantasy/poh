package poh

const TestShardingConfigUuidFindShards_RD = `[
  {
    "epoch_name": "ep3",
    "version": "v1",
    "sc_cfg": {
      "from": "018df74f-8400-4000-8000-000000000000",
      "to": null,
      "name": "s4",
      "bucket_qty": 2,
      "bucket_names": [
        "a",
        "d"
      ]
    },
    "is_special": false
  }
]`

const TestShardingConfigUuidFindShards_RC = `[
  {
    "epoch_name": "ep3",
    "version": "v1",
    "sc_cfg": {
      "from": "0199bc5b-1400-4000-8000-000000000000",
      "to": "0199bc5b-1400-4000-8000-000000000000",
      "name": "",
      "bucket_qty": 1,
      "bucket_names": [
        "s1"
      ]
    },
    "is_special": true
  },
  {
    "epoch_name": "ep2",
    "version": "v1",
    "sc_cfg": {
      "from": "0199bc5b-1400-4000-8000-000000000000",
      "to": "0199bc5b-1400-4000-8000-000000000000",
      "name": "",
      "bucket_qty": 1,
      "bucket_names": [
        "s1"
      ]
    },
    "is_special": true
  },
  {
    "epoch_name": "ep1",
    "version": "v1",
    "sc_cfg": {
      "from": "0199bc5b-1400-4000-8000-000000000000",
      "to": "0199bc5b-1400-4000-8000-000000000000",
      "name": "",
      "bucket_qty": 1,
      "bucket_names": [
        "s1"
      ]
    },
    "is_special": true
  },
  {
    "epoch_name": "ep0",
    "version": "v1",
    "sc_cfg": {
      "from": "0199bc5b-1400-4000-8000-000000000000",
      "to": "0199bc5b-1400-4000-8000-000000000000",
      "name": "",
      "bucket_qty": 1,
      "bucket_names": [
        "s1"
      ]
    },
    "is_special": true
  }
]`

const TestShardingConfigUuidFindShards_RB = `[
  {
    "epoch_name": "ep3",
    "version": "v1",
    "sc_cfg": {
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
    "is_special": false
  },
  {
    "epoch_name": "ep2",
    "version": "v1",
    "sc_cfg": {
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
    "is_special": false
  },
  {
    "epoch_name": "ep1",
    "version": "v1",
    "sc_cfg": {
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
    "is_special": false
  },
  {
    "epoch_name": "ep0",
    "version": "v1",
    "sc_cfg": {
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
    "is_special": false
  }
]`

const TestShardingConfigUuidFindShards_RA = `[
  {
    "epoch_name": "ep3",
    "version": "v1",
    "sc_cfg": {
      "from": "0188bc5c-1400-4000-8000-000000000000",
      "to": "0188bc5c-1400-4000-8000-000000000000",
      "name": "",
      "bucket_qty": 1,
      "bucket_names": [
        "s2"
      ]
    },
    "is_special": true
  },
  {
    "epoch_name": "ep2",
    "version": "v1",
    "sc_cfg": {
      "from": "0188bc5c-1400-4000-8000-000000000000",
      "to": "0188bc5c-1400-4000-8000-000000000000",
      "name": "",
      "bucket_qty": 1,
      "bucket_names": [
        "s2"
      ]
    },
    "is_special": true
  },
  {
    "epoch_name": "ep1",
    "version": "v1",
    "sc_cfg": {
      "from": "0188bc5c-1400-4000-8000-000000000000",
      "to": "0188bc5c-1400-4000-8000-000000000000",
      "name": "",
      "bucket_qty": 1,
      "bucket_names": [
        "s2"
      ]
    },
    "is_special": true
  },
  {
    "epoch_name": "ep0",
    "version": "v1",
    "sc_cfg": {
      "from": "0188bc5c-1400-4000-8000-000000000000",
      "to": "0188bc5c-1400-4000-8000-000000000000",
      "name": "",
      "bucket_qty": 1,
      "bucket_names": [
        "s2"
      ]
    },
    "is_special": true
  }
]`
