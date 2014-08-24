package metricusdb

import (
	"bytes"
	"fmt"
	"github.com/kr/pretty"
	"testing"
	"text/tabwriter"
)

func TestLex(t *testing.T) {
	ob := bytes.Buffer{}
	tr := tabwriter.NewWriter(&ob, 5, 5, 1, ' ', 0)
	//l := lex("", `aliasSub(aliasByNode(natasha.production.stack-servers.18.*.processor._total.*_processor_time,4),'(iis-.*?)_.*','\1_%cpu_time')`, "", "")
	l := lex("", `scale(averageSeries(natasha.production.stack-servers.18.*.processor._total.*_processor_time),5)`, "", "")
	b := bytes.Buffer{}
	for {
		i := l.nextItem()
		fmt.Println(i.val, "[=]", itemLookup[i.typ])

		val := i.val
		if i.typ == itemEOF {
			val = "EOF"
		}
		b.WriteString(val + "\t[=]\t" + itemLookup[i.typ] + "\n")
		if i.typ == itemEOF || i.typ == itemError {
			break
		}
	}
	tr.Write(b.Bytes())
	tr.Flush()
	fmt.Print(ob.String())
}

func TestParse(t *testing.T) {
	//l := lex("", `aliasSub(aliasByNode(natasha.production.stack-servers.18.*.processor._total.*_processor_time,4),'(iis-.*?)_.*','\1_%cpu_time')`, "", "")
	l := lex("", `scale(averageSeries(natasha.production.stack-servers.18.*.processor._total.*_processor_time),5)`, "", "")
	p := &GraphiteParser{
		lex:         l,
		Funcs:       make([]GraphiteFunction, 0, 0),
		currentFunc: -1}
	p.Parse()
	pretty.Println(p)
	pipeline := p.BuildPipeLine()
	go pipeline.Run()
	pipeline.Input <- SeriesList{
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
	result := <-pipeline.Output
	fmt.Println(result)
}
