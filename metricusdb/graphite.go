package metricusdb

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type GraphiteNode struct {
	Leaf          int      `json:"leaf"`
	Context       struct{} `json:"context"`
	Text          string   `json:"text"`
	Expandable    int      `json:"expandable"`
	Id            string   `json:"id"`
	AllowChildren int      `json:"allowChildren"`
}

// Takes a graphite target and returns a list of dimensions used to find the
// metric stream.
func TargetToDimensions(target string) [][]string {
	splitTarget := strings.Split(target, ".")
	dimensions := make([][]string, len(splitTarget))
	for i, s := range splitTarget {
		dimensions[i] = make([]string, 2, 2)
		dimensions[i][0] = fmt.Sprintf("nodes.%d_s", i)
		dimensions[i][1] = s
	}
	return dimensions
}

// Takes a list of metrics and the query used to find them, then returns
// a list of nodes in the format expected by graphic clients.
func StreamsToTree(query string, streams []Stream, depth int) []*GraphiteNode {
	tree := make(map[string]*GraphiteNode)
	nodes := make([]*GraphiteNode, 0)
	splitQuery := strings.Split(query, ".")
	for _, m := range streams {
		splitMetric := strings.Split(m.Name, ".")
		idParts := splitQuery[:depth]
		idParts = append(splitMetric[depth : depth+1])
		id := strings.Join(idParts, ".")
		if _, ok := tree[id]; !ok {
			node := &GraphiteNode{
				Leaf:          1,
				Text:          splitMetric[depth],
				Id:            id,
				Expandable:    0,
				AllowChildren: 0}
			if len(splitMetric) > depth+1 {
				node.Leaf = 0
				node.Expandable = 1
				node.AllowChildren = 1
			}
			tree[id] = node
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func ConvertGraphiteTime(from string) (time.Time, error) {
	if duration, err := time.ParseDuration(from); err == nil {
		return time.Now().Add(duration), nil
	} else if t, err := time.Parse(time.UnixDate, from); err == nil {
		return t, nil
	} else {
		return time.Time{}, errors.New("Unable to parse time!")
	}
}
