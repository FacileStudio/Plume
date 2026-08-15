package docs

import "github.com/FacileStudio/tronc/apiref"

type (
	// Response is the API reference registry served as documentation.
	Response = apiref.Registry
	// Module groups related routes in the API reference.
	Module = apiref.Module
	// Route describes a single documented endpoint.
	Route = apiref.Route
	// Field describes a single documented request or response field.
	Field = apiref.Field
	// Error describes a documented error response.
	Error = apiref.Error
)
