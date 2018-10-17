package pkg

import "sync"

type SafeMap struct {
	m sync.Map
}

func NewSafeMap() *SafeMap {
	return &SafeMap{}
}

func (mm *SafeMap) Insert(k, v interface{}) {
	mm.m.Store(k, v)
}

func (mm *SafeMap) Remove(k interface{}) {
	mm.m.Delete(k)
}

func (mm *SafeMap) GetNum() int {
	num := 0
	mm.m.Range(func(key, value interface{}) bool {
		num += 1
		return true
	})
	return num
}
