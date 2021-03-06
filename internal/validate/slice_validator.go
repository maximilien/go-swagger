package validate

import (
	"fmt"
	"reflect"

	"github.com/casualjim/go-swagger/errors"
	"github.com/casualjim/go-swagger/httpkit/validate"
	"github.com/casualjim/go-swagger/spec"
	"github.com/casualjim/go-swagger/strfmt"
)

type schemaSliceValidator struct {
	Path            string
	In              string
	MaxItems        *int64
	MinItems        *int64
	UniqueItems     bool
	AdditionalItems *spec.SchemaOrBool
	Items           *spec.SchemaOrArray
	Root            interface{}
	KnownFormats    strfmt.Registry
}

func (s *schemaSliceValidator) SetPath(path string) {
	s.Path = path
}

func (s *schemaSliceValidator) Applies(source interface{}, kind reflect.Kind) bool {
	_, ok := source.(*spec.Schema)
	r := ok && kind == reflect.Slice
	return r
}

func (s *schemaSliceValidator) Validate(data interface{}) *Result {
	result := new(Result)
	if data == nil {
		return result
	}
	val := reflect.ValueOf(data)
	size := val.Len()

	if s.Items != nil && s.Items.Schema != nil {
		validator := NewSchemaValidator(s.Items.Schema, s.Root, s.Path, s.KnownFormats)
		for i := 0; i < size; i++ {
			validator.SetPath(fmt.Sprintf("%s.%d", s.Path, i))
			value := val.Index(i)
			result.Merge(validator.Validate(value.Interface()))
		}
	}

	itemsSize := int64(0)
	if s.Items != nil && len(s.Items.Schemas) > 0 {
		itemsSize = int64(len(s.Items.Schemas))
		for i := int64(0); i < itemsSize; i++ {
			validator := NewSchemaValidator(&s.Items.Schemas[i], s.Root, fmt.Sprintf("%s.%d", s.Path, i), s.KnownFormats)
			result.Merge(validator.Validate(val.Index(int(i)).Interface()))
		}

	}
	if s.AdditionalItems != nil && itemsSize < int64(size) {
		if s.Items != nil && len(s.Items.Schemas) > 0 && !s.AdditionalItems.Allows {
			result.AddErrors(errors.New(422, "array doesn't allow for additional items"))
		}
		if s.AdditionalItems.Schema != nil {
			for i := itemsSize; i < (int64(size)-itemsSize)+1; i++ {
				validator := NewSchemaValidator(s.AdditionalItems.Schema, s.Root, fmt.Sprintf("%s.%d", s.Path, i), s.KnownFormats)
				result.Merge(validator.Validate(val.Index(int(i)).Interface()))
			}
		}
	}

	if s.MinItems != nil {
		if err := validate.MinItems(s.Path, s.In, int64(size), *s.MinItems); err != nil {
			result.AddErrors(err)
		}
	}
	if s.MaxItems != nil {
		if err := validate.MaxItems(s.Path, s.In, int64(size), *s.MaxItems); err != nil {
			result.AddErrors(err)
		}
	}
	if s.UniqueItems {
		if err := validate.UniqueItems(s.Path, s.In, val.Interface()); err != nil {
			result.AddErrors(err)
		}
	}
	result.Inc()
	return result
}
