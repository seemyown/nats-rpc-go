package natsrpc

// HandlerFunc defines the handler function type for RPC requests.
// A handler takes a *Context and is responsible for processing the request
// and sending a response (usually via Context.JSON or other Context methods).
type HandlerFunc func(*Context)

// MiddlewareFunc defines a middleware function type.
// Middleware wraps a HandlerFunc, allowing you to execute code
// before and/or after the next handler in the chain.
// It should call next(c) to pass control to the next handler.
type MiddlewareFunc func(HandlerFunc) HandlerFunc
