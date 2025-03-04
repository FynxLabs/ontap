package test

import (
	"os"
	"testing"

	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

func TestLibOpenAPITypes(t *testing.T) {
	// Read the OpenAPI spec
	data, err := os.ReadFile("./fixtures/openapi.yaml")
	if err != nil {
		t.Fatalf("Failed to read OpenAPI spec: %v", err)
	}

	// Create a new document
	document, err := libopenapi.NewDocument(data)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Build the model
	model, errs := document.BuildV3Model()
	if len(errs) > 0 {
		for _, err := range errs {
			t.Logf("Error building model: %v", err)
		}
		t.Fatalf("Failed to build model: %d errors", len(errs))
	}

	// Print the model
	t.Logf("OpenAPI version: %s", model.Model.Info.Version)
	t.Logf("OpenAPI title: %s", model.Model.Info.Title)

	// Print the paths
	for pathPairs := model.Model.Paths.PathItems.First(); pathPairs != nil; pathPairs = pathPairs.Next() {
		path := pathPairs.Key()
		pathItem := pathPairs.Value()
		t.Logf("Path: %s", path)

		// Print operations
		if pathItem.Get != nil {
			printOperation(t, "GET", pathItem.Get)
		}
		if pathItem.Post != nil {
			printOperation(t, "POST", pathItem.Post)
		}
		// Add other methods as needed
	}
}

func printOperation(t *testing.T, method string, operation *v3.Operation) {
	t.Logf("  %s: %s", method, operation.OperationId)
	t.Logf("    Summary: %s", operation.Summary)
	t.Logf("    Description: %s", operation.Description)

	// Print parameters
	t.Logf("    Parameters:")
	for _, param := range operation.Parameters {
		t.Logf("      Name: %s", param.Name)
		t.Logf("      In: %s", param.In)
		t.Logf("      Required: %v", param.Required)
		t.Logf("      Deprecated: %v", param.Deprecated)
		t.Logf("      Type: %T", param.Required)
		t.Logf("      Type: %T", param.Deprecated)
	}

	// Print security
	t.Logf("    Security:")
	if operation.Security != nil {
		t.Logf("      Security is not nil")
		t.Logf("      Security type: %T", operation.Security)
		for i, security := range operation.Security {
			t.Logf("      Security[%d] type: %T", i, security)
			t.Logf("      Security[%d] Requirements type: %T", i, security.Requirements)
			for reqPairs := security.Requirements.First(); reqPairs != nil; reqPairs = reqPairs.Next() {
				name := reqPairs.Key()
				scopes := reqPairs.Value()
				t.Logf("        %s: %v", name, scopes)
			}
		}
	} else {
		t.Logf("      Security is nil")
	}
}
