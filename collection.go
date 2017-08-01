package metrics

import (
	"github.com/leisurespecials/golib/observability/metrics/internal"
	"sync"
)

// NewMetricCollection creates a new collections of metrics.
func NewMetricCollection() *MetricCollection {
	return &MetricCollection{
		counterSpaces:   make(map[string]*MetricVector),
		gaugeSpaces:     make(map[string]*MetricVector),
		histogramSpaces: make(map[string]*MetricVector),
	}
}

// a MetricCollection is a collection of metrics (counters, gauges and histrograms) in an n-dimensional space.
// The primary dimension is the name of the metric. The other dimensions are 0 or more labels associated with
// the metric.
//
// From a metric, other metrics can be created with additional Labels, further differentiating/parititioning the metric.
//
// A MetricCollection is unaware of any metrics backend.
//
// It can be argued that the name of metrics is just another dimension, so we could implement it without name.
// However:
// - backends tend to require a name (AWS, Influx), and we would have to have a 'magic label name' to designate
//   the actual name used for the backend. E.g. the label "name" would be the name, but then the actual label
//   would be unusable in the backend
// - It seems to make sense have *have* an actual name, as it is the thing that guarantees all metrics of that name
//   have the same unit (e.g. Count, time (ms), Gauge (MB/s)) and can be aggregated; the labels are merely a
//   way to partition them.
//
type MetricCollection struct {
	mu              sync.RWMutex
	counterSpaces   map[string]*MetricVector // mapping metric name to all the dimensions of that metric
	gaugeSpaces     map[string]*MetricVector
	histogramSpaces map[string]*MetricVector
}

// Counter returns a counter with no labels for the space of the given name.
func (m *MetricCollection) Counter(name string) DimensionedCounter {
	m.mu.RLock()
	space := m.counterSpaces[name]
	m.mu.RUnlock()

	if space == nil {
		space = NewMetricVector(name)
		// fixme maybe this should be done in NewMetricVector, using a newCounter(space, name, lbls) constructor
		space.createMetric = func(labels LabelSet) DimensionedMetric {
			return &DimensCounter{
				space:   space,
				name:    name,
				labels:  labels,
				Counter: &internal.Counter{}, // some generic counter
			}
		}

		m.mu.Lock()
		m.counterSpaces[name] = space
		m.mu.Unlock()

	}
	return space.GetMetricWith(LabelSet{}).(DimensionedCounter)
}

// iterate over the metrics of a certain type
// Fixme this isn't the nicest way to allow the client to iterate over the metrics
func iterableMetrics(mv map[string]*MetricVector) []DimensionedMetric {
	allMetrics := make([]DimensionedMetric, 0, 0)
	for _, space := range mv {
		for _, metric := range space.List() {
			allMetrics = append(allMetrics, metric)
		}
	}
	return allMetrics
}

// Counters returns the counters that are registered at the moment of calling.
// It returns an array, because iterating the original map requires a lock, so
// using a callback iterator doesn't give time guarantees we need.
// fixme the caller has to typecast to .(DimensionedCounter)... for now;
// as we don't want to instantiate a whole new array of another type, or duplicate
// the iterableMetrics code.
func (m *MetricCollection) Counters() []DimensionedMetric {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return iterableMetrics(m.counterSpaces)
}

// not yet implemented
//func (m *MetricCollection) Gauge(name string) []DimensionedMetric {
//	m.mu.RLock()
//	defer m.mu.RUnlock()
//	return iterableMetrics(m.gaugeSpaces)
//}
//
//func (m *MetricCollection) Histogram(name string) []DimensionedMetric {
//	m.mu.RLock()
//	defer m.mu.RUnlock()
//	return iterableMetrics(m.histogramSpaces)
//}
