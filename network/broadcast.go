package network

import "sync"

type broadcastGroup struct {
	mu  sync.Mutex
	arr []int
}

func (b *broadcastGroup) Append(id int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.arr = append(b.arr, id)
	return
}

func (b *broadcastGroup) Remove(id int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 原地修改arr并过滤id
	n := 0
	for _, i := range b.arr {
		if i != id {
			b.arr[n] = i
			n++
		}
	}
	b.arr = b.arr[:n]
}

func (b *broadcastGroup) GetArray() []int {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.arr
}
