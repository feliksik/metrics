# Metrics

This package implements some common metrics operations.

It's inspired by the metrics implementation of Prometheus and Go-Kit, but they didn't suit
our needs for CloudWatch.

See the example.go for example.


The package consists of 2 parts:

## metrics.MetricCollection

Allows you to create metrics in a multi-dimensional space (i.e. with labels or 'dimensions') easily.
Currently only counter is implemented, but gauge and histogram must be done

## metrics/writers/

Provides packages with metrics writers.

Only Cloudwatch is implemented, but it should be easy to implement Prometheus/Influx writers
if you want.



