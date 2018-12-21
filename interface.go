package metrics

import (
	"time"

	"github.com/trafficstars/fastmetrics/worker"
)

type Worker = metricworker.Worker
type TimingValues = metricworker.TimingValues
type TimingValue = metricworker.TimingValue

type WorkerGauge interface {
	Worker

	Increment() int64
	Decrement() int64
	Add(delta int64) int64
	Set(newValue int64)
	SetValuePointer(newValuePtr *int64)
}

type WorkerGaugeFunc interface {
	Worker
}

type WorkerCount interface {
	Worker

	Increment() uint64
	Add(delta uint64) uint64
	Set(newValue uint64)
	SetValuePointer(newValuePtr *uint64)
}

type WorkerTiming interface {
	Worker

	ConsiderValue(d time.Duration)
	GetValuePointers() *TimingValues
}