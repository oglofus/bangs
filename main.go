package main

import (
	"bytes"
	"crypto/sha3"
	_ "embed"
	"encoding/binary"
	"errors"
	"flag"
	"github.com/valyala/fasthttp"
	"log"
	"net/http"
	"strings"
)

const IndexLength uint64 = 44
const KeyLength uint64 = 28
const QueryPlaceholder = 0xC0

//go:embed bangs.idx
var idx []byte

//go:embed bangs.dat
var data []byte

// Default search URL template with placeholder (0xC0) for query
var def []byte

func createDefaultBang(template string) ([]byte, error) {
	if !strings.Contains(template, "<q>") {
		return nil, errors.New("default bang template must contain '<q>' as the query placeholder")
	}

	// Replace <q> with the binary placeholder (0xC0)
	result := strings.Replace(template, "<q>", string([]byte{QueryPlaceholder}), -1)

	return []byte(result), nil
}

func findBang(hash []byte) (bang []byte) {
	var valueLength uint64
	var offset uint64

	var idxLength = uint64(len(idx))

	if idxLength == 0 {
		log.Println("Warning: Index file appears to be empty")

		return
	}

	var rows = idxLength / IndexLength
	var left = uint64(0)
	var right = rows
	var entry = make([]byte, IndexLength)

	for left < right {
		var row = (left + right) / 2
		var i = row * IndexLength

		if i >= idxLength {
			break
		}

		copy(entry, idx[i:i+IndexLength])

		var diff = bytes.Compare(hash[:], entry[:KeyLength])

		if diff == 0 {
			offset = binary.LittleEndian.Uint64(entry[KeyLength:])
			valueLength = binary.LittleEndian.Uint64(entry[KeyLength+8:])

			if offset+valueLength <= uint64(len(data)) {
				bang = data[offset : offset+valueLength]

				return
			} else {
				log.Printf(
					"Warning: Data offset out of bounds: offset=%d, length=%d, data size=%d",
					offset,
					valueLength,
					len(data),
				)

				return
			}
		} else if diff > 0 {
			left = row + 1
		} else {
			right = row
		}
	}

	return
}

func handler(ctx *fasthttp.RequestCtx) {
	var args = ctx.QueryArgs()

	if args.Has("q") {
		var q = args.Peek("q")
		var qLen = len(q)

		if qLen > 0 {
			var bang = def

			var searchLimit = 32
			if searchLimit > qLen {
				searchLimit = qLen
			}

			for i := qLen - 1; i >= qLen-searchLimit; i-- {
				if i < 0 {
					break
				}

				if q[i] == '!' {
					if i+1 < qLen {
						var hash = sha3.Sum224(q[i+1:])
						var foundBang = findBang(hash[:])

						if len(foundBang) > 0 {
							q = q[:i]
							bang = foundBang
						}
					}

					break
				}
			}

			var url = bytes.Replace(bang, []byte{0xC0}, q, -1)

			ctx.Response.Header.Set("Location", string(url))
			ctx.RedirectBytes(url, http.StatusFound)

			return
		}
	}

	ctx.SetStatusCode(200)
}

func main() {
	var addr = flag.String("addr", ":8080", "HTTP server address")
	var defaultBang = flag.String(
		"default",
		"https://www.google.com/search?q=<q>",
		"Default search URL template (must contain <q> as query placeholder)",
	)

	flag.Parse()

	var err error
	if def, err = createDefaultBang(*defaultBang); err != nil {
		log.Fatalf("Error setting default bang: %v", err)
	}

	if len(idx) == 0 {
		log.Println("Warning: Index file is empty")
	}
	if len(data) == 0 {
		log.Println("Warning: Data file is empty")
	}

	log.Printf("Starting server on %s", *addr)
	log.Printf("Using default search: %s", *defaultBang)

	if err = fasthttp.ListenAndServe(*addr, handler); err != nil {
		log.Fatalf("Error in ListenAndServe: %v", err)
	}
}
