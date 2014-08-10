package metricusdb

import (
	"bytes"
	"fmt"
	"testing"
	"text/tabwriter"
)

func TestParse(t *testing.T) {
	ob := bytes.Buffer{}
	tr := tabwriter.NewWriter(&ob, 5, 5, 1, ' ', 0)
	l := lex("", `aliasSub(aliasByNode(natasha.production.stack-servers.18.*.processor._total.*_processor_time,4),'(iis-.*?)_.*','\1_%cpu_time')`, "", "")
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
