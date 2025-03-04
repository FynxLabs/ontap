package test

import (
	"os"
	"testing"

	"github.com/pb33f/libopenapi"
)

func TestLibOpenAPI(t *testing.T) {
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
	if model.Model.Paths.Extensions != nil {
		for path, pathItem := range model.Model.Paths.Extensions.FromOldest() {
			t.Logf("Path: %s", path)
			t.Logf("Path item: %v", pathItem)
		}
	}

	// Print the path items
	if model.Model.Paths.PathItems != nil {
		for pathPairs := model.Model.Paths.PathItems.First(); pathPairs != nil; pathPairs = pathPairs.Next() {
			path := pathPairs.Key()
			pathItem := pathPairs.Value()
			t.Logf("Path: %s", path)
			t.Logf("Path item: %v", pathItem)
		}
	}
}
