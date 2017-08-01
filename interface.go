package metrics

import "time"

// some whould argue that a name "myName" is just a tag named "name" with value "myName"
// but this is not what InfluxDB, Cloudwatch and Prometheus seem to be doing in their API.
// So we're not trying to be innovative here.
// Also we're not sure if making an interface for this Name() adds much.
type DimensionedMetric interface {
	Name() string
	Dimensions() LabelSet
}

// interfaces inspired by gokit
// We are not using the gokit impl, because it should be redesigned.
//
// https://gophers.slack.com/archives/C04S3T99A/p1494935965976164
// https://github.com/go-kit/kit/issues/528
// https://github.com/go-kit/kit/issues/529

// Counter describes a metric that accumulates values monotonically.
// An example of a counter is the number of received HTTP requests.

// interface for counter, gauge and histogram.
// you can send a value to a MetricCollector, and it'll process it in the way it sees fit.

type Counter interface {
	Add(delta float64)
	Value() float64
	ValueReset() float64
}

type Gauge interface {
	Set(value float64)
}

type Histogram interface {
	Observe(value float64)
}

type DimensionedCounter interface {
	Counter
	DimensionedMetric
	With(labels LabelSet) DimensionedCounter
}

// DimensionedGauge describes a metric that takes specific values over time.
// An example of a gauge is the current depth of a job queue.
type DimensionedGauge interface {
	Gauge
	DimensionedMetric
	With(labels LabelSet) DimensionedGauge
}

// DimensionedHistogram describes a metric that takes repeated observations of the same
// kind of thing, and produces a statistical summary of those observations,
// typically expressed as quantiles or buckets. An example of a histogram is
// HTTP request latencies.
type DimensionedHistogram interface {
	Histogram
	DimensionedMetric
	With(labels LabelSet) DimensionedHistogram
}

// MetricSender sends metrics to a metrics backend
type Writer interface {
	// Write collects the metrics and writes to the backend.
	Write() error

	// WriteLoop invokes Write every time the passed
	// channel fires. This method blocks until the channel is closed, so clients
	// probably want to run it in its own goroutine. For typical usage, create a
	// time.Ticker and pass its C channel to this method.
	//
	// implementation must *not* crash if errHandler==nil.
	//
	// If you make this more frequent, your counters will have higher sampling resolution,
	// but you'll be doing more write requests to your backend.
	//
	// For example:  AWS can only graph every minute (or 5 minutes in the free tier),
	// so anything below that should suffice. You can write faster if you want to see updates
	// faster, but with AWS this costs money.
	WriteLoop(c <-chan time.Time, errHandler func(error))
}
