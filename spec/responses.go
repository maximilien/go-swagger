package spec

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/casualjim/go-swagger/swag"
)

// Responses is a container for the expected responses of an operation.
// The container maps a HTTP response code to the expected response.
// It is not expected from the documentation to necessarily cover all possible HTTP response codes,
// since they may not be known in advance. However, it is expected from the documentation to cover
// a successful operation response and any known errors.
//
// The `default` can be used a default response object for all HTTP codes that are not covered
// individually by the specification.
//
// The `Responses Object` MUST contain at least one response code, and it SHOULD be the response
// for a successful operation call.
//
// For more information: http://goo.gl/8us55a#responsesObject
type Responses struct {
	vendorExtensible
	responsesProps
}

// JSONLookup implements an interface to customize json pointer lookup
func (r Responses) JSONLookup(token string) (interface{}, error) {
	if token == "default" {
		return r.Default, nil
	}
	if ex, ok := r.Extensions[token]; ok {
		return &ex, nil
	}
	if i, err := strconv.Atoi(token); err == nil {
		if scr, ok := r.StatusCodeResponses[i]; ok {
			return &scr, nil
		}
	}
	return nil, fmt.Errorf("object has no field %q", token)
}

// UnmarshalJSON hydrates this items instance with the data from JSON
func (r *Responses) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &r.responsesProps); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &r.vendorExtensible); err != nil {
		return err
	}
	if reflect.DeepEqual(responsesProps{}, r.responsesProps) {
		r.responsesProps = responsesProps{}
	}
	return nil
}

// MarshalJSON converts this items object to JSON
func (r Responses) MarshalJSON() ([]byte, error) {
	b1, err := json.Marshal(r.responsesProps)
	if err != nil {
		return nil, err
	}
	b2, err := json.Marshal(r.vendorExtensible)
	if err != nil {
		return nil, err
	}
	concated := swag.ConcatJSON(b1, b2)
	return concated, nil
}

type responsesProps struct {
	Default             *Response
	StatusCodeResponses map[int]Response
}

func (r responsesProps) MarshalJSON() ([]byte, error) {
	toser := map[string]Response{}
	if r.Default != nil {
		toser["default"] = *r.Default
	}
	for k, v := range r.StatusCodeResponses {
		toser[strconv.Itoa(k)] = v
	}
	return json.Marshal(toser)
}

func (r *responsesProps) UnmarshalJSON(data []byte) error {
	var res map[string]Response
	if err := json.Unmarshal(data, &res); err != nil {
		return nil
	}
	if v, ok := res["default"]; ok {
		r.Default = &v
		delete(res, "default")
	}
	for k, v := range res {
		if nk, err := strconv.Atoi(k); err == nil {
			if r.StatusCodeResponses == nil {
				r.StatusCodeResponses = map[int]Response{}
			}
			r.StatusCodeResponses[nk] = v
		}
	}
	return nil
}
