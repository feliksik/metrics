package metrics

import (
	"sync"
)

// factory to construct a DimensionedMetric in a Vector with a LabelSet
type dimensionFactory func(n *node) DimensionedMetric

// a MetricCollection is a collection of metrics (counters, gauges and histrograms) in an n-dimensional space.
// The primary dimension is the name of the metric. The other dimensions are 0 or more labels associated with
// the metric.
//
// From a metric, other metrics can be created with additional LabelPairs, further differentiating/parititioning the metric.
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
	mu sync.RWMutex

	// vectors, by name
	counterVectors   vectorSet
	gaugeVectors     vectorSet
	histogramVectors vectorSet
}

// NewMetricCollection creates a new collections of metrics.
func NewMetricCollection() *MetricCollection {
	return &MetricCollection{
		counterVectors: vectorSet{
			namedSet: make(map[string]*MetricVector),
			factory: func(n *node) DimensionedMetric {
				return &DimensCounter{
					node: n,
				}
			},
		},
		gaugeVectors: vectorSet{
			namedSet: make(map[string]*MetricVector),
			factory: func(n *node) DimensionedMetric {
				return &DimensGauge{
					node: n,
				}
			},
		},
		histogramVectors: vectorSet{
			namedSet: make(map[string]*MetricVector),
			factory: func(n *node) DimensionedMetric {
				return &DimensHistogram{
					Histogram: NewHistogram(),
					node:      n,
				}
			},
		},
	}
}

// Counter returns a counter with no labels for the space of the given name.
func (m *MetricCollection) NewCounter(name string) DimensionedCounter {
	return m.counterVectors.vector(name).rootNode.Metric().(DimensionedCounter)
}

// Histogram returns a histogram with no labels for the space of the given name.
func (m *MetricCollection) NewHistogram(name string) DimensionedHistogram {
	return m.histogramVectors.vector(name).rootNode.Metric().(DimensionedHistogram)
}

// Counters returns the counters that are registered at the moment of calling.
// It returns an array, because iterating the original map requires a lock, so
// using a callback iterator doesn't give time guarantees we need.
// fixme the caller has to typecast to .(DimensionedCounter)... for now;
// as we don't want to instantiate a whole new array of another type, or duplicate
// the iterableMetrics code.
func (m *MetricCollection) Counters() vectorSet {
	return m.counterVectors
}

func (m *MetricCollection) IterCounters(func(counter DimensionedCounter)) {
	panic("not impl")
}

// not yet implemented
//func (m *MetricCollection) Gauge(name string) []DimensionedMetric {
//	return iterableMetrics(m.gaugeVectors)
//}
//

func (m *MetricCollection) Histograms(name string) []DimensionedMetric {
	return nil
}
