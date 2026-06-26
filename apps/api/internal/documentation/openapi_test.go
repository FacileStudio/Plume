package docs

import "testing"

func TestOpenAPI(t *testing.T) {
	spec := OpenAPI(Response{
		Modules: []Module{
			{
				Name:        "signers",
				Description: "Manage signers.",
				Routes: []Route{
					{
						Method:       "POST",
						Path:         "/documents/{docId}/signers",
						Summary:      "Add a signer",
						Auth:         "bearer",
						RequestBody:  "AddSignerRequest",
						ResponseBody: "SignerResponse",
						PathParams:   []Field{{Name: "docId", Type: "int", Description: "Document ID"}},
						Errors:       []Error{{Status: 404, Code: "not_found", Description: "document missing"}},
					},
				},
			},
		},
	})

	if spec["openapi"] != "3.1.0" {
		t.Fatalf("openapi version = %v, want 3.1.0", spec["openapi"])
	}

	paths := spec["paths"].(map[string]any)
	path, ok := paths["/documents/{docId}/signers"].(map[string]any)
	if !ok {
		t.Fatalf("missing path; got %v", paths)
	}
	op, ok := path["post"].(map[string]any)
	if !ok {
		t.Fatalf("method should be lowercased to 'post'; got %v", path)
	}

	if _, ok := op["security"]; !ok {
		t.Error("bearer route must carry security")
	}
	params := op["parameters"].([]any)
	if len(params) != 1 {
		t.Fatalf("want 1 path param, got %d", len(params))
	}
	if got := params[0].(map[string]any)["schema"].(map[string]any)["type"]; got != "integer" {
		t.Errorf("int param schema type = %v, want integer", got)
	}
	if _, ok := op["requestBody"]; !ok {
		t.Error("missing requestBody")
	}

	responses := op["responses"].(map[string]any)
	if _, ok := responses["200"]; !ok {
		t.Error("missing 200 response")
	}
	if _, ok := responses["404"]; !ok {
		t.Error("declared error status 404 not mapped into responses")
	}

	if _, ok := spec["components"].(map[string]any)["securitySchemes"].(map[string]any)["bearerAuth"]; !ok {
		t.Error("missing bearerAuth security scheme")
	}
}
