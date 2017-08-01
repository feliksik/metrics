package metrics

// LabelSet maps labels to values. Neither should contain the '|' character
// (as it's used for encoding -- we could fix this).
type LabelSet map[string]string

type Dimension struct {
	Key   string
	Value string
}

/////////////////////////////////////////

type DimensCounter struct {
	Counter
	name   string
	labels LabelSet
	space  *MetricVector
}

func (d *DimensCounter) With(labels LabelSet) DimensionedCounter {
	mergedLabels := mergeLabels(d.labels, labels)

	// fixme: could use a shortcut/cache here in the DimensCounter to apply only these labels,
	// and fetch it from there if possible.
	return d.space.GetMetricWith(mergedLabels).(DimensionedCounter)
}

func (d *DimensCounter) Dimensions() LabelSet {
	return d.labels
}

func (d *DimensCounter) Name() string {
	return d.name
}

type DimensGauge struct {
	DimensionedGauge
	name   string
	labels LabelSet
	space  *MetricVector
}

func (d *DimensGauge) With(labels LabelSet) DimensionedGauge {
	return d.space.GetMetricWith(labels).(DimensionedGauge)
}

func (d *DimensGauge) Dimensions() LabelSet {
	return d.labels
}

func (d *DimensGauge) Name() string {
	return d.name
}

type DimensHistogram struct {
	DimensionedHistogram
	name   string
	labels LabelSet
	space  *MetricVector
}

func (d *DimensHistogram) With(labels LabelSet) DimensionedHistogram {
	return d.space.GetMetricWith(labels).(DimensionedHistogram)
}

func (d *DimensHistogram) Dimensions() LabelSet {
	return d.labels
}

func (d *DimensHistogram) Name() string {
	return d.name
}
