package tracing

import (
	"io/ioutil"
	"testing"
)

func TestTracer(t *testing.T) {
	span := StartSpan("get.data")
	defer span.Finish()

	child := StartSpan("read.file", WithParent(span.Context()))
	child.SetTag(ResourceName, "test.json")

	_, err := ioutil.ReadFile("./test.json")
	child.Finish(WithError(err))
	if err != nil {
		t.Fatal(err)
	}
}
