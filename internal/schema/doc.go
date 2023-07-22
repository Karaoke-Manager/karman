// Package schema contains HTTP request and response schemas used throughout the Karman API.
// These differ from models mainly through the presence of JSON tags.
// Some schemas however hide model fields or present them in a different way.
//
// A schema must implement render.Renderer if they can be used as a response schema.
// A schema must implement render.Binder if they can be used as a request schema.
package schema
