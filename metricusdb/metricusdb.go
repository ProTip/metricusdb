package metricusdb

import (
	"time"
)

type Query struct {
	Dimensions map[string]interface{}
}

type Stream struct {
	Dimensions map[string]interface{}
	Name       string
}

type StoreQuery struct {
	*Query
}

type RetrieveQuery struct {
	*Query
	From time.Time
	To   time.Time
}

type QueryResult struct {
	Query *Query
}

type DbEngine interface {
	Store(*StoreQuery) QueryResult
	Retrieve(*RetrieveQuery) QueryResult
	LookupStreams(*Query) []*Stream
}

type StoreQueryParser interface {
	Parse(string) StoreQuery
}

type RetrieveQueryParser interface {
	Parse(string) RetrieveQuery
}
