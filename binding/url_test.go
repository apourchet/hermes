package binding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var binding1 = &URLBinding{[]string{"param1", "param2"}, []string{"query1", "query2"}}

func TestTransformURLParam(t *testing.T) {
	input := struct{ Param1 string }{"myname"}
	path, err := binding1.TransformURL("http://something.com/api/v1/test/:param1/something", input)
	assert.Nil(t, err)
	assert.Equal(t, "http://something.com/api/v1/test/myname/something", path)
}

func TestTransformURLParams(t *testing.T) {
	input := struct {
		Param1 string
		Param2 int
	}{"myname", 2}

	path, err := binding1.TransformURL("http://something.com/api/v1/:param1/:param2", input)
	assert.Nil(t, err)
	assert.Equal(t, "http://something.com/api/v1/myname/2", path)
}

func TestTransformURLQuery(t *testing.T) {
	input := struct{ Query1 string }{"myname"}
	path, err := binding1.TransformURL("http://something.com/api/v1/test", input)
	assert.Nil(t, err)
	assert.Equal(t, "http://something.com/api/v1/test?query1=myname", path)
}

func TestTransformURLQueries(t *testing.T) {
	input := struct {
		Query1 string
		Query2 int
	}{"myname", 2}
	path, err := binding1.TransformURL("http://something.com/api/v1/test", input)
	assert.Nil(t, err)
	assert.Equal(t, "http://something.com/api/v1/test?query1=myname&query2=2", path)
}

func TestTransformURLMix(t *testing.T) {
	input := struct {
		Param1 string
		Query2 int
	}{"myname", 2}
	path, err := binding1.TransformURL("http://something.com/api/v1/:param1", input)
	assert.Nil(t, err)
	assert.Equal(t, "http://something.com/api/v1/myname?query2=2", path)
}
