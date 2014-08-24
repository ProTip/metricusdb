package metricusdb

type GraphiteParser struct {
	lex         *lexer
	Func        []GraphiteFunction
	currentFunc int
}

type GraphiteFunction struct {
	Func string
	Args []interface{}
}

func (parser *GraphiteParser) Parse() {

}
