# Authy - Golang Authorization Middleware library.

# :rocket: *Compatible with [Chi Router](https://github.com/go-chi/chi)*

## How does this library work?

The focus of this library is HTTP authorization (not to be confused with authentication) through HTTP Golang middlewares.

There are two main structures defined in this library, the **`Chain`**, **`Context`** and related functions.

### `Chain`:
This structure defines a middleware chain/stack where user defined middlewares are chained, you can define your own authorization middlewares to handle different kind of authorization methods (through header `Authorization`, request body, different kind of authorization schemes like `Bearer` or `Basic`, etc.)

  - #### `Chain#Use()`:
    Adds the middlewares to the chain.

  - #### `Chain#Build()`:
    Builds the `Chain` to a final middleware that executes all of the middlewares passed to `Chain#Use()` calls, returned value can be passed to a `chi.Mux#Use()` call.

  - #### `Chain#UnauthorizedHandler()`:
    Sets handler to be the unathorized handler, this handler is called when either:
      1. At most 1 middleware called `Context#Unauthorize()`, or no middlewares called `Context#Authorize()`
      2. No middlewares called `http.ResponseWriter#Write()` nor `http.ResponseWriter#WriteHeader()`, this is done this way for backwards compatibility with middlewares that in case of unauthorized request, they just stops the chain and just respond to the client with a status code 4xx.

### `Context`:
This structure is a very important one, this structure tells to `Chain` that authorization was succesful and the chain can continue to the next HTTP handler, if the authorization was unsuccesful, then `Chain` doesn't call the next HTTP handler and responds with a 4xx HTTP status code.

  - #### `Context#Authorize()`:
    Authorize the request, can be called multiple times.

  - #### `Context#Unauthorize()`:
    Unauthorize the request, can be called multiple times.


## Functions:
### `GetContext()`

### `NewChain()`
