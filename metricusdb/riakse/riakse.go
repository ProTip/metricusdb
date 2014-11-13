package riakse

import (
	"fmt"
	"github.com/protip/metricusdb/metricusdb"
	"github.com/tpjg/goriakpbc"
	"strings"
)

type RiakSE struct {
	Address string
}

func (se *RiakSE) ListStreams(dimensions [][]string) ([]metricusdb.Stream, error) {
	terms := make([]string, 0, len(dimensions))
	for _, e := range dimensions {
		terms = append(terms, fmt.Sprintf("%s:%s", e[0], e[1]))
	}
	query := strings.Join(terms, " AND ")
	fmt.Println("Querying metrics with query: ", query)
	c := riak.NewClient(se.Address)
	streams := make([]metricusdb.Stream, 0, 0)
	if docs, _, _, err := c.Search(&riak.Search{Q: "*", Filter: query, Index: "famous"}); err == nil {
		stream := metricusdb.Stream{
			Dimensions: make(map[string]string),
		}
		for _, doc := range docs {
			for k, v := range doc {
				if (len(k) >= 3 && k[:3] == "_yz") || k == "score" {
					continue
				}
				stream.Dimensions[k] = string(v)
			}
			stream.Name = string(doc["_yz_rk"])
			streams = append(streams, stream)
		}
	}
	return streams, nil
}

func (se *RiakSE) StoreMetric(metricusdb.Metric, metricusdb.Stream) {

}
