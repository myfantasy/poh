package poh

// Pool - your pool with connect info
type Pool struct {
	Items map[string]PoolItem
	Clear func([]PoolItem)
}

// PoolItem - your item of pool
type PoolItem struct {
	Hash string
	Meta string
	Conn interface{}
}

// Reload - Reload pool with new items
func (p Pool) Reload(items []PoolItem) {

	nm := make(map[string]PoolItem)

	rm := make([]PoolItem, 0)

	for _, i := range items {
		if v, e := p.Items[i.Hash]; e {
			nm[i.Hash] = i
		} else {
			nm[i.Hash] = v
		}
	}

	for k, v := range p.Items {
		if _, e := nm[k]; e {
			rm = append(rm, v)
		}
	}

	p.Items = nm

	if p.Clear != nil {
		p.Clear(rm)
	}
}

//func (p Pool) Get
