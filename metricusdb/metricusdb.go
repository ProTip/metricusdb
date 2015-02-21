package metricusdb

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	_ "fmt"
	"sort"
	"time"
)

type Dimension struct {
	Name  string
	Value string
}

type Dimensions []Dimension

type ByName []Dimension

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i].Name < a[j].Name }

type Query struct {
	Dimensions
}

type Stream struct {
	Dimensions
	Name       string
	Namespace  string
	identifier string
}

func (s *Stream) Identifier() string {
	if s.identifier == "" {
		var buffer bytes.Buffer
		sort.Sort(ByName(s.Dimensions))
		buffer.WriteString(s.Namespace)
		for _, dimension := range s.Dimensions {
			buffer.WriteString(dimension.Name)
			buffer.WriteString(dimension.Value)
		}
		buffer.WriteString(s.Name)
		sum := md5.Sum(buffer.Bytes())
		s.identifier = hex.EncodeToString(sum[:])
	}
	return s.identifier
}

type InsertQuery struct {
	Stream
	Metric
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
	Store(*InsertQuery) QueryResult
	Retrieve(*RetrieveQuery) QueryResult
	LookupStreams(*Query) []*Stream
}

type InsertQueryParser interface {
	Parse(string) InsertQuery
}

type RetrieveQueryParser interface {
	Parse(string) RetrieveQuery
}
