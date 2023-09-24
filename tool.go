package poh

import (
	"github.com/cespare/xxhash"
	"github.com/myfantasy/ints"
)

func ToSet[T comparable](s []T) map[T]struct{} {
	res := make(map[T]struct{}, len(s))

	for _, v := range s {
		res[v] = struct{}{}
	}

	return res
}

// BytesToIntHashXX generates int hash from []byte
func BytesToIntHashXX(body []byte) int {
	res := xxhash.Sum64(body)

	return int(res)
}

// StringToIntHashXX generates int hash from string
func StringToIntHashXX(s string) int {
	return BytesToIntHashXX([]byte(s))
}

// Int128ToIntHashXX generates int hash from ints.UInt128
func Int128ToIntHashXX(i ints.UInt128) int {
	bts := i.AsBytes()
	return BytesToIntHashXX(bts[:])
}
