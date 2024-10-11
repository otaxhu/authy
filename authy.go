// Copyright 2024 Oscar Pernia
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package authy

import (
	"context"
	"net/http"
)

// This structure defines a middleware chain/stack where user defined middlewares are chained,
// you can define your own authorization middlewares to handle different kind of authorization
// methods (through header Authorization, request body, different kind of authorization schemes
// like Bearer, Basic, etc.)
type Chain struct {
	chain []func(next http.Handler) http.Handler

	unauthorizedHandler http.Handler
}

// Adds the middlewares to the chain.
func (c *Chain) Use(mw ...func(next http.Handler) http.Handler) {
	c.chain = append(c.chain, mw...)
}

type responseWriterWrapper struct {
	written bool

	http.ResponseWriter
}

func (rw *responseWriterWrapper) Write(b []byte) (int, error) {
	rw.written = true
	return rw.ResponseWriter.Write(b)
}

func (rw *responseWriterWrapper) WriteHeader(statusCode int) {
	rw.written = true
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Builds the Chain to a final middleware that executes all of the middlewares passed to
// Chain#Use() calls, returned value can be passed to a chi.Mux#Use() call.
func (c *Chain) Build() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		ctx := newContext()

		var finalHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if ctx.isUnauthorized || !ctx.isAuthorized {
				wrapper, ok := w.(*responseWriterWrapper)

				if !ok || !wrapper.written {
					c.unauthorizedHandler.ServeHTTP(w, r)
				}
				return
			}

			// Don't spread the Context to other handlers
			r = r.WithContext(context.WithValue(r.Context(), keyContext, nil))

			next.ServeHTTP(w, r)
		})

		var handler http.Handler = finalHandler

		for i := len(c.chain) - 1; i >= 0; i-- {
			handler = c.chain[i](handler)
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wrapper := &responseWriterWrapper{
				ResponseWriter: w,
			}

			r = r.WithContext(context.WithValue(r.Context(), keyContext, ctx))

			handler.ServeHTTP(wrapper, r)

			if ctx.isUnauthorized || !ctx.isAuthorized {
				finalHandler.ServeHTTP(wrapper, r)
			}
		})
	}
}

// Sets handler to be the unathorized handler, this handler is called when either:
//
// 1. At most 1 middleware called `Context#Unauthorize()`, or no middlewares called
// `Context#Authorize()`
//
// 2. No middlewares called `http.ResponseWriter#Write()` nor `http.ResponseWriter#WriteHeader()`,
// this is done this way for backwards compatibility with middlewares that in case of unauthorized
// request, they just stops the chain and just respond to the client with a status code 4xx.
func (c *Chain) UnauthorizedHandler(handler http.Handler) {
	c.unauthorizedHandler = handler
}

// Returns a new Chain, with UnauthorizedHandler set to [DefaultUnauthorizedHandler]
func NewChain() *Chain {
	return &Chain{
		unauthorizedHandler: http.HandlerFunc(DefaultUnauthorizedHandler),
	}
}

// Default handler for unauthorized requests, sets status code to [http.StatusUnauthorized]
func DefaultUnauthorizedHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusUnauthorized)
}
