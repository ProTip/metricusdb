package riakse

import (
	"fmt"
	"github.com/kr/pretty"
	"github.com/protip/metricusdb/metricusdb"
	"runtime/debug"
	"strconv"
	"testing"
	"time"
)

var SE = &RiakSE{Address: "192.168.1.66:8087", StreamIndex: "famous", C: Continuum{StreamLines: make(map[string]*StreamLine)}}

func TestHello(t *testing.T) {
	fmt.Println("Hello world!")
}

func TestListMetrics(t *testing.T) {
	se := &RiakSE{Address: "192.168.1.66:8087", StreamIndex: "famous"}
	streams, err := se.ListStreams(metricusdb.Dimensions{
		{Name: "nodes.0_s", Value: "natasha"},
	})
	if err != nil {
		fmt.Println(err.Error())
	}
	for _, stream := range streams {
		pretty.Println(stream.Identifier())
	}
}

//func TestInsertMetric(t *testing.T) {
//	se := &RiakSE{Address: "192.168.1.66:8087", StreamIndex: "famous", c: Continuum{StreamLines: make(map[string]*StreamLine)}}
//	se.InsertMetric(metricusdb.InsertQuery{
//		Stream: metricusdb.Stream{
//			Namespace: "modusOps", Dimensions: metricusdb.Dimensions{
//				{Name: "Host", Value: "local"},
//			}},
//		Metric: metricusdb.Metric{time.Now().UTC().UnixNano(), 5.0},
//	})
//	se.InsertMetric(metricusdb.InsertQuery{
//		Stream: metricusdb.Stream{
//			Namespace: "modusOps", Dimensions: metricusdb.Dimensions{
//				{Name: "Host", Value: "local"},
//			}},
//		Metric: metricusdb.Metric{time.Now().UTC().Add(time.Minute).UnixNano(), 5.0},
//	})
//	for k, sl := range se.c.StreamLines {
//		var count int
//		for e := sl.Windows.Front(); e != nil; e = e.Next() {
//			w := e.Value.(*Window)
//			count += len(w.Metrics)
//		}
//		fmt.Println("StreamLineFor ", k, " has ", sl.Windows.Len(), " open windows and ", count, " metrics")
//		firstWindow := sl.Windows.Front().Value.(*Window)
//		pretty.Println(firstWindow.Metrics)
//	}
//}

func TestFlushMetric(t *testing.T) {
	hostCount := 1000
	metricsPerHost := 500

	fmt.Println("Running flush test")
	se := &RiakSE{
		Address:     "192.168.1.66:8087",
		StreamIndex: "famous",
		C: Continuum{
			StreamLines: make(map[string]*StreamLine),
			FlushChan:   make(chan *Window, 50),
		},
	}

	insertFunc := func() {
		start := time.Now()
		var tick int
		ticker := time.NewTicker(time.Second)
		select {
		case _ <- ticker:
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
				w := e.Value.(*Window)
				metricCount += len(w.Metrics)
			}
		}
		fmt.Println(len(se.C.StreamLines), " open stream lines have ", windowCount, " open windows and ", metricCount, " metrics")
		debug.FreeOSMemory()
		time.Sleep(2 * time.Second)
	}
}

func BenchmarkInsertMetrics(b *testing.B) {
	se := &RiakSE{Address: "192.168.1.66:8087", StreamIndex: "famous", C: Continuum{StreamLines: make(map[string]*StreamLine)}}
	metricTime := time.Now().UTC().UnixNano()
	for i := 0; i < b.N; i++ {
		se.InsertMetric(metricusdb.InsertQuery{
			Stream: metricusdb.Stream{
				Namespace: "modusOps", Dimensions: metricusdb.Dimensions{
					{Name: "Host", Value: "local"},
				}},
			Metric: metricusdb.Metric{metricTime, 5.0},
		})
	}
	//for k, sl := range se.c.StreamLines {
	//	fmt.Println("Stream ", k, " has ", sl.Windows.Len(), " open windows")
	//}
}

//func TestTargetToDimensions(t *testing.T) {
//	dimensions := TargetToDimensions("natasha.production.stack-servers.5")
//	fmt.Println(dimensions)
//}

//func TestMetricsToTree(t *testing.T) {
//	metrics := []string{
//		`natasha.production.stack.5.cpu_load`,
//		`natasha.production.stack.6.cpu_load`,
//		`natasha.production.rds.5.load`,
//		`natasha.development.stack.1.cpu_load`,
//	}
//	tree := MetricsToTree("natasha.*.*", metrics, 4)
//	for _, node := range tree {
//		fmt.Printf("%+v\n", *node)
//	}
//}
