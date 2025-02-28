# Oglofus Bangs [![Docker](https://github.com/oglofus/bangs/actions/workflows/docker-publish.yml/badge.svg?branch=main)](https://github.com/oglofus/bangs/actions/workflows/docker-publish.yml)

Oglofus Bangs is a high-performance, open-source alternative to DuckDuckGo's bangs. It provides lightning-fast
redirections using hashed binary search for bang commands.

## What are bangs?

Bangs are shortcuts that quickly take you to search results on other sites. For example, when you know you want to
search on another site like Wikipedia or Amazon, bangs get you there fastest. A search for `filter bubble !w` will take
you directly to Wikipedia.

## Why Oglofus Bangs?

- **Speed**: Utilizes hashed binary search for extremely fast bang lookups
- **Performance**: Built with [fasthttp](https://github.com/valyala/fasthttp) for high-throughput, low-latency HTTP
  handling
- **Lightweight**: Minimal dependencies and efficient memory usage
- **Self-hosted**: Run your own bang service without relying on third parties
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

### Using Docker

Oglofus Bangs is available as a Docker image for easy deployment. You can pull the image from GitHub Container Registry:

```bash
docker pull ghcr.io/oglofus/bangs:latest
````

Run the container:

```bash
docker run -p 8080:8080 ghcr.io/oglofus/bangs:latest
```

You can specify a custom address and default search engine:

```bash
docker run -p 9000:9000 ghcr.io/oglofus/bangs:latest -addr :9000 -default "https://duckduckgo.com/?q=<q>"
```

### Building Your Own Docker Image

If you prefer to build your own Docker image, you can use the provided Dockerfile:

```bash
git clone https://github.com/oglofus/bangs.git
cd bangs
docker build -t oglofus-bangs .
docker run -p 8080:8080 oglofus-bangs
```

The Dockerfile uses a multi-stage build process to create a minimal image:

1. Builds the application in a Golang Alpine container
2. Copies only the necessary binaries to a clean Alpine image
3. Exposes port 8080 for the service

## Installation

### Prerequisites

- Go 1.24 or higher

### Building from Source

1. Clone the repository:
    ```bash
    git clone https://github.com/oglofus/bangs.git
    cd bangs
    ```

2. Build the project:
    ```bash
    go build -o bangs
    ```

3. Run the server:
   ```bash
    ./bangs [options]
    ```

By default, the server runs on address `:8080` if no address is specified.

## Usage

Once the server is running, you can use it by sending HTTP requests with a query parameter `q` containing your search
term and bang:

http://localhost:8080/?q=filter%20bubble%20!w

This will redirect you to Wikipedia's search for "filter bubble".

**Important:** Unlike DuckDuckGo's implementation, Oglofus Bangs only recognizes bangs that appear at the end of the
query string. For example:

- ✅ `filter bubble !w` - Will work correctly
- ❌ `!w filter bubble` - Will not be recognized as a bang command

## Command-Line Options

Oglofus Bangs supports the following command-line options:

- `-addr string`: HTTP server address (default ":8080")
- `-default string`: Default search URL template (must contain `<q>` as query placeholder) (
  default "https://www.google.com/search?q=<q>")

Examples:

```bash
# Run on default port with Google as default search
./bangs

# Run on port 9000
./bangs -addr :9000

# Use DuckDuckGo as the default search engine
./bangs -default "https://duckduckgo.com/?q=<q>"

# Combine options
./bangs -addr :9000 -default "https://bing.com/search?q=<q>"
```

## Customizing Bangs

Bangs are defined in the `bangs.json` file. Each bang has a trigger (`z`) and a URL template (`u`):

```json
[
  {
    "t": "w",
    "u": "https://en.wikipedia.org/wiki/Special:Search?search=<q>"
  },
  {
    "t": "gh",
    "u": "https://github.com/search?q=<q>"
  }
]
```

After modifying the `bangs.json` file, you need to convert it to the binary format used by the application:

For Unix/Linux/macOS users:

```bash
./convert.sh
```

For other platforms:

```shell
go run ./preprocessor/main.go
```

This will generate the required `bangs.idx` and `bangs.dat` files.

## Technical Details

- The application uses SHA3-224 hashing for bang lookups
- Bang data is stored in two binary files:
    - `bangs.idx`: Contains hashed keys and offsets into the data file
    - `bangs.dat`: Contains the actual URL templates
- The `<q>` placeholder in URL templates is replaced with the user's search query
- Default search engine can be customized via command-line flag
- URL templates use `<q>` as the placeholder in configuration, which is converted to a binary placeholder internally

## Performance

Oglofus Bangs is designed for high performance:

- O(log n) lookup time for bangs
- Minimal memory footprint
- Fast HTTP handling with fasthttp
- Efficient binary data format

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
