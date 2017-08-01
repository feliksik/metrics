package metrics

import (
	"bytes"
	"sort"
	"sync"
)

// a vectorSet is a collection of named vectors, of the same type
// (e.g. all counters, histogram, or gauges)
type vectorSet struct {
	namedSet map[string]*MetricVector
	mu       sync.RWMutex
	factory  dimensionFactory
}

//  get the metricVector with given name.
func (v vectorSet) vector(name string) *MetricVector {
	v.mu.RLock()
	r, ok := v.namedSet[name]
	v.mu.RUnlock()
	if ok {
		return r
	}

	v.mu.Lock()
	defer v.mu.Unlock()
	newVector := NewMetricVector(name, v.factory)
	v.namedSet[name] = newVector
	return newVector
}

// MetricVector contains metrics, differentiated per label set.
// so it is a metric with different dimensions, as defined by the set of labels.
type MetricVector struct {
	name             string
	rootNode         *node
	mtx              sync.RWMutex
	dimensionFactory dimensionFactory
}

func NewMetricVector(name string, factory dimensionFactory) *MetricVector {
	m := &MetricVector{
		name:             name,
		dimensionFactory: factory,
		rootNode:         nil, // set below, cross-reference
	}
	m.rootNode = newNode(m, LabelPairs{})
	return m
}

// Reset empties the current MetricVector and returns a new MetricVector with the old
// contents. Reset a MetricVector to get an immutable copy suitable for walking.
func (m *MetricVector) Reset() *MetricVector {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	n := NewMetricVector(m.name, m.dimensionFactory)
	n.rootNode, m.rootNode = m.rootNode, n.rootNode
	return n
}
