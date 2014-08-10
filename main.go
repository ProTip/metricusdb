package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/cors"
	"github.com/protip/metricusdb/metricusdb"
	"net/http"
	"strings"
)

func main() {
	m := martini.Classic()
	m.Use(cors.Allow(&cors.Options{
		AllowOrigins: []string{"*"}}))
	m.Get("/render", GetRender)
	m.Post("/render", GetRender)
	m.Get("/metrics/find/", GetMetricQuery)
	http.ListenAndServe(":8080", m)
}

func GetRender(req *http.Request, p martini.Params) string {
	v := req.URL.Query()
	from := v.Get("from")
	to := v.Get("to")
	fmt.Println(from, to)
	return "true"
}

func GetMetricQuery(req *http.Request, p martini.Params) string {
	v := req.URL.Query()
	query := v.Get("query")
	fmt.Println(query)
	dimensions := metricusdb.TargetToDimensions(query)
	fmt.Println(dimensions)
	se := &metricusdb.RiakSE{Address: "10.1.1.8:8087"}
	metrics, _ := se.ListMetrics(dimensions)
	fmt.Println(metrics)
	tree := metricusdb.MetricsToTree(query, metrics, len(strings.Split(query, "."))-1)
	bSON, _ := json.Marshal(tree)
	return string(bSON)
}
