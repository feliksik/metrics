package cloudwatch

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
	"github.com/leisurespecials/golib/observability/metrics"
	"time"
)

const maxCloudwatchDatums = 20 // max cloudwatch datums per PutMetricData request

// the percentiles we track; should become option, later.
var percentiles = []struct {
	s string
	f float64
}{
	{"p50", 0.50},
	{"p90", 0.90},
	{"p95", 0.95},
	{"p99", 0.99},
}

// Writer implements metrics.Writer
type Writer struct {
	collection *metrics.MetricCollection
	svc        cloudwatchiface.CloudWatchAPI
	namespace  string
	datumChan  chan *cloudwatch.MetricDatum
}

// NewWriter creates metrics Writer.
// It writes the collected metrics to Cloudwatch. So instantiate only 1 per Cloudwatch namespace, and
// don't do so repeatedly.
func NewWriter(collection *metrics.MetricCollection, svc cloudwatchiface.CloudWatchAPI, namespace string) metrics.Writer {
	m := &Writer{
		svc:        svc,
		namespace:  namespace,
		collection: collection,
		datumChan:  make(chan *cloudwatch.MetricDatum),
	}
	return m
}

// WriteLoop periodically writes the metrics.
// iff errHandler==nil, it doens't report errors.
func (w *Writer) WriteLoop(c <-chan time.Time, errHandler func(error)) {
	for range c {
		err := w.Write()
		if err != nil && errHandler != nil {
			errHandler(err)
		}
	}
}

// Send will fire an API request to CloudWatch with the latest stats for
// all metrics. This is a blocking call.
// For most cases, it is preferred that the WriteLoop method is used.
func (w *Writer) Write() error {
	now := time.Now()

	// there should be no need to lock, as, even if we do two Send()s in parallel,
	// the counter is reset when reading.
	sender := metricSender{
		namespace: w.namespace,
		svc:       w.svc,
	}

	for _, metric := range w.collection.Counters() {
		c := metric.(metrics.DimensionedCounter)
		value := c.ValueReset()
		if value == 0 {
			// should we actually make a call?
			// if the counter is no longer used, this costs money.
			// and it seems CloudWatch doesn't plot it.
			continue
		}
		d := &cloudwatch.MetricDatum{
			MetricName: aws.String(c.Name()),
			Dimensions: makeDimensions(c.Dimensions()),
			Value:      aws.Float64(value),
			Timestamp:  aws.Time(now),
			Unit:       aws.String(cloudwatch.StandardUnitCount),
		}
		sender.addDatum(d)
	}

	// todo RB-1163 process gauges and histograms

	sender.Flush()

	if sender.lastError != nil {
		return fmt.Errorf("%d errors occured during send (%d MetricData elements dropped). Last error was: %s", sender.errCount, sender.itemsDropped, sender.lastError.Error())
	}

	return nil
}

// metricSender sends MetricData objects to cloudwatch. It's job is to buffer items,
// allow flushing, and work in-fire-and-forget mode in case of errors.
// Not threadsafe! instantiate 1 for every batch you want to send.
type metricSender struct {
	svc          cloudwatchiface.CloudWatchAPI
	namespace    string
	datums       []*cloudwatch.MetricDatum
	lastError    error
	errCount     int
	itemsDropped int
}

// addDatum is NOT threadsafe.
func (m *metricSender) addDatum(d *cloudwatch.MetricDatum) {
	m.datums = append(m.datums, d)
	if len(m.datums) == maxCloudwatchDatums {
		m.Flush()
	}
}

func (m *metricSender) Flush() {
	if len(m.datums) == 0 {
		return
	}
	// send to AWS Cloudwatch. PutMetricData may block for a while
	_, err := m.svc.PutMetricData(&cloudwatch.PutMetricDataInput{
		Namespace:  aws.String(m.namespace),
		MetricData: m.datums,
	})

	if err != nil {
		m.lastError = err
		m.errCount++
		m.itemsDropped += len(m.datums)
	}

	// drop previous datums (whether success or failure writing)
	m.datums = make([]*cloudwatch.MetricDatum, 0, maxCloudwatchDatums)
}

func makeDimensions(labels metrics.LabelSet) []*cloudwatch.Dimension {
	dimensions := make([]*cloudwatch.Dimension, len(labels), len(labels))
	i := 0
	for k, v := range labels {
		dimensions[i] = &cloudwatch.Dimension{
			Name:  aws.String(k),
			Value: aws.String(v),
		}
		i++
	}
	return dimensions
}
