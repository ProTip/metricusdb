package metricusdb

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

type GraphiteParser struct {
	lex         *lexer
	Targets     []string
	Funcs       []GraphiteFunction
	currentFunc int
}

type GraphiteFunction struct {
	Func string
	Args []interface{}
}

type GraphitePipeline struct {
	from time.Time
	to   time.Time
	*Pipeline
}

var GraphiteFuncToPipelineFunc = map[string]string{
	"averageSeries": "GNewAverageSeries",
	"scale":         "GNewScale"}

func (parser *GraphiteParser) Parse() {
	b := bytes.Buffer{}
	for {
		i := parser.lex.nextItem()
		switch i.typ {
		case itemEOF, itemError:
			break
		case itemTarget:
			parser.Targets = append(parser.Targets, i.val)
		case itemFunction:
			parser.Funcs = append(parser.Funcs, GraphiteFunction{
				Func: i.val,
				Args: make([]interface{}, 0, 0)})
		case itemLeftParen:
			parser.currentFunc += 1
		case itemRightParen:
			parser.currentFunc -= 1
		case itemString, itemNumber:
			args := &parser.Funcs[parser.currentFunc].Args
			*args = append(*args, i.val)
		}
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
}

func (parser *GraphiteParser) BuildPipeLine() *Pipeline {
	pp := NewPipeLine()
	gp := &GraphitePipeline{
		Pipeline: pp}
	for i := len(parser.Funcs) - 1; i >= 0; i-- {
		methodVal := reflect.ValueOf(gp).MethodByName(GraphiteFuncToPipelineFunc[parser.Funcs[i].Func])
		methodIface := methodVal.Interface()
		method := methodIface.(func(...interface{}) error)
		err := method(parser.Funcs[i].Args...)
		if err != nil {
			panic(err.Error())
		}
	}
	return pp
}

func (gp *GraphitePipeline) GNewAverageSeries(_ ...interface{}) error {
	fmt.Println("Adding AverageSeries PP")
	gp.AddProcessor(NewAverageSeriesPP())
	return nil
}

func (gp *GraphitePipeline) GNewScale(args ...interface{}) error {
	scale, err := strconv.ParseFloat(args[0].(string), 64)
	fmt.Println("Scale is : ", scale)
	if err != nil {
		return err
	}
	fmt.Println("Adding Scale PP")
	gp.AddProcessor(NewScaleSeriesPP(scale))
	return nil
}
