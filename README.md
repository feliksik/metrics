# Metrics

This package implements some common metrics operations.

It's inspired by the metrics implementation of Prometheus and Go-Kit, but they didn't suit
our needs for CloudWatch.


# Some thoughts

After thinking about this, I think the following components would make sense

* a `MetricsWriter` interface (not sure if this can be done uniformly), e.g. something like with 
```
type MetricsWriter interface {
  // WriteLoop calls Write() repeatedly, and triggers errHandler on error
  WriteLoop(c <-chan time.Time, errHandler func(error)) 

  // 
  Write() error  // flush stuff 
}
```


* a `MetricsFactory` interface
```
type MetricsFactory interface {
    NewCounter(name string) metrics.Counter
    NewGauge(name string) metrics.Gauge
    NewHistogram(name string) metrics.Histogram
}
````

A Provider (e.g. Prometheus, Cloudwatch) could implement the MetricsFactory (and possibly also the MetricsWriter?).

But in `cloudwatch/cloudwatch.go` actually the MetricsFactory is generic, and is used as a dependency for Cloudwatch 
Provider, because it must iterate the managed metrics. 
