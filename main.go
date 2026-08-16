package main

import (
	"bytes"
	"crypto/sha3"
	"encoding/binary"
	"log"
	"net/http"
	"net/url"
	"strings"

	h "github.com/oglofus/bangs/internal/http"

	_ "embed"
)

const IndexLength uint64 = 44
const KeyLength uint64 = 28
const QueryPlaceholder = 0xC0

//go:embed bangs.idx
var idx []byte

//go:embed bangs.dat
var data []byte

//go:embed index.html
var index []byte

// Default search URL template with placeholder (0xC0) for query
var def = append([]byte("https://www.google.com/search?q="), QueryPlaceholder)

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

func buildRedirectURL(template []byte, query string) (string, bool) {
	encodedQuery := url.QueryEscape(strings.TrimSpace(query))
	rawURL := bytes.Replace(template, []byte{QueryPlaceholder}, []byte(encodedQuery), -1)
	target, err := url.Parse(string(rawURL))
	if err != nil || target.Scheme != "https" || target.Hostname() == "" || target.User != nil {
		return "", false
	}

	return target.String(), true
}

func queryHandler(w http.ResponseWriter, req *http.Request) {
	var args = req.URL.Query()
	var bang = def

	if args.Has("q") {
		var q = []byte(args.Get("q"))
		var qLen = len(q)

		if qLen > 0 {
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

			redirectURL, ok := buildRedirectURL(bang, string(q))
			if !ok {
				http.Error(w, "invalid redirect target", http.StatusInternalServerError)
				return
			}

			http.Redirect(w, req, redirectURL, http.StatusFound)

			return
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	_, _ = w.Write(index)
}

func main() {
	http.HandleFunc("/", queryHandler)
	r := h.Router{Port: "8080"}
	r.Serve()
}
