package binding

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/fatih/structs"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

// Binds the header <headername> to the field <fieldname> of obj
func BindHeader(ctx *gin.Context, obj interface{}, headername string, fieldname string) error {
	headerval := ctx.Request.Header.Get(headername)
	if headerval == "" {
		return nil
	}
	if err := SetField(obj, fieldname, headerval); err != nil {
		return fmt.Errorf("Failed to set header binding %s: %v", headername, err)
	}
	return nil
}

func BindQuery(ctx *gin.Context, obj interface{}, queryparam string, fieldname string) error {
	vals, found := ctx.Request.URL.Query()[queryparam]
	if !found || len(vals) == 0 {
		return nil
	}

	if len(vals) > 1 {
		err := fmt.Errorf("Query parameter had multiple values; which is unsupported.")
		if (QueryFlags | IGNORE_MULTIPLE_QUERYVALS) == 0 {
			glog.Warningf("%v", err)
		} else {
			return err
		}
	}
	queryval, err := url.QueryUnescape(vals[0])
	if err != nil {
		return fmt.Errorf("Failed to unescape query value %s: %v", vals[0], err)
	}

	if err := SetField(obj, fieldname, queryval); err != nil {
		return fmt.Errorf("Failed to set url query binding %s: %v", queryparam, err)
	}
	return nil
}

func BindPath(ctx *gin.Context, obj interface{}, pathparam string, fieldname string) error {
	pathval := ctx.Param(pathparam)
	if pathval == "" {
		return nil
	}

	pathval, err := url.QueryUnescape(pathval)
	if err != nil {
		return fmt.Errorf("Failed to unescape path value %s: %v", pathval, err)
	}

	if err := SetField(obj, fieldname, pathval); err != nil {
		return fmt.Errorf("Failed to set url parameter binding %s: %v", pathparam, err)
	}
	return nil
}

func ApplyHeader(req *http.Request, headername string, fieldvalue string) error {
	req.Header.Set(headername, fieldvalue)
	return nil
}

func ApplyQuery(req *http.Request, queryname string, fieldvalue string) error {
	if req.URL.RawQuery != "" {
		req.URL.RawQuery += "&"
	}
	req.URL.RawQuery += fmt.Sprintf("%s=%s", url.QueryEscape(queryname), url.QueryEscape(fieldvalue))
	return nil
}

func ApplyPath(req *http.Request, paramname string, fieldvalue string) error {
	req.URL.Path = strings.Replace(req.URL.Path, ":"+paramname, url.QueryEscape(fieldvalue), -1)
	return nil
}

func Deref(obj interface{}) (reflect.Value, bool) {
	v := reflect.ValueOf(obj)
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() { // If the chain ends in a nil, skip this
			return v, false
		}
		v = v.Elem()
	}
	return v, true
}

func DerefStruct(obj interface{}) (reflect.Type, bool) {
	st := reflect.TypeOf(obj)
	for st.Kind() == reflect.Ptr || st.Kind() == reflect.Interface {
		st = st.Elem()
	}

	return st, st.Kind() == reflect.Struct
}

func FieldMap(obj interface{}) (map[string]string, error) {
	rawFields := structs.Map(obj)
	fields := map[string]string{}
	for name, value := range rawFields {
		skip, value, err := Stringify(value)
		if err != nil {
			return fields, fmt.Errorf("Failed to construct path: %v", err)
		} else if !skip {
			fields[strings.ToLower(name)] = value
		}
	}
	return fields, nil
}

func Stringify(val interface{}) (bool, string, error) {
	v, valid := Deref(val)
	if !valid {
		return true, "", nil
	}

	switch v.Kind() {
	case reflect.Bool:
		return false, fmt.Sprintf("%v", v.Bool()), nil
	case reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64:
		return false, fmt.Sprintf("%v", v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint32, reflect.Uint64:
		return false, fmt.Sprintf("%v", v.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return false, fmt.Sprintf("%v", v.Float()), nil
	case reflect.String:
		return false, fmt.Sprintf("%v", v.String()), nil
	case reflect.Slice, reflect.Map:
		content, err := json.Marshal(val)
		if err != nil {
			return false, "", fmt.Errorf("Failed to stringify value into url: %v", err)
		}
		return false, string(content), nil
	}
	return false, "", fmt.Errorf("Unsupported type: %T", val)
}

// Sets the field of the object using a string that
// was retrieved from the URI of the request
func SetField(obj interface{}, fieldname, value string) error {
	v, valid := Deref(obj)
	if !valid || v.Kind() != reflect.Struct {
		return nil
	}

	field := v.FieldByNameFunc(func(a string) bool { return strings.ToLower(a) == fieldname })

	if !field.IsValid() {
		return fmt.Errorf("Field not found when binding: %s", fieldname)
	}

	val, err := ParseString(field.Type(), value)
	if err != nil {
		return fmt.Errorf("Failed to parse value: %v", err)
	}

	field.Set(val)
	return nil
}

func ParseString(t reflect.Type, value string) (reflect.Value, error) {
	switch t.Kind() {
	case reflect.Ptr:
		subval, err := ParseString(t.Elem(), value)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse url parameter/query to %s: %v", t, err)
		}
		val := reflect.New(t.Elem())
		val.Elem().Set(subval)
		return val, nil
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse url parameter/query to %T: %v", b, err)
		}
		return reflect.ValueOf(b), nil
	case reflect.String:
		return reflect.ValueOf(value), nil
	case reflect.Int:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse url parameter/query to %T: %v", i, err)
		}
		return reflect.ValueOf(int(i)), nil
	case reflect.Int8:
		i, err := strconv.ParseInt(value, 10, 8)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse url parameter/query to %T: %v", i, err)
		}
		return reflect.ValueOf(int8(i)), nil
	case reflect.Int32:
		i, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse url parameter/query to %T: %v", i, err)
		}
		return reflect.ValueOf(int32(i)), nil
	case reflect.Int64:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse url parameter/query to %T: %v", i, err)
		}
		return reflect.ValueOf(int64(i)), nil
	case reflect.Uint:
		i, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse url parameter/query to %T: %v", i, err)
		}
		return reflect.ValueOf(uint(i)), nil
	case reflect.Uint8:
		i, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse url parameter/query to %T: %v", i, err)
		}
		return reflect.ValueOf(uint8(i)), nil
	case reflect.Uint32:
		i, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse url parameter/query to %T: %v", i, err)
		}
		return reflect.ValueOf(uint32(i)), nil
	case reflect.Uint64:
		i, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse url parameter/query to %T: %v", i, err)
		}
		return reflect.ValueOf(uint64(i)), nil
	case reflect.Float32:
		f, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse url parameter/query to %T: %v", f, err)
		}
		return reflect.ValueOf(float32(f)), nil
	case reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse url parameter/query to %T: %v", f, err)
		}
		return reflect.ValueOf(float64(f)), nil
	case reflect.Slice:
		var s []interface{}
		err := json.Unmarshal([]byte(value), &s)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse url parameter/query to %T: %v", s, err)
		}
		return reflect.ValueOf(s), nil
	case reflect.Map:
		var m map[string]interface{}
		err := json.Unmarshal([]byte(value), &m)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("Failed to parse url parameter/query to %T: %v", m, err)
		}
		return reflect.ValueOf(m), nil
	}
	return reflect.ValueOf(nil), fmt.Errorf("Unsupported type: %v", t)
}
