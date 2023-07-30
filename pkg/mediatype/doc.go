// Package mediatype provides an implementation of RFC 6838 media types (also known as mime types).
// The package offers a type MediaType that can be used to conveniently work with media types.
// In addition, the mediatype package implements a media type negotiation algorithm
// that can be used to process a HTTP Accept header.
//
// This package only understands media types in the format "type/subtype", optionally followed by parameters.
// This package does not support media types that only consist of a major type.
//
// A MediaType value can identify a concrete type or a wildcard type.
// You can use MediaType.IsCompatibleWith and MediaType.Includes to find out if two media types fit together.
// There are different kinds of wildcard types supported by the mediatype package:
//   - The special media type "*/*" includes any other media type
//   - A wildcard subtype matches all types in a specific tree.
//     For example "text/*" matches all types beginning with "text/", such as "text/plain".
//   - A suffix wildcard matches all types in a specific tree that have a given suffix.
//     For example "application/*+json" matches "application/problem+json" and "application/json", but not "application/problem+xml".
//     Note that this is the only special case of a wildcard.
//     Arbitrary wildcards (as in "application/j*n") are not supported.
package mediatype
