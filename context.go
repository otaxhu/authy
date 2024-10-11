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
	"net/http"
)

// This structure defines whether the request is authorized or not, it is handled in middleware
// returned by Chain#Build()
type Context struct {
	// It's counter-intuitive to have this two variables, but it's this way so that if no
	// middleware authorize or unauthorize the request, that should be unauthorized.
	//
	// If a middleware authorize but another unauthorize, that would be a conflict, it should be
	// unauthorized.
	//
	// It's only authorized if isAuthorized && !isUnauthorized.
	isAuthorized, isUnauthorized bool
}

type keyAuthContext struct{}

var keyContext = keyAuthContext{}

// Gets the Context from the request
func GetContext(r *http.Request) *Context {
	return r.Context().Value(keyContext).(*Context)
}

func newContext() *Context {
	return &Context{}
}

// Authorize the request
func (c *Context) Authorize() {
	c.isAuthorized = true
}

// Unauthorize the request
func (c *Context) Unauthorize() {
	c.isUnauthorized = true
}
