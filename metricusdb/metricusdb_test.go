package metricusdb

import (
	"crypto/md5"
	"fmt"
	_ "github.com/kr/pretty"
	"testing"
)

func TestStreamIdentifier(t *testing.T) {
	stream := Stream{
		Name: "CPU",
		Dimensions: []Dimension{
			{Name: "Host", Value: "Local"},
		},
	}
	fmt.Println(stream.Identifier())
}

func BenchmarkStreamIdentifier(b *testing.B) {
	stream := Stream{
		Namespace: "modusOps", Dimensions: Dimensions{
			{Name: "Host", Value: "local"},
			{Name: "Region", Value: "ap-southeast-2"},
			{Name: "VPC", Value: "staging"},
			{Name: "instanceID", Value: "i-2323234"},
		}}
	for i := 0; i < b.N; i++ {
		_ = stream.Identifier()
	}
}

func BenchmarkMd5(b *testing.B) {
	summer := md5.New()
	target := []byte("Hello world")
	for i := 0; i < b.N; i++ {
		summer.Write(target)
		_ = summer.Sum(nil)
	}
}
