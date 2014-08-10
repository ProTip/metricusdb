package metricusdb

import (
	"fmt"
	"github.com/tpjg/goriakpbc"
	"strings"
)

type RiakSE struct {
	Address string
}

func (se *RiakSE) ListMetrics(dimensions [][]string) ([]string, error) {
	terms := make([]string, 0, len(dimensions))
	for _, e := range dimensions {
		terms = append(terms, fmt.Sprintf("%s:%s", e[0], e[1]))
	}
	query := strings.Join(terms, " AND ")
	fmt.Println("Querying metrics with query: ", query)
	c := riak.NewClient(se.Address)
	metrics := make([]string, 0, 0)
	if docs, _, _, err := c.Search(&riak.Search{Q: query, Index: "famous"}); err == nil {
		for _, doc := range docs {
			fmt.Println(doc)
			metrics = append(metrics, string(doc["_yz_rk"]))
		}
	}
	return metrics, nil
}
