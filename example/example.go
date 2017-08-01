package example

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
	"github.com/leisurespecials/golib/observability/metrics"
	"github.com/leisurespecials/golib/observability/metrics/writers/cloudwatch"
	"time"
)

func ExampleReadMetricsUseCase() {
	// register variables for use by your handler
	collection := metrics.NewMetricCollection()
	fillCollection(collection)

	// get all registered counters and print value.
	for _, c := range collection.Counters() {
		dc := c.(metrics.DimensionedCounter)
		val := dc.ValueReset()
		fmt.Printf("Counter %s has dimensions %s: %d", c.Name(), c.Dimensions(), val)
	}
}

func ExampleSendMetricsUseCase() {
	// register variables for use by your handler
	collection := metrics.NewMetricCollection()
	fillCollection(collection)

	// create a metrics writer
	var svc cloudwatchiface.CloudWatchAPI = nil // initialize a cloudwatch session
	namespace := "my-cloudwatch-namespace"
	cloudwatchWriter := cloudwatch.NewWriter(collection, svc, namespace)

	// let the writer asynchronously write periodically
	errHandler := func(err error) { fmt.Errorf(err.Error()) }
	go cloudwatchWriter.WriteLoop(time.Tick(30*time.Second), errHandler)

}

// you wouldn't have a getMetrics in your code, but this is to differentiate the
// readMetrics and sendMetrics usecase
func fillCollection(collection *metrics.MetricCollection) {

	reqCount := collection.Counter("requests")

	handleRequest := func(success bool) {
		// do work
		status := "ok"
		if success {
			status = "error"
		}

		// Labels add dimensions to your metrics, i.e. facets that you can filter.
		reqCount.With(metrics.LabelSet{
			"status": status,
			// watch out with creating too many label-values,
			// this is problematic for backends (either in perf or pricing).
			"caller": "Johnnie",
		}).Add(1)

	}

	// do operations
	handleRequest(true)
	handleRequest(false)
	handleRequest(true)
}
