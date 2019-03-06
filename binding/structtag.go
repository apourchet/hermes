package binding

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

type StructTagBinding struct{}

type ValueBinder func(*http.Request, interface{}, string, string) error
type ValueApplier func(req *http.Request, directivevalue string, fieldvalue string) error

var StructTagBinds = map[string]ValueBinder{
	"header": BindHeader,
	"query":  BindQuery,
	"path":   BindPath,
	"cookie": BindCookie,
}

var StructTagApps = map[string]ValueApplier{
	"header": ApplyHeader,
	"query":  ApplyQuery,
	"path":   ApplyPath,
	"cookie": ApplyCookie,
}

// Given a struct definition:
// type Request struct {
//		Token string `hermes:"header=Authorization"`
//		Limit int `hermes:"query=limit"`
//		Resource string `hermes:"param=resource"`
// }
// The Authorization header will be parsed into the field Token of the
// request struct
// The Limit field will come from the query string
// The Resource field will come from the resource value of the path
func (b StructTagBinding) Bind(req *http.Request, obj interface{}) error {
	st, valid := DerefStruct(obj)
	if !valid {
		return nil
	}

	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		if alias, ok := field.Tag.Lookup("hermes"); ok && alias != "" {
			split := strings.Split(alias, ",")
			for _, directive := range split {
				if directive == "" {
					continue
				}
				if err := b.BindDirective(req, obj, field.Name, directive); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (b StructTagBinding) Apply(obj interface{}, req *http.Request) error {
	if req == nil || obj == nil {
		return nil
	}

	v, valid := Deref(obj)
	if !valid || v.Kind() != reflect.Struct {
		return nil
	}

	fields, err := FieldMap(obj)
	if err != nil {
		return fmt.Errorf("Failed to apply header binding: %v", err)
	}

	st, valid := DerefStruct(obj)
	if !valid {
		return nil
	}

	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		lowername := strings.ToLower(field.Name)
		if alias, ok := field.Tag.Lookup("hermes"); ok && alias != "" {
			split := strings.Split(alias, ",")
			for _, directive := range split {
				if directive == "" {
					continue
				}

				// Get the string value for this field of the input object
				fieldval, found := fields[lowername]
				if !found {
					continue
				}

				// Apply the directive
				if err := b.ApplyDirective(req, directive, fieldval); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (b StructTagBinding) BindDirective(req *http.Request, obj interface{}, fieldname string, directive string) error {
	split := strings.Split(directive, "=")
	if len(split) != 2 {
		return fmt.Errorf("Malformed struct tag: %v", directive)
	}
	tagkey, tagval := split[0], split[1]
	operation, found := StructTagBinds[tagkey]
	if !found {
		return fmt.Errorf("Failed to resolve struct tag operation: %v", tagkey)
	}

	err := operation(req, obj, tagval, fieldname)
	return err
}

func (b StructTagBinding) ApplyDirective(req *http.Request, directive string, fieldvalue string) error {
	split := strings.Split(directive, "=")
	if len(split) != 2 {
		return fmt.Errorf("Malformed struct tag: %v", directive)
	}

	tagkey, tagval := split[0], split[1]
	operation, found := StructTagApps[tagkey]
	if !found {
		return fmt.Errorf("Failed to resolve struct tag operation: %v", tagkey)
	}

	err := operation(req, tagval, fieldvalue)
	return err
}
