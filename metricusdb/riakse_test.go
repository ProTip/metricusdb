package metricusdb

import (
	"fmt"
	"testing"
)

func TestHello(t *testing.T) {
	fmt.Println("Hello world!")
}

func TestListMetrics(t *testing.T) {
	se := &RiakSE{Address: "10.1.1.8:8087"}
	metrics, _ := se.ListMetrics([][]string{
		{"node0_s", "natasha"},
	})
	fmt.Println(metrics)
}

func TestTargetToDimensions(t *testing.T) {
	dimensions := TargetToDimensions("natasha.production.stack-servers.5")
	fmt.Println(dimensions)
}

func TestMetricsToTree(t *testing.T) {
	metrics := []string{
		`natasha.production.stack.5.cpu_load`,
		`natasha.production.stack.6.cpu_load`,
		`natasha.production.rds.5.load`,
		`natasha.development.stack.1.cpu_load`,
	}
	tree := MetricsToTree("natasha.*.*", metrics, 4)
	for _, node := range tree {
		fmt.Printf("%+v\n", *node)
	}
}
