package fixedmap

import (
	"container/ring"
	"sync"
)

type FixedMap struct {
	r  *ring.Ring
	m  map[interface{}]interface{}
	mu sync.RWMutex
}

func NewFixLenMap(len int) *FixedMap {
	return &FixedMap{
		m: map[interface{}]interface{}{},
		r: ring.New(len),
	}
}

func (f *FixedMap) Delete(key interface{}) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.m, key)
}

func (f *FixedMap) Range(fn func(key, value interface{}) bool) {
	// todo: how to fix it if a FixedMap.Delete() called inside fn?
	f.mu.Lock()
	defer f.mu.Unlock()
	for i := range f.m {
		if cont := fn(i, f.m[i]); !cont {
			return
		}
	}
}

func (f *FixedMap) LoadAndDelete(key interface{}) (value interface{}, loaded bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	value, loaded = f.m[key]
	delete(f.m, key)
	return
}

func (f *FixedMap) LoadOrStore(key interface{}, value interface{}) (actual interface{}, loaded bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	actual, loaded = f.m[key]
	if loaded {
		return
	}

	expireKey := f.keyEnqueue(key)
	delete(f.m, expireKey)
	f.m[key] = value
	return value, false
}

func (f *FixedMap) Load(key interface{}) (value interface{}, exist bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	value, exist = f.m[key]
	return
}

func (f *FixedMap) Store(key interface{}, value interface{}) {
	f.mu.Lock()
	defer f.mu.Unlock()

	expireKey := f.keyEnqueue(key)
	if expireKey != nil {
		delete(f.m, expireKey)
	}
	f.m[key] = value
}

func (f *FixedMap) keyEnqueue(key interface{}) (dequeue interface{}) {
	if f.r.Value != nil {
		dequeue = f.r.Value
	}

	f.r.Value = key
	f.r = f.r.Next()
	return
}
