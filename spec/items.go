package spec

import (
	"encoding/json"

	"github.com/casualjim/go-swagger/swag"
)

type simpleSchema struct {
	Type             string      `json:"type,omitempty"`
	Format           string      `json:"format,omitempty"`
	Items            *Items      `json:"items,omitempty"`
	CollectionFormat string      `json:"collectionFormat,omitempty"`
	Default          interface{} `json:"default,omitempty"`
}

func (s *simpleSchema) TypeName() string {
	if s.Format != "" {
		return s.Format
	}
	return s.Type
}

func (s *simpleSchema) ItemsTypeName() string {
	if s.Items == nil {
		return ""
	}
	return s.Items.TypeName()
}

type commonValidations struct {
	Maximum          *float64      `json:"maximum,omitempty"`
	ExclusiveMaximum bool          `json:"exclusiveMaximum,omitempty"`
	Minimum          *float64      `json:"minimum,omitempty"`
	ExclusiveMinimum bool          `json:"exclusiveMinimum,omitempty"`
	MaxLength        *int64        `json:"maxLength,omitempty"`
	MinLength        *int64        `json:"minLength,omitempty"`
	Pattern          string        `json:"pattern,omitempty"`
	MaxItems         *int64        `json:"maxItems,omitempty"`
	MinItems         *int64        `json:"minItems,omitempty"`
	UniqueItems      bool          `json:"uniqueItems,omitempty"`
	MultipleOf       *float64      `json:"multipleOf,omitempty"`
	Enum             []interface{} `json:"enum,omitempty"`
}

// Items a limited subset of JSON-Schema's items object.
// It is used by parameter definitions that are not located in "body".
//
// For more information: http://goo.gl/8us55a#items-object-
type Items struct {
	refable
	commonValidations
	simpleSchema
}

// NewItems creates a new instance of items
func NewItems() *Items {
	return &Items{}
}

// Typed a fluent builder method for the type of item
func (i *Items) Typed(tpe, format string) *Items {
	i.Type = tpe
	i.Format = format
	return i
}

// CollectionOf a fluent builder method for an array item
func (i *Items) CollectionOf(items *Items, format string) *Items {
	i.Type = "array"
	i.Items = items
	i.CollectionFormat = format
	return i
}

// WithDefault sets the default value on this item
func (i *Items) WithDefault(defaultValue interface{}) *Items {
	i.Default = defaultValue
	return i
}

// WithMaxLength sets a max length value
func (i *Items) WithMaxLength(max int64) *Items {
	i.MaxLength = &max
	return i
}

// WithMinLength sets a min length value
func (i *Items) WithMinLength(min int64) *Items {
	i.MinLength = &min
	return i
}

// WithPattern sets a pattern value
func (i *Items) WithPattern(pattern string) *Items {
	i.Pattern = pattern
	return i
}

// WithMultipleOf sets a multiple of value
func (i *Items) WithMultipleOf(number float64) *Items {
	i.MultipleOf = &number
	return i
}

// WithMaximum sets a maximum number value
func (i *Items) WithMaximum(max float64, exclusive bool) *Items {
	i.Maximum = &max
	i.ExclusiveMaximum = exclusive
	return i
}

// WithMinimum sets a minimum number value
func (i *Items) WithMinimum(min float64, exclusive bool) *Items {
	i.Minimum = &min
	i.ExclusiveMinimum = exclusive
	return i
}

// WithEnum sets a the enum values (replace)
func (i *Items) WithEnum(values ...interface{}) *Items {
	i.Enum = append([]interface{}{}, values...)
	return i
}

// WithMaxItems sets the max items
func (i *Items) WithMaxItems(size int64) *Items {
	i.MaxItems = &size
	return i
}

// WithMinItems sets the min items
func (i *Items) WithMinItems(size int64) *Items {
	i.MinItems = &size
	return i
}

// UniqueValues dictates that this array can only have unique items
func (i *Items) UniqueValues() *Items {
	i.UniqueItems = true
	return i
}

// AllowDuplicates this array can have duplicates
func (i *Items) AllowDuplicates() *Items {
	i.UniqueItems = false
	return i
}

// UnmarshalJSON hydrates this items instance with the data from JSON
func (i *Items) UnmarshalJSON(data []byte) error {
	var validations commonValidations
	if err := json.Unmarshal(data, &validations); err != nil {
		return err
	}
	var ref refable
	if err := json.Unmarshal(data, &ref); err != nil {
		return err
	}
	var simpleSchema simpleSchema
	if err := json.Unmarshal(data, &simpleSchema); err != nil {
		return err
	}
	i.refable = ref
	i.commonValidations = validations
	i.simpleSchema = simpleSchema
	return nil
}

// MarshalJSON converts this items object to JSON
func (i Items) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(i.commonValidations)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(i.simpleSchema)
	if err != nil {
		return nil, err
	}
	b3, err := json.Marshal(i.refable)
	if err != nil {
		return nil, err
	}
	return swag.ConcatJSON(b3, b1, b2), nil
}
