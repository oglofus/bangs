//go:build wasm

package http

import (
	"github.com/syumai/workers"
)

func (r *Router) Serve() {
	workers.Serve(nil)
}
