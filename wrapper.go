package main

import (
	"fmt"
	"net/http"
)

// HTTPHandlerWithMethod only allows the given
// method to a handler, else returning an
// http.StatusMethodNotAllowed status code as
// well as an error
func HTTPHandlerWithMethod(method string, h http.HandlerFunc) http.HandlerFunc {
	return func(r http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			r.Header().Add("Content-Type", "application/json")
			r.WriteHeader(http.StatusMethodNotAllowed)
			r.Write([]byte(fmt.Sprintf(`{"message": "Given method not allowed", "method": %v}`, req.Method)))
			return
		}

		h(r, req)
	}
}

// Post only allows POST requests to a
// handler
func Post(h http.HandlerFunc) http.HandlerFunc {
	return HTTPHandlerWithMethod("POST", h)
}

// Get only allows GET requests to a handler
func Get(h http.HandlerFunc) http.HandlerFunc {
	return HTTPHandlerWithMethod("GET", h)
}

// Patch only allows PATCH requests to a handler
func Patch(h http.HandlerFunc) http.HandlerFunc {
	return HTTPHandlerWithMethod("PATCH", h)
}

// Delete only allows DELETE requests to a handler
func Delete(h http.HandlerFunc) http.HandlerFunc {
	return HTTPHandlerWithMethod("DELETE", h)
}

// Put only allows PUT requests to a handler
func Put(h http.HandlerFunc) http.HandlerFunc {
	return HTTPHandlerWithMethod("PUT", h)
}
