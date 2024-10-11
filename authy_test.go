package authy

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestChainUse(t *testing.T) {

	buf := &strings.Builder{}

	fnNoopFactory := func(message string) func(next http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				buf.WriteString(message)
				next.ServeHTTP(w, r)
			})
		}
	}

	fnAuthorizedFactory := func(message string) func(next http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := GetContext(r)
				ctx.Authorize()
				buf.WriteString(message)
				next.ServeHTTP(w, r)
			})
		}
	}

	fnUnauthorizedFactory := func(message string) func(next http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := GetContext(r)
				ctx.Unauthorize()
				buf.WriteString(message)
				next.ServeHTTP(w, r)
			})
		}
	}

	testCases := map[string]struct {
		Middlewares         []func(next http.Handler) http.Handler
		ExpectedBuf         string
		FinalHandlerMessage string
		MustAuthorize       bool
	}{
		"Unauthorized_AllNoop": {
			Middlewares: []func(http.Handler) http.Handler{
				fnNoopFactory("1"),
				fnNoopFactory("2"),
				fnNoopFactory("3"),
			},
			FinalHandlerMessage: "4",
			ExpectedBuf:         "123",
			MustAuthorize:       false,
		},
		"Unauthorized_AllUnauthorized": {
			Middlewares: []func(http.Handler) http.Handler{
				fnUnauthorizedFactory("1"),
				fnUnauthorizedFactory("2"),
				fnUnauthorizedFactory("3"),
			},
			FinalHandlerMessage: "4",
			ExpectedBuf:         "123",
			MustAuthorize:       false,
		},
		"Unauthorized_SomeUnauthorized_SomeAuthorized": {
			Middlewares: []func(http.Handler) http.Handler{
				fnUnauthorizedFactory("1"),
				fnAuthorizedFactory("2"),
				fnAuthorizedFactory("3"),
			},
			FinalHandlerMessage: "4",
			ExpectedBuf:         "123",
			MustAuthorize:       false,
		},
		"Unauthorized_SomeUnauthorized_SomeAuthorized_SomeNoop": {
			Middlewares: []func(http.Handler) http.Handler{
				fnUnauthorizedFactory("1"),
				fnAuthorizedFactory("2"),
				fnNoopFactory("3"),
			},
			FinalHandlerMessage: "4",
			ExpectedBuf:         "123",
			MustAuthorize:       false,
		},
		"Unauthorized_SomeUnauthorized_SomeNoop": {
			Middlewares: []func(http.Handler) http.Handler{
				fnUnauthorizedFactory("1"),
				fnNoopFactory("2"),
				fnNoopFactory("3"),
			},
			FinalHandlerMessage: "4",
			ExpectedBuf:         "123",
			MustAuthorize:       false,
		},
		"Authorized_AllAuthorized": {
			Middlewares: []func(next http.Handler) http.Handler{
				fnAuthorizedFactory("1"),
				fnAuthorizedFactory("2"),
				fnAuthorizedFactory("3"),
			},
			FinalHandlerMessage: "4",
			ExpectedBuf:         "1234",
			MustAuthorize:       true,
		},
		"Authorized_SomeAuthorized_SomeNoop": {
			Middlewares: []func(next http.Handler) http.Handler{
				fnAuthorizedFactory("1"),
				fnNoopFactory("2"),
				fnNoopFactory("3"),
			},
			FinalHandlerMessage: "4",
			ExpectedBuf:         "1234",
			MustAuthorize:       true,
		},
	}

	for name, tc := range testCases {
		buf.Reset()
		t.Run(name, func(t *testing.T) {
			authChain := NewChain()

			authChain.Use(tc.Middlewares...)

			req := httptest.NewRequest("", "/", nil)
			res := httptest.NewRecorder()

			mw := authChain.Build()

			mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !tc.MustAuthorize {
					t.Fatalf("middleware authorization chain should not reach final handler")
				}

				buf.WriteString(tc.FinalHandlerMessage)
			})).ServeHTTP(res, req)

			if buf.String() != tc.ExpectedBuf {
				t.Fatalf("expected '%s', got '%s'", tc.ExpectedBuf, buf.String())
			}
		})
	}
}
