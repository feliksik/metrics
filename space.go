package metrics

import (
	"bytes"
	"sort"
	"sync"
)

// FlatLabelSet is a represenation of label set, that is comparable so we can use it a map key.
// An int64 hash would be more efficient (see Prometheus), but for now it's a silly string
// as it is easier to implement.
type FlatLabelSet string

// mergeLabels adds the given labels, overriding where necessary.
func mergeLabels(a, b LabelSet) LabelSet {
	merged := LabelSet{}
	for k, v := range a {
		merged[k] = v
	}
	for k, v := range b {
		merged[k] = v
	}
	return merged
}

func flattenLabels(b LabelSet) FlatLabelSet {
	// iteration order is not guaranteed, so we need to sort.
	keyList := make(sort.StringSlice, len(b), len(b))
	i := 0
	for k, _ := range b {
		keyList[i] = k
		i++
	}
	keyList.Sort()
	buf := bytes.NewBuffer([]byte{})
	for _, k := range keyList {
		v := b[k]
		buf.WriteString(k)
		buf.WriteByte('|')
		buf.WriteString(v)
		buf.WriteByte('|')
	}
	return FlatLabelSet(buf.String())
}

// MetricVector contains metrics, differentiated per label set.
type MetricVector struct {
	// factory method creates the metric of the right type
	name         string
	index        map[FlatLabelSet]DimensionedMetric
	list         []DimensionedMetric
	createMetric func(labels LabelSet) DimensionedMetric
	mu           sync.RWMutex
}

func NewMetricVector(name string) *MetricVector {
	m := &MetricVector{
		name: name,

		// index and list are duplicating.
		// index is for fast indexing.
		// slice allows lock-free iteration.
		index:        make(map[FlatLabelSet]DimensionedMetric),
		list:         make([]DimensionedMetric, 0),
		createMetric: nil, // has to be set by caller
	}
	return m
}

// GetMetricWith returns the metric with the specified labels.
func (m *MetricVector) GetMetricWith(labels LabelSet) DimensionedMetric {
	hash := flattenLabels(labels)
	m.mu.RLock()
	dimMetric, ok := m.index[hash]
	m.mu.RUnlock()
	if ok {
		return dimMetric
	}

	dimMetric = m.createMetric(labels)

	m.mu.Lock()
	m.index[hash] = dimMetric
	m.list = append(m.list, dimMetric)
	m.mu.Unlock()

	return dimMetric
}

// List returns the metric list.
// Fixme Not this is not immutable, so handle with care.
func (m *MetricVector) List() []DimensionedMetric {
	return m.list
}
