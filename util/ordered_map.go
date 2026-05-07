package util

type OrderedMap[K comparable, V any] struct {
	keys []K
	vals map[K]V
}

func (m *OrderedMap[K, V]) Set(k K, v V) {
	if _, ok := m.vals[k]; !ok {
		m.keys = append(m.keys, k)
	}
	m.vals[k] = v
}

func (m *OrderedMap[K, V]) Get(k K) (V, bool) {
	v, ok := m.vals[k]
	return v, ok
}

func (m *OrderedMap[K, V]) Range(fn func(k K, v V)) {
	for _, k := range m.keys {
		fn(k, m.vals[k])
	}
}

func NewOrderedMap[K comparable, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{vals: make(map[K]V)}
}
