package scan

import (
	goparser "go/parser"
	"log"
	"testing"

	"github.com/casualjim/go-swagger/spec"
	"github.com/stretchr/testify/assert"
)

func TestRoutesParser(t *testing.T) {
	docFile := "../fixtures/goparsing/classification/operations/todo_operation.go"
	fileTree, err := goparser.ParseFile(classificationProg.Fset, docFile, nil, goparser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	rp := newRoutesParser(classificationProg)
	var ops spec.Paths
	err = rp.Parse(fileTree, &ops)
	assert.NoError(t, err)

	assert.Len(t, ops.Paths, 3)

	po, ok := ops.Paths["/pets"]
	assert.True(t, ok)
	assert.NotNil(t, po.Get)
	assertOperation(t,
		po.Get,
		"listPets",
		"Lists pets filtered by some parameters.",
		"This will show all available pets by default.\nYou can get the pets that are out of stock",
		[]string{"pets", "users"},
	)
	assertOperation(t,
		po.Post,
		"createPet",
		"Create a pet based on the parameters.",
		"",
		[]string{"pets", "users"},
	)

	po, ok = ops.Paths["/orders"]
	assert.True(t, ok)
	assert.NotNil(t, po.Get)
	assertOperation(t,
		po.Get,
		"listOrders",
		"lists orders filtered by some parameters.",
		"",
		[]string{"orders"},
	)
	assertOperation(t,
		po.Post,
		"createOrder",
		"create an order based on the parameters.",
		"",
		[]string{"orders"},
	)

	po, ok = ops.Paths["/orders/{id}"]
	assert.True(t, ok)
	assert.NotNil(t, po.Get)
	assertOperation(t,
		po.Get,
		"orderDetails",
		"gets the details for an order.",
		"",
		[]string{"orders"},
	)
	assertOperation(t,
		po.Put,
		"updateOrder",
		"Update the details for an order.",
		"When the order doesn't exist this will return an error.",
		[]string{"orders"},
	)
	assertOperation(t,
		po.Delete,
		"deleteOrder",
		"delete a particular order.",
		"",
		[]string{"orders"},
	)
}

func assertOperation(t *testing.T, op *spec.Operation, id, summary, description string, tags []string) {
	assert.NotNil(t, op)
	assert.Equal(t, summary, op.Summary)
	assert.Equal(t, description, op.Description)
	assert.Equal(t, id, op.ID)
	assert.EqualValues(t, tags, op.Tags)
	assert.EqualValues(t, []string{"application/json", "application/x-protobuf"}, op.Consumes)
	assert.EqualValues(t, []string{"application/json", "application/x-protobuf"}, op.Produces)
	assert.EqualValues(t, []string{"http", "https", "ws", "wss"}, op.Schemes)
	assert.Len(t, op.Security, 2)
	_, ok := op.Security[0]["api_key"]
	assert.True(t, ok)

	vv, ok := op.Security[1]["oauth"]
	assert.True(t, ok)
	assert.EqualValues(t, []string{"read", "write"}, vv)

	assert.NotNil(t, op.Responses.Default)
	assert.Equal(t, "#/responses/genericError", op.Responses.Default.Ref.String())

	rsp, ok := op.Responses.StatusCodeResponses[200]
	assert.True(t, ok)
	assert.Equal(t, "#/responses/someResponse", rsp.Ref.String())
	rsp, ok = op.Responses.StatusCodeResponses[422]
	assert.True(t, ok)
	assert.Equal(t, "#/responses/validationError", rsp.Ref.String())
}
