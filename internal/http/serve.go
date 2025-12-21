//go:build !wasm

package http

import (
	"net/http"
)

func (r *Router) Serve() {
	http.ListenAndServe(":"+r.Port, nil)
}
