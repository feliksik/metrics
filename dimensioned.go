package metrics

type Dimension struct {
	Key   string
	Value string
}

/////////////////////////////////////////

type DimensCounter struct {
	Counter
	*node // the node from which this Dimension spun off
}

func (d *DimensCounter) With(labelValues ...string) DimensionedCounter {
	return d.node.specializedMetric(labelValues).(DimensionedCounter)
}

type DimensGauge struct {
	DimensionedGauge
	*node // the node from which this Dimension spun off
}

func (d *DimensGauge) With(labelValues ...string) DimensionedGauge {
	return d.node.specializedMetric(labelValues).(DimensionedGauge)
}

type DimensHistogram struct {
	Histogram
	*node // the node from which this Dimension spun off
}

func (d *DimensHistogram) With(labelValues ...string) DimensionedHistogram {
	return d.node.specializedMetric(labelValues).(DimensionedHistogram)
}
