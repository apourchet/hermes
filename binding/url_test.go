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

func TestSetFieldString(t *testing.T) {
	obj := struct {
		Field1 string
	}{}
	err := SetField(&obj, "field1", "test")
	assert.Nil(t, err)
	assert.Equal(t, "test", obj.Field1)
}

func TestSetFieldBool(t *testing.T) {
	obj := struct {
		Field1 bool
	}{}
	err := SetField(&obj, "field1", "true")
	assert.Nil(t, err)
	assert.Equal(t, true, obj.Field1)
}

func TestSetFieldInt(t *testing.T) {
	obj := struct {
		A int
		B int8
		C int32
		D int64
	}{}
	err := SetField(&obj, "a", "-1")
	assert.Nil(t, err)
	assert.Equal(t, -1, obj.A)

	err = SetField(&obj, "b", "100")
	assert.Nil(t, err)
	assert.Equal(t, int8(100), obj.B)

	err = SetField(&obj, "c", "123123")
	assert.Nil(t, err)
	assert.Equal(t, int32(123123), obj.C)

	err = SetField(&obj, "d", "99999999")
	assert.Nil(t, err)
	assert.Equal(t, int64(99999999), obj.D)
}

func TestSetFieldUInt(t *testing.T) {
	obj := struct {
		A uint
		B uint8
		C uint32
		D uint64
	}{}
	err := SetField(&obj, "a", "3")
	assert.Nil(t, err)
	assert.Equal(t, uint(3), obj.A)

	err = SetField(&obj, "b", "120")
	assert.Nil(t, err)
	assert.Equal(t, uint8(120), obj.B)

	err = SetField(&obj, "c", "123123")
	assert.Nil(t, err)
	assert.Equal(t, uint32(123123), obj.C)

	err = SetField(&obj, "d", "999999999")
	assert.Nil(t, err)
	assert.Equal(t, uint64(999999999), obj.D)
}

func TestSetFieldFloat(t *testing.T) {
	obj := struct {
		A float32
		B float64
	}{}
	err := SetField(&obj, "a", "0.1")
	assert.Nil(t, err)
	assert.Equal(t, float32(0.1), obj.A)

	err = SetField(&obj, "b", "255.2")
	assert.Nil(t, err)
	assert.Equal(t, float64(255.2), obj.B)
}

func TestSetFieldSlice(t *testing.T) {
	obj := struct {
		A []interface{}
	}{}
	err := SetField(&obj, "a", `["a","b"]`)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(obj.A))
}

func TestSetFieldMap(t *testing.T) {
	obj := struct {
		A map[string]interface{}
	}{}
	err := SetField(&obj, "a", `{"a":1}`)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(obj.A))
}
