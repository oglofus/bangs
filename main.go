package main

import (
	"bytes"
	"crypto/sha3"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	h "github.com/oglofus/bangs/internal/http"

	_ "embed"
)

const IndexLength uint64 = 44
const KeyLength uint64 = 28
const QueryPlaceholder = 0xC0
const canonicalOrigin = "https://bangs.oglofus.com"

var seoDescription = "Oglofus Bangs provides fast search redirects using bangs, so you can search other sites directly from one URL."

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

func isValidRedirect(target *url.URL) bool {
	return target.Scheme == "https" && target.Hostname() != "" && target.User == nil
}

func buildRedirectURL(template []byte, query string) (string, bool) {
	encodedQuery := url.QueryEscape(strings.TrimSpace(query))
	rawURL := bytes.Replace(template, []byte{QueryPlaceholder}, []byte(encodedQuery), -1)
	target, err := url.Parse(string(rawURL))
	if err != nil || !isValidRedirect(target) {
		return "", false
	}

	return target.String(), true
}

func validHostname(hostname string) bool {
	if hostname == "" || len(hostname) > 253 {
		return false
	}
	if net.ParseIP(hostname) != nil {
		return true
	}
	for _, label := range strings.Split(hostname, ".") {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, char := range label {
			if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') && (char < '0' || char > '9') && char != '-' {
				return false
			}
		}
	}
	return true
}

func validRequestHost(host string) bool {
	if host == "" || strings.ContainsAny(host, "/?#@") {
		return false
	}
	hostname := host
	if strings.HasPrefix(host, "[") {
		var port string
		var err error
		hostname, port, err = net.SplitHostPort(host)
		if err != nil || net.ParseIP(hostname) == nil || !validPort(port) {
			return false
		}
		return true
	}
	if strings.Count(host, ":") == 1 {
		var port string
		var err error
		hostname, port, err = net.SplitHostPort(host)
		if err != nil || !validPort(port) {
			return false
		}
	} else if strings.Contains(host, ":") {
		return false
	}
	return validHostname(hostname)
}

func validPort(port string) bool {
	if port == "" {
		return false
	}
	for _, char := range port {
		if char < '0' || char > '9' {
			return false
		}
	}
	var value int
	if _, err := fmt.Sscanf(port, "%d", &value); err != nil {
		return false
	}
	return value > 0 && value <= 65535
}

func requestOrigin(req *http.Request) string {
	host := req.Host
	if host == "" {
		host = req.URL.Host
	}
	if !validRequestHost(host) {
		return canonicalOrigin
	}

	scheme := strings.ToLower(req.URL.Scheme)
	if scheme != "http" && scheme != "https" {
		scheme = strings.ToLower(strings.TrimSpace(strings.Split(req.Header.Get("X-Forwarded-Proto"), ",")[0]))
	}
	if scheme != "http" && scheme != "https" {
		scheme = "https"
	}
	return scheme + "://" + host
}

func jsonLD(origin string) string {
	document := map[string]any{
		"@context": "https://schema.org",
		"@graph": []any{
			map[string]any{
				"@type":               "WebApplication",
				"@id":                 origin + "/#application",
				"name":                "Oglofus Bangs",
				"url":                 origin + "/",
				"description":         seoDescription,
				"applicationCategory": "UtilitiesApplication",
				"isPartOf": map[string]string{
					"@id": canonicalOrigin + "/#website",
				},
				"sameAs": []string{canonicalOrigin + "/"},
			},
			map[string]any{
				"@type":       "WebSite",
				"@id":         canonicalOrigin + "/#website",
				"url":         canonicalOrigin + "/",
				"name":        "Oglofus Bangs",
				"description": seoDescription,
			},
		},
	}
	encoded, err := json.Marshal(document)
	if err != nil {
		return "{}"
	}
	return string(encoded)
}

func renderIndex(origin string) []byte {
	page := strings.ReplaceAll(string(index), "__BANGS_CANONICAL__", html.EscapeString(origin+"/"))
	page = strings.ReplaceAll(page, "__BANGS_DESCRIPTION__", html.EscapeString(seoDescription))
	page = strings.ReplaceAll(page, "__BANGS_JSONLD__", jsonLD(origin))
	return []byte(page)
}

func seoHandler(w http.ResponseWriter, req *http.Request) bool {
	origin := requestOrigin(req)
	switch req.URL.Path {
	case "/robots.txt":
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = fmt.Fprintf(w, "User-agent: *\nAllow: /\nSitemap: %s/sitemap.xml\n", origin)
		return true
	case "/sitemap.xml":
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		_, _ = fmt.Fprintf(w, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\"><url><loc>%s/</loc></url></urlset>\n", html.EscapeString(origin))
		return true
	case "/":
		if req.URL.Query().Get("q") == "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(renderIndex(origin))
			return true
		}
	}
	return false
}

func queryHandler(w http.ResponseWriter, req *http.Request) {
	if seoHandler(w, req) {
		return
	}
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
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(renderIndex(requestOrigin(req)))
}

func main() {
	http.HandleFunc("/", queryHandler)
	r := h.Router{Port: "8080"}
	r.Serve()
}
