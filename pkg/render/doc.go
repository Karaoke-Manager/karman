// Package render implements parsing and serializing HTTP requests and responses.
// This package was originally forked from https://github.com/go-chi/render.
// Its functionality can be split into three parts that complement each other:
//  1. Parsing Request Data
//  2. Serializing Response Data
//  3. Content Type Negotiation
//
// # General Design Decisions
//
// This package aims to be compliant with defined HTTP standard behavior,
// giving you the opportunity to hook into the process at different points.
// This generally means that the render package makes decoding decisions based on the Content-Type header of a request and
// maes encoding decisions based on its Accept header.
// If you want to enforce a certain behavior (e.g. always use a specific content type) the intended way is to implement
// a middleware that overwrites the respective request headers.
//
// The render package becomes relevant at different points in an application's lifecycle:
//   - During the initialization phase you import packages that use [RegisterDecoder] and [RegisterEncoder]
//     to provide implementations for various data formats.
//   - When the request reaches a handler you might use the functions [Bind] or [Decode] to automatically parse the
//     request data using one of the registered decoders. See below for details.
//   - During the processing of a request you might want to use the [ContentTypeNegotiation] middleware or the
//     functions [GetAcceptedMediaTypes] and [NegotiateContentType] to determine the format and content of response you want to send.
//   - To send the response data you can use the [Render] or [Respond] function.
//     This function encodes the response using the results of content type negotiation and one of the registered encoders.
//
// # Content Type Negotiation
//
// The render package supports content negotiation using the "Accept" header.
// The required algorithms are implemented largely by the [github.com/Karaoke-Manager/karman/pkg/mediatype] package.
// Content type negotiation can be done in several ways:
//   - The most convenient way is probably to use the [ContentTypeNegotiation] middleware.
//     This middleware automatically selects a fitting content type from the intersection of the Accept header
//     and a list of available types.
//   - In your code you can then use the functions [GetAcceptedMediaTypes] and [GetNegotiatedContentType] to make decisions
//     based on the result.
//     If necessary you can also initiate a re-negotiation using the [NegotiateContentType] function.
//   - The [Render] and [Respond] functions will then use the negotiation result to decide how to encode the response.
//     See the documentation on [Respond] for details.
//
// Other types of content negotiation (read: other Accept-* headers) are not handled by this package.
//
// # Decoding and Binding Requests
//
// The render package supports automatic decoding of requests using the [Decode] function.
// The decoding process works like this:
//  1. The request is routed through middlewares that potentially restrict the set of allowed content types.
//     This is also where you might want to use the middleware [DefaultRequestContentType] and [SetRequestContentType] to potentially
//     force a specific decoding behavior in step 3.
//  2. You request handler hands the request to the [Decode] or the [Bind] function (which internally also uses [Decode]).
//  3. The [Decode] function selects one of the decoders registered at initialization time via [RegisterDecoder] based
//     on the Content-Type of the request. See the documentation on those functions for details.
//     The request body is then decoded into a Go object by the chosen decoder.
//  4. If you used the [Bind] function the resulting value's [Binder.Bind] method is called to finish the decoding process.
//  5. Any errors along the way are passed to the caller in step 2 to handle the error appropriately.
//
// # Rendering Responses
//
// Similarly to the decoding process there exists an analogous process for rendering and encoding responses.
// The response process works like this:
//  1. During the handling of the request, content type negotiation is performed to determine an appropriate response format.
//     The determined format may or may not influence the behavior of your handler.
//  2. Your handler passes the response data as a Go value to [Respond] or [Render] (which uses [Respond] internally).
//  3. If you used the [Render] function the response value's [Renderer.Render] function gets called to prepare the response.
//  4. If the response object implements the [Responder] interface the [Responder.PrepareResponse] method is called,
//     yielding the final response value.
//  5. Using the content type from step 1 an appropriate encoder is chosen among the ones registered via [RegisterEncoder].
//  6. The chosen encoder serializes the response, finishing the process.
//     Most errors are returned to the caller of [Render] or [Respond].
//     There are however some errors that are considered programmer errors that cause a panic. See [Respond] for details.
//
// # Using Encoders and Decoders
//
// The render package intentionally does not register any encoders or decoders by default.
// There are, however, several subpackages that provide implementations for various common formats.
// Usually packages that provide encoders or decoders register these in their init() functions so
// it is enough to import these packages unnamed.
package render
