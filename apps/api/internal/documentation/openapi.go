package docs

import (
	"fmt"
	"strings"
)

// OpenAPI converts the documentation registry into a minimal OpenAPI 3.1
// document so the same hand-maintained route metadata can drive a Scalar UI.
func OpenAPI(resp Response) map[string]any {
	paths := map[string]any{}
	tags := make([]any, 0, len(resp.Modules))

	for _, module := range resp.Modules {
		tags = append(tags, map[string]any{
			"name":        module.Name,
			"description": module.Description,
		})
		for _, route := range module.Routes {
			op := map[string]any{
				"tags":        []any{module.Name},
				"summary":     route.Summary,
				"operationId": operationID(route.Method, route.Path),
				"responses":   responses(route),
			}
			if route.Description != "" {
				op["description"] = route.Description
			}
			if route.Auth == "bearer" {
				op["security"] = []any{map[string]any{"bearerAuth": []any{}}}
			}
			if params := pathParameters(route.PathParams); len(params) > 0 {
				op["parameters"] = params
			}
			if route.RequestBody != "" {
				op["requestBody"] = map[string]any{
					"required": true,
					"content": map[string]any{
						"application/json": map[string]any{"schema": namedSchema(route.RequestBody)},
					},
				}
			}

			path, ok := paths[route.Path].(map[string]any)
			if !ok {
				path = map[string]any{}
				paths[route.Path] = path
			}
			path[strings.ToLower(route.Method)] = op
		}
	}

	return map[string]any{
		"openapi": "3.1.0",
		"info": map[string]any{
			"title":       "Plume API",
			"version":     "1.0.0",
			"description": "Self-hosted document signing platform. All routes are served under the /api prefix.",
		},
		"servers": []any{map[string]any{"url": "/api"}},
		"tags":    tags,
		"paths":   paths,
		"components": map[string]any{
			"securitySchemes": map[string]any{
				"bearerAuth": map[string]any{"type": "http", "scheme": "bearer"},
			},
		},
	}
}

func operationID(method, path string) string {
	cleaned := strings.NewReplacer("/", "_", "{", "", "}", "").Replace(path)
	return strings.ToLower(method) + strings.TrimRight("_"+strings.Trim(cleaned, "_"), "_")
}

func pathParameters(fields []Field) []any {
	params := make([]any, 0, len(fields))
	for _, f := range fields {
		params = append(params, map[string]any{
			"name":        f.Name,
			"in":          "path",
			"required":    true,
			"description": f.Description,
			"schema":      map[string]any{"type": openapiType(f.Type)},
		})
	}
	return params
}

func responses(route Route) map[string]any {
	desc := route.ResponseBody
	if desc == "" {
		desc = "Success"
	}
	success := map[string]any{"description": desc}
	if route.ResponseBody != "" {
		success["content"] = map[string]any{
			"application/json": map[string]any{"schema": namedSchema(route.ResponseBody)},
		}
	}
	out := map[string]any{"200": success}
	for _, e := range route.Errors {
		out[fmt.Sprintf("%d", e.Status)] = map[string]any{
			"description": strings.TrimSpace(e.Code + " " + e.Description),
		}
	}
	return out
}

func namedSchema(name string) map[string]any {
	if inner, ok := strings.CutPrefix(name, "[]"); ok {
		return map[string]any{
			"type":  "array",
			"items": map[string]any{"type": "object", "description": inner},
		}
	}
	return map[string]any{"type": "object", "description": name}
}

func openapiType(t string) string {
	switch t {
	case "int", "integer", "int64":
		return "integer"
	case "bool", "boolean":
		return "boolean"
	case "number", "float", "float64":
		return "number"
	default:
		return "string"
	}
}
