package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/cors"
	"github.com/protip/metricusdb/metricusdb"
	"github.com/protip/metricusdb/metricusdb/riakse"
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

func GetRender(req *http.Request, p martini.Params, w http.ResponseWriter) (int, string) {
	v := req.URL.Query()
	w.Header().Set("Content-Type", "application/json")
	from := v.Get("from")
	to := v.Get("to")
	fmt.Println(from, to)
	return 500, "Not implemented"
}

func GetMetricQuery(req *http.Request, p martini.Params, w http.ResponseWriter) string {
	v := req.URL.Query()
	w.Header().Set("Content-Type", "application/json")
	query := v.Get("query")
	fmt.Println(query)
	dimensions := metricusdb.TargetToDimensions(query)
	fmt.Println(dimensions)
	se := &riakse.RiakSE{Address: "192.168.1.66:8087", StreamIndex: "famous"}
	streams, _ := se.ListStreams(dimensions)
	fmt.Println(streams)
	tree := metricusdb.StreamsToTree(query, streams, len(strings.Split(query, "."))-1)
	bSON, _ := json.Marshal(tree)
	return string(bSON)
}
