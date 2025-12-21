# Oglofus Bangs (Cloudflare Workers)

Oglofus Bangs is a high-performance, open-source alternative to DuckDuckGo's bangs. It provides lightning-fast
redirections using hashed binary search for bang commands, and now runs on Cloudflare Workers for global, low‑latency
availability.

## What are bangs?

Bangs are shortcuts that quickly take you to search results on other sites. For example, when you know you want to
search on another site like Wikipedia or Amazon, bangs get you there fastest. A search for `filter bubble !w` will take
you directly to Wikipedia.

## Why Oglofus Bangs?

- **Speed**: Utilizes hashed binary search for extremely fast bang lookups
- **Edge-native**: Runs on Cloudflare Workers for globally distributed, low‑latency responses
- **Lightweight**: Minimal dependencies and efficient memory usage (Go → WebAssembly)
- **Self-hosted**: Deploy to your own Cloudflare account
- **Customizable**: Easily add or modify bangs to suit your needs

## How It Works

Oglofus Bangs uses SHA3-224 hashing and binary search to efficiently locate the appropriate redirection URL for a given
bang command. The system:

1. Extracts the bang command from search queries (text following the `!` character)
2. Hashes the command using SHA3-224
3. Performs a binary search on a pre-compiled index file to find the matching URL
4. Redirects the user to the target site with their search query

This approach provides O(log n) lookup performance even with thousands of bangs.

## Deployment 

### Cloudflare Workers

This project targets Cloudflare Workers using the `github.com/syumai/workers` runtime. You can run it locally via
Wrangler and deploy it to your Cloudflare account.

One-click deploy:

[![Deploy to Cloudflare](https://deploy.workers.cloudflare.com/button)](https://deploy.workers.cloudflare.com/?url=https%3A%2F%2Fgithub.com%2Foglofus%2Fbangs)

#### Prerequisites

- Go 1.25+ (WASM build target)
- Node.js 18+ and npm
- Cloudflare account
- Wrangler CLI (installed locally via devDependency)

#### Quick start

```bash
# Install JS tooling (Wrangler)
npm install

# (Optional) Login once to your Cloudflare account
npx wrangler login

# Build the worker (generates build/app.wasm and build/worker.mjs)
npm run build

# Start local dev server (Miniflare)
npm run dev
# → Open the printed http://127.0.0.1:8787 URL

# Deploy to your Cloudflare account
npm run deploy
```

Notes:
- The worker entry is configured in `wrangler.jsonc` (main: ./build/worker.mjs).
- On first deploy, Wrangler will guide you to select an account and create the worker.

### Container

Alternatively this project can also be run with Docker or another compatible container platform, like Podman. The project provides a Dockerfile which exposes `8080` port.

## Usage

Once the worker is running (locally or deployed), use the `q` query parameter containing your search term and bang:

Example:

```
http://127.0.0.1:8787/?q=filter%20bubble%20!w
```

This will redirect you to Wikipedia's search for "filter bubble".

Important: Unlike DuckDuckGo's implementation, Oglofus Bangs only recognizes bangs that appear at the end of the
query string. For example:

- ✅ `filter bubble !w` - Will work correctly
- ❌ `!w filter bubble` - Will not be recognized as a bang command

## Managing/Updating Bangs

Bangs are defined in `bangs.json` (not checked into the worker bundle). Each bang has a trigger (`t`) and a URL template (`u`):

```json
[
  { "t": "w",  "u": "https://en.wikipedia.org/wiki/Special:Search?search=<q>" },
  { "t": "gh", "u": "https://github.com/search?q=<q>" }
]
```

After modifying `bangs.json`, regenerate the binary files used by the worker:

```bash
# Generate bangs.idx and bangs.dat from bangs.json
go run ./preprocessor/main.go

# Rebuild the worker to embed the updated data files
npm run build
```

This produces:
- `bangs.idx`: hashed keys + offsets
- `bangs.dat`: URL templates with `<q>` replaced by a binary placeholder

## Configuration

- Default search engine: The worker falls back to Google when no bang is found. To change the default, edit `main.go`
  (the `def` variable) and rebuild:
  ```
  var def = append([]byte("https://duckduckgo.com/?q="), QueryPlaceholder)
  ```
  Then run `npm run build` and redeploy.

## Technical Details

- Runtime: Cloudflare Workers via `github.com/syumai/workers`
- Language: Go compiled to WebAssembly (GOOS=js, GOARCH=wasm)
- Hashing: SHA3-224 for bang keys
- Storage: two embedded binary files (`bangs.idx`, `bangs.dat`)
- `<q>` in URL templates is converted to a single-byte placeholder (0xC0) at build time

## Performance

Oglofus Bangs is designed for high performance:

- O(log n) lookup time for bangs
- Minimal memory footprint
- Runs at the edge on Cloudflare's global network
- Efficient binary data format

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
