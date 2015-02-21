package main

import (
	"fmt"
	"github.com/protip/metricusdb/metricusdb"
	"github.com/protip/metricusdb/metricusdb/riakse"
	"runtime/debug"
	"strconv"
	"time"
)

var SE = &riakse.RiakSE{Address: "192.168.1.66:8087", StreamIndex: "famous", C: riakse.Continuum{StreamLines: make(map[string]*riakse.StreamLine)}}

func main() {
	hostCount := 500
	metricsPerHost := 500

	fmt.Println("Running flush test")
	se := &riakse.RiakSE{
		Address:     "192.168.1.66:8087",
		StreamIndex: "famous",
		C: riakse.Continuum{
			StreamLines: make(map[string]*riakse.StreamLine),
			FlushChan:   make(chan *riakse.Window, 50),
		},
	}

	insertFunc := func() {
		start := time.Now()
		var tick int
		ticker := time.NewTicker(time.Second)
		select {
		case <-ticker.C:
			for start.Add(300 * time.Second).After(time.Now()) {
				for h := 0; h < hostCount; h++ {
					for m := 0; m < metricsPerHost; m++ {
						se.InsertMetric(metricusdb.InsertQuery{
							Stream: metricusdb.Stream{
								Namespace: "modusOps",
								Name:      "metric-" + strconv.Itoa(m),
								Dimensions: metricusdb.Dimensions{
									{Name: "Host", Value: "local-" + strconv.Itoa(h)},
								}},
							Metric: metricusdb.Metric{time.Now().UTC().UnixNano(), float64(tick)},
						})
					}
				}
				tick++

			}
		}
	}

	quit := make(chan bool)
	go se.C.Flusher(quit)
	go se.WindowSink(se.C.FlushChan)
	go insertFunc()
	start := time.Now()
	for start.Add(600 * time.Second).After(time.Now()) {
		var windowCount int
		var metricCount int
		for _, sl := range se.C.StreamLines {
			windowCount += sl.Windows.Len()
			for e := sl.Windows.Front(); e != nil; e = e.Next() {
				w := e.Value.(*riakse.Window)
				metricCount += len(w.Metrics)
			}
		}
		fmt.Println(len(se.C.StreamLines), " open stream lines have ", windowCount, " open windows and ", metricCount, " metrics")
		debug.FreeOSMemory()
		time.Sleep(2 * time.Second)
	}
}
