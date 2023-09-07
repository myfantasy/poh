package poh

func ToSet[T comparable](s []T) map[T]struct{} {
	res := make(map[T]struct{}, len(s))

	for _, v := range s {
		res[v] = struct{}{}
	}

	return res
}
