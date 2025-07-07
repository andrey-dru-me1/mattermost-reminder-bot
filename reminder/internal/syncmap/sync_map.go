package syncmap

import (
	"fmt"
	"sync"
)

type Map[K comparable, V any] struct {
	mu sync.RWMutex
	m  map[K]V
}

func New[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{m: make(map[K]V)}
}

func (m *Map[K, V]) Get(key K) (value V, ok bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, ok = m.m[key]
	return
}

func (m *Map[K, V]) Set(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.m[key] = value
}

func (m *Map[K, V]) Delete(key K) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.m, key)
}

func (m *Map[K, V]) Range(f func(key K, val V) error) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for key, val := range m.m {
		err := f(key, val)
		if err != nil {
			return fmt.Errorf(
				"error executing func from arg. key: %v, val: %v %w",
				key, val, err,
			)
		}
	}
	return nil
}

func (m *Map[K, V]) Apply(key K, f func(val V) V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if val, ok := m.m[key]; ok {
		m.m[key] = f(val)
	}
}
