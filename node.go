package metrics

import "sync"

type LabelPairs []*pair

// Labels converts list of labels to list of pairs
func Labels(labelValues ...string) LabelPairs {
	if len(labelValues)%2 != 0 {
		labelValues = append(labelValues, "unknown")
	}

	lbl := LabelPairs{}
	for i := 0; i < len(labelValues); i += 2 {
		lbl = append(lbl, &pair{labelValues[i], labelValues[i+1]})
	}
	return lbl
}

// private fields, immutable.
type pair struct{ name, value string }

func (p *pair) Name() string {
	return p.name
}
func (p *pair) Value() string {
	return p.value
}

// node exists at a specific point in the N-dimensional vector space of all
// possible label values. The node has a metric, and child nodes
// with greater specificity.
type node struct {
	metricVector *MetricVector // vector containing this node
	myLabels     LabelPairs    // labels of this node

	mtx sync.RWMutex

	children  map[pair]*node    // childNodes, by label
	dimMetric DimensionedMetric // the metric associated with this node
}

// create a new node with the given labels.
func newNode(vector *MetricVector, labels LabelPairs) *node {
	n := &node{
		metricVector: vector,
		myLabels:     labels,
		children:     make(map[pair]*node),
	}
	n.dimMetric = vector.dimensionFactory(n)
	return n
}

func (n *node) Name() string {
	return n.metricVector.name
}

// Metric gives the metric of this node
func (n *node) Metric() DimensionedMetric {
	return n.dimMetric
}

// Dimensions gives back the LabelPairs
func (n *node) Dimensions() LabelPairs {
	return n.myLabels
}

// specialize returns a specialized sub-node.
func (n *node) specializedMetric(subLabelsValues []string) DimensionedMetric {
	subLabels := Labels(subLabelsValues...)
	allLabels := append(n.myLabels, subLabels)
	return n.specializeSkipping(allLabels, len(n.myLabels)).Metric()
}

// metricWith returns the metric of the subnode that is further specialized with given labels.
// startIx specifies the number of pairs to skip in the LabelValues list, where we must start
// to specialize.
func (n *node) specializeSkipping(labels LabelPairs, startIx int) *node {
	if startIx >= len(labels) {
		return n
	}

	nextPair := *labels[startIx]

	n.mtx.RLock()
	child, ok := n.children[nextPair]
	n.mtx.RUnlock()
	if !ok {
		child = newNode(n.metricVector, labels[:startIx+1]) // all labels of parent, plus the nextPair
		n.mtx.Lock()
		n.children[nextPair] = child
		n.mtx.Unlock()
	}

	return child
}
