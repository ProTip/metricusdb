package metricusdb

import (
	"fmt"
	"testing"
)

func TestPipeLine(t *testing.T) {
	p := NewPipeLine()
	p.AddProcessor(ProcessorMap["averageSeries"]())
	go p.Run()
	p.Input <- SeriesList{
		{
			Name: "SeriesA",
			Metrics: []Metric{
				{
					Time:  0,
					Value: 5},
			},
		},
		{
			Name: "SeriesB",
			Metrics: []Metric{
				{
					Time:  0,
					Value: 10},
			},
		},
	}
	result := <-p.Output
	fmt.Println(result)
}
