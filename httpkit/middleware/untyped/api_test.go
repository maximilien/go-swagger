package untyped

import (
	"io"
	"sort"
	"testing"

	"github.com/casualjim/go-swagger/errors"
	"github.com/casualjim/go-swagger/httpkit"
	swaggerspec "github.com/casualjim/go-swagger/spec"
	"github.com/stretchr/testify/assert"
)

func stubAutenticator() httpkit.Authenticator {
	return httpkit.AuthenticatorFunc(func(_ interface{}) (bool, interface{}, error) { return false, nil, nil })
}

type stubConsumer struct {
}

func (s *stubConsumer) Consume(_ io.Reader, _ interface{}) error {
	return nil
}

type stubProducer struct {
}

func (s *stubProducer) Produce(_ io.Writer, _ interface{}) error {
	return nil
}

type stubOperationHandler struct {
}

func (s *stubOperationHandler) ParameterModel() interface{} {
	return nil
}

func (s *stubOperationHandler) Handle(params interface{}) (interface{}, error) {
	return nil, nil
}

func TestUntypedAPIRegistrations(t *testing.T) {
	api := NewAPI(new(swaggerspec.Document))

	api.RegisterConsumer("application/yada", new(stubConsumer))
	api.RegisterProducer("application/yada-2", new(stubProducer))
	api.RegisterOperation("someId", new(stubOperationHandler))
	api.RegisterAuth("basic", stubAutenticator())

	assert.NotEmpty(t, api.authenticators)

	_, ok := api.authenticators["basic"]
	assert.True(t, ok)
	_, ok = api.consumers["application/yada"]
	assert.True(t, ok)
	_, ok = api.producers["application/yada-2"]
	assert.True(t, ok)
	_, ok = api.consumers["application/json"]
	assert.True(t, ok)
	_, ok = api.producers["application/json"]
	assert.True(t, ok)
	_, ok = api.operations["someId"]
	assert.True(t, ok)

	h, ok := api.OperationHandlerFor("someId")
	assert.True(t, ok)
	assert.NotNil(t, h)

	_, ok = api.OperationHandlerFor("doesntExist")
	assert.False(t, ok)
}

func TestUntypedAppValidation(t *testing.T) {
	invalidSpecStr := `{
  "consumes": ["application/json"],
  "produces": ["application/json"],
  "security": [
    {"apiKey":[]}
  ],
  "parameters": {
    "format": {
      "in": "query",
      "name": "format",
      "type": "string"
    }
  },
  "paths": {
    "/": {
      "parameters": [
        {
          "name": "limit",
          "type": "integer",
          "format": "int32",
          "x-go-name": "Limit"
        }
      ],
      "get": {
        "consumes": ["application/x-yaml"],
        "produces": ["application/x-yaml"],
        "security": [
          {"basic":[]}
        ],
        "operationId": "someOperation",
        "parameters": [
          {
            "name": "skip",
            "type": "integer",
            "format": "int32"
          }
        ]
      }
    }
  }
}`
	specStr := `{
	  "consumes": ["application/json"],
	  "produces": ["application/json"],
	  "security": [
	    {"apiKey":[]}
	  ],
	  "securityDefinitions": {
	    "basic": { "type": "basic" },
	    "apiKey": { "type": "apiKey", "in":"header", "name":"X-API-KEY" }
	  },
	  "parameters": {
	  	"format": {
	  		"in": "query",
	  		"name": "format",
	  		"type": "string"
	  	}
	  },
	  "paths": {
	  	"/": {
	  		"parameters": [
	  			{
	  				"name": "limit",
			  		"type": "integer",
			  		"format": "int32",
			  		"x-go-name": "Limit"
			  	}
	  		],
	  		"get": {
	  			"consumes": ["application/x-yaml"],
	  			"produces": ["application/x-yaml"],
	        "security": [
	          {"basic":[]}
	        ],
	  			"operationId": "someOperation",
	  			"parameters": [
	  				{
				  		"name": "skip",
				  		"type": "integer",
				  		"format": "int32"
				  	}
	  			]
	  		}
	  	}
	  }
	}`
	validSpec, err := swaggerspec.New([]byte(specStr), "")
	assert.NoError(t, err)
	assert.NotNil(t, validSpec)

	spec, err := swaggerspec.New([]byte(invalidSpecStr), "")
	assert.NoError(t, err)
	assert.NotNil(t, spec)

	cons := spec.ConsumesFor(spec.AllPaths()["/"].Get)
	assert.Len(t, cons, 2)
	prods := spec.RequiredProduces()
	assert.Len(t, prods, 2)

	api1 := NewAPI(spec)
	err = api1.Validate()
	assert.Error(t, err)
	assert.Equal(t, "missing [application/x-yaml] consumes registrations", err.Error())
	api1.RegisterConsumer("application/x-yaml", new(stubConsumer))
	err = api1.validate()
	assert.Error(t, err)
	assert.Equal(t, "missing [application/x-yaml] produces registrations", err.Error())
	api1.RegisterProducer("application/x-yaml", new(stubProducer))
	err = api1.validate()
	assert.Error(t, err)
	assert.Equal(t, "missing [someOperation] operation registrations", err.Error())
	api1.RegisterOperation("someOperation", new(stubOperationHandler))
	err = api1.validate()
	assert.Error(t, err)
	assert.Equal(t, "missing [apiKey, basic] auth scheme registrations", err.Error())
	api1.RegisterAuth("basic", stubAutenticator())
	api1.RegisterAuth("apiKey", stubAutenticator())
	err = api1.validate()
	assert.Error(t, err)
	assert.Equal(t, "missing [apiKey, basic] security definitions registrations", err.Error())

	api3 := NewAPI(validSpec)
	api3.RegisterConsumer("application/x-yaml", new(stubConsumer))
	api3.RegisterProducer("application/x-yaml", new(stubProducer))
	api3.RegisterOperation("someOperation", new(stubOperationHandler))
	api3.RegisterAuth("basic", stubAutenticator())
	api3.RegisterAuth("apiKey", stubAutenticator())
	err = api3.validate()
	assert.NoError(t, err)
	api3.RegisterConsumer("application/something", new(stubConsumer))
	err = api3.validate()
	assert.Error(t, err)
	assert.Equal(t, "missing from spec file [application/something] consumes", err.Error())

	api2 := NewAPI(spec)
	api2.RegisterConsumer("application/something", new(stubConsumer))
	err = api2.validate()
	assert.Error(t, err)
	assert.Equal(t, "missing [application/x-yaml] consumes registrations\nmissing from spec file [application/something] consumes", err.Error())
	api2.RegisterConsumer("application/x-yaml", new(stubConsumer))
	delete(api2.consumers, "application/something")
	api2.RegisterProducer("application/something", new(stubProducer))
	err = api2.validate()
	assert.Error(t, err)
	assert.Equal(t, "missing [application/x-yaml] produces registrations\nmissing from spec file [application/something] produces", err.Error())
	delete(api2.producers, "application/something")
	api2.RegisterProducer("application/x-yaml", new(stubProducer))

	expected := []string{"application/json", "application/x-yaml"}
	sort.Sort(sort.StringSlice(expected))
	consumes := spec.ConsumesFor(spec.AllPaths()["/"].Get)
	sort.Sort(sort.StringSlice(consumes))
	assert.Equal(t, expected, consumes)
	consumers := api1.ConsumersFor(consumes)
	assert.Len(t, consumers, 2)

	produces := spec.ProducesFor(spec.AllPaths()["/"].Get)
	sort.Sort(sort.StringSlice(produces))
	assert.Equal(t, expected, produces)
	producers := api1.ProducersFor(produces)
	assert.Len(t, producers, 2)

	definitions := validSpec.SecurityDefinitionsFor(validSpec.AllPaths()["/"].Get)
	expectedSchemes := map[string]swaggerspec.SecurityScheme{"basic": *swaggerspec.BasicAuth()}
	assert.Equal(t, expectedSchemes, definitions)
	authenticators := api3.AuthenticatorsFor(definitions)
	assert.Len(t, authenticators, 1)

	opHandler := httpkit.OperationHandlerFunc(func(data interface{}) (interface{}, error) {
		return data, nil
	})
	d, err := opHandler.Handle(1)
	assert.NoError(t, err)
	assert.Equal(t, 1, d)

	authenticator := httpkit.AuthenticatorFunc(func(params interface{}) (bool, interface{}, error) {
		if str, ok := params.(string); ok {
			return ok, str, nil
		}
		return true, nil, errors.Unauthenticated("authenticator")
	})
	ok, p, err := authenticator.Authenticate("hello")
	assert.True(t, ok)
	assert.NoError(t, err)
	assert.Equal(t, "hello", p)
}
