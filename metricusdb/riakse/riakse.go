package riakse

import (
	"bytes"
	"container/list"
	"encoding/binary"
	"fmt"
	"github.com/protip/metricusdb/metricusdb"
	"github.com/tpjg/goriakpbc"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

type RiakSE struct {
	Address     string
	StreamIndex string
	C           Continuum
}

type Window struct {
	stream    metricusdb.Stream
	FlushTime time.Time //Flush the window after this amount of time
	StartTime time.Time //Represents the start of the time window
	EndTime   time.Time
	Metrics   []metricusdb.Metric
}

func (w *Window) InsertMetric(m metricusdb.Metric) {
	w.Metrics = append(w.Metrics, m)
}

type StreamLine struct {
	Stream  metricusdb.Stream
	Windows *list.List
}

func (s *StreamLine) OpenWindowFor(q metricusdb.InsertQuery) {
	metricTime := time.Unix(q.Metric.Time/int64(time.Second), int64(math.Mod(float64(q.Metric.Time), float64(time.Second))))
	//metricTimeTrunc := metricTime.Truncate(time.Minute)

	startTime := time.Unix((metricTime.Unix()/600)*600, 0)
	endTime := startTime.Add(600 * time.Second)

	window := &Window{
		stream:    q.Stream,
		FlushTime: time.Now().UTC().Add(310 * time.Second),
		StartTime: startTime,
		EndTime:   endTime,
		Metrics:   make([]metricusdb.Metric, 0, 300),
	}
	s.Windows.PushBack(window)
}

func (s *StreamLine) InsertMetric(q metricusdb.InsertQuery) {
	if !s.insertMetric(q) {
		s.OpenWindowFor(q)
		s.insertMetric(q)
	}
}

func (s *StreamLine) insertMetric(q metricusdb.InsertQuery) bool {
	inserted := false
	for e := s.Windows.Front(); e != nil; e = e.Next() {
		window := e.Value.(*Window)
		if q.Metric.Time >= window.StartTime.UnixNano() && q.Metric.Time < window.EndTime.UnixNano() {
			window.InsertMetric(q.Metric)
			inserted = true
		}
	}
	return inserted
}

type Continuum struct {
	StreamLines map[string]*StreamLine
	FlushChan   chan *Window
}

func (c *Continuum) Flusher(quit chan bool) {
	for {
		select {
		case <-quit:
			break
		default:
			now := time.Now().UTC()
			for _, sl := range c.StreamLines {
				for e := sl.Windows.Front(); e != nil; {
					w := e.Value.(*Window)
					if w.FlushTime.Before(now) {
						n := e.Next()
						sl.Windows.Remove(e)
						e = n
						c.FlushChan <- w
					} else {
						break
					}
				}
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func (c *Continuum) StreamLineIsOpen(s metricusdb.Stream) {

}

func (c *Continuum) EnsureStreamLineOpen(stream *metricusdb.Stream) {
	if _, ok := c.StreamLines[stream.Identifier()]; !ok {
		c.StreamLines[stream.Identifier()] = &StreamLine{
			Stream:  *stream,
			Windows: list.New(),
		}
	}
}

func (c *Continuum) StreamLineForStream(stream *metricusdb.Stream) *StreamLine {
	c.EnsureStreamLineOpen(stream)
	return c.StreamLines[stream.Identifier()]
}

func (c *Continuum) InsertMetric(q metricusdb.InsertQuery) {
	streamLine := c.StreamLineForStream(&q.Stream)
	streamLine.InsertMetric(q)
}

func (se *RiakSE) ListStreams(dimensions metricusdb.Dimensions) ([]metricusdb.Stream, error) {
	terms := make([]string, 0, len(dimensions))
	for _, e := range dimensions {
		terms = append(terms, fmt.Sprintf("%s:%s", e.Name, e.Value))
	}
	query := strings.Join(terms, " AND ")
	fmt.Println("Querying metrics with query: ", query)
	c := riak.NewClient(se.Address)
	streams := make([]metricusdb.Stream, 0, 0)
	if docs, _, _, err := c.Search(&riak.Search{Q: "*", Filter: query, Index: se.StreamIndex}); err == nil {
		stream := metricusdb.Stream{
			Dimensions: make([]metricusdb.Dimension, 0, 0),
		}
		for _, doc := range docs {
			for k, v := range doc {
				if (len(k) >= 3 && k[:3] == "_yz") || k == "score" {
					continue
				}
				stream.Dimensions = append(stream.Dimensions, metricusdb.Dimension{Name: k, Value: string(v)})
			}
			stream.Name = string(doc["_yz_rk"])
			streams = append(streams, stream)
		}
	}
	return streams, nil
}

func (se *RiakSE) InsertMetric(q metricusdb.InsertQuery) {
	se.C.InsertMetric(q)
}

func (se *RiakSE) WindowSink(in chan *Window) {
	riak.ConnectClientPool(se.Address, 200)
	bucket, _ := riak.NewBucketType("metrics", "metrics")

	conc := 200
	wg := sync.WaitGroup{}
	sendC := make(chan *riak.RObject, 200)
	for x := 0; x < conc; x++ {
		go func(in chan *riak.RObject) {
			for o := range in {
				o.Store()
			}
			wg.Done()
		}(sendC)
		wg.Add(1)
	}

	for w := range in {
		//fmt.Println("Received flushed window for ", w.stream.Identifier())
		obj := bucket.NewObject(w.stream.Identifier() + strconv.Itoa(int(w.StartTime.Unix())))
		obj.ContentType = "application/json"
		if len(w.Metrics) > 0 {
			buffer := new(bytes.Buffer)
			err := binary.Write(buffer, binary.LittleEndian, w.Metrics)
			if err != nil {
				panic(err.Error())
			}
			obj.Data = buffer.Bytes()
			sendC <- obj
		}
	}
	close(sendC)
	wg.Wait()
}
