package metricusdb

import ()

type Pipeline struct {
	Input      chan SeriesList
	Output     chan SeriesList
	Processors []PipelineProcessor
}

type Metric struct {
	Time  int64
	Value float64
}

type Series struct {
	Name    string
	Metrics []Metric
}

type SeriesList []Series

type PipelineProcessor func(SeriesList) SeriesList

func NewPipeLine() *Pipeline {
	return &Pipeline{
		Input:      make(chan SeriesList),
		Output:     make(chan SeriesList),
		Processors: make([]PipelineProcessor, 0)}
}

func (p *Pipeline) AddProcessor(pp PipelineProcessor) {
	p.Processors = append(p.Processors, pp)
}

func (p *Pipeline) Run() {
	for sl := range p.Input {
		for _, pp := range p.Processors {
			sl = pp(sl)
		}
		p.Output <- sl
	}
	close(p.Output)
}

func NewAverageSeriesPP() (pp PipelineProcessor) {
	pp = func(sl SeriesList) SeriesList {
		var acc float64
		for i, _ := range sl[0].Metrics {
			for x, _ := range sl {
				acc += sl[x].Metrics[i].Value
			}
			sl[0].Metrics[i].Value = acc / float64(len(sl))
		}
		return sl[:1]
	}
	return
}

func NewScaleSeriesPP(scale float64) (pp PipelineProcessor) {
	pp = func(sl SeriesList) SeriesList {
		for seriesIndex, _ := range sl {
			for metricIndex, _ := range sl[seriesIndex].Metrics {
				metric := &sl[seriesIndex].Metrics[metricIndex]
				metric.Value *= scale
			}
		}
		return sl
	}
	return
}
