// NOTE:
//
// This example currently doesn't compile nor shows anything because it depends on router library
// Chi which is not added to the project's go.mod
package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/otaxhu/authy"
)

func BasicAuthMiddleware(username, password string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, p, ok := r.BasicAuth()
			if !ok {
				return
			}

			authCtx := authy.GetContext(r)

			if u == username && p == password {
				authCtx.Authorize()
				r = r.WithContext(context.WithValue(r.Context(), "username", u))
				next.ServeHTTP(w, r)
			} else {
				authCtx.Unauthorize()
			}
		})
	}
}

// You could define another Authorization Scheme:

func JwtBearerAuth(dependencyArgs ...any) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			// If it's not Bearer, then let another middleware authorize the request,
			// call next.ServeHTTP and return
			if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
				next.ServeHTTP(w, r)
				return
			}

			// Do jwt authorization with your preferred jwt library and your business logic...

			token := authHeader[len("bearer "):]
			_ = token

			// ... Get claims ...
			var claims map[string]any

			// ... Then check for authorized ...
			var authorized bool

			authCtx := authy.GetContext(r)

			if authorized {
				authCtx.Authorize()

				r = r.WithContext(context.WithValue(r.Context(), "username", claims["sub"].(string)))

				// We always call next.ServeHTTP() if authorized
				//
				// Your business logic may allow stopping the middleware chain if authorized in
				// this middleware. You're free to do so
				next.ServeHTTP(w, r)
			} else {
				authCtx.Unauthorize()
			}
		})
	}
}

// import "github.com/chi-go/chi/v5"
var chi any

func main() {
	mux := chi.NewMux()
	mux.Route("/api/protected-resource", func(router chi.Router) {
		authChain := authy.NewChain()
		authChain.Use(BasicAuthMiddleware("admin", "admin123"))
		authChain.Use(JwtBearerAuth())

		authMw := authChain.Build()

		router.Use(authMw)

		router.Post("/post", func(w http.ResponseWriter, r *http.Request) {
			username := r.Context().Value("username").(string)

			fmt.Fprintf(w, "Hello %s you are authenticated succesfully!", username)
		})
	})

	http.ListenAndServe(":8080", mux)
}
