package metricworker

import (
	"sync"
	"sync/atomic"
	"time"
)

type workerGauge struct {
	sync.Mutex

	id         int64
	sender     MetricSender
	metricsKey string
	valuePtr   *int64
	//stopChan         chan bool
	senderLoopFuncID uint64
	interval         time.Duration
	running          uint64
	isGCEnabled      uint64
	uselessCounter   uint64
}

func NewWorkerGauge(sender MetricSender, metricsKey string) *workerGauge {
	w := &workerGauge{}
	w.id = atomic.AddInt64(&workersCount, 1)
	w.sender = sender
	w.metricsKey = metricsKey
	w.valuePtr = &[]int64{0}[0]
	//w.stopChan = make(chan bool)
	return w
}

func (w *workerGauge) SetGCEnabled(enabled bool) {
	if w == nil {
		return
	}
	if enabled {
		atomic.StoreUint64(&w.isGCEnabled, 1)
	} else {
		atomic.StoreUint64(&w.isGCEnabled, 0)
	}
}

func (w *workerGauge) IsGCEnabled() bool {
	if w == nil {
		return false
	}
	return atomic.LoadUint64(&w.isGCEnabled) > 0
}

func (w *workerGauge) IsRunning() bool {
	if w == nil {
		return false
	}
	return atomic.LoadUint64(&w.running) > 0
}

func (w *workerGauge) GetType() MetricType {
	return MetricTypeGauge
}

func (w *workerGauge) Increment() int64 {
	if w == nil {
		return 0
	}
	return atomic.AddInt64(w.valuePtr, 1)
}

func (w *workerGauge) Decrement() int64 {
	if w == nil {
		return 0
	}
	return atomic.AddInt64(w.valuePtr, -1)
}

func (w *workerGauge) Add(delta int64) int64 {
	if w == nil {
		return 0
	}
	return atomic.AddInt64(w.valuePtr, delta)
}

func (w *workerGauge) Set(newValue int64) {
	if w == nil {
		return
	}
	atomic.StoreInt64(w.valuePtr, newValue)
}

func (w *workerGauge) Get() int64 {
	if w == nil {
		return 0
	}
	return atomic.LoadInt64(w.valuePtr)
}

func (w *workerGauge) GetKey() string {
	if w == nil {
		return ``
	}
	return w.metricsKey
}

func (w *workerGauge) SetValuePointer(newValuePtr *int64) {
	if w == nil {
		return
	}
	w.valuePtr = newValuePtr
}

func (w *workerGauge) doSend() {
	value := w.Get()
	if w.IsGCEnabled() {
		if value == 0 {
			if atomic.AddUint64(&w.uselessCounter, 1) > gcUselessLimit {
				if w.IsRunning() {
					go w.Stop()
				}
			}
			return
		} else {
			atomic.StoreUint64(&w.uselessCounter, 0)
		}
	}
	if w.sender == nil {
		return
	}
	dataMap := map[string]int{
		w.metricsKey: int(value),
	}
	w.sender.Send(string(MetricTypeGauge), dataMap) // TODO: process the returned error somehow
}

func (w *workerGauge) Run(interval time.Duration) {
	if w == nil {
		return
	}
	w.Lock()
	defer w.Unlock()
	if w.IsRunning() {
		return
	}
	w.senderLoopFuncID = appendToSenderLoop(interval, func() {
		w.doSend()
	})
	w.interval = interval
	atomic.StoreUint64(&w.uselessCounter, 0)
	atomic.StoreUint64(&w.running, 1)
	return
}

func (w *workerGauge) Stop() {
	if w == nil {
		return
	}
	if !w.IsRunning() {
		return
	}
	w.Lock()
	defer w.Unlock()
	//w.stopChan <- true
	removeFromSenderLoop(w.interval, w.senderLoopFuncID)
	w.interval = time.Duration(0)
	atomic.StoreUint64(&w.running, 0)
}
