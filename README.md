# Tifybe CLI

[![Go Report Card](https://goreportcard.com/badge/github.com/emirhannsarial/tifybe-cli)](https://goreportcard.com/report/github.com/emirhannsarial/tifybe-cli)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

Tifybe CLI is a fast, lightweight, and open-source command-line tool built in Go that allows developers to securely receive and inspect webhooks on their local machine.

Gone are the days of setting up complex reverse proxies or wrestling with NAT configurations. Tifybe establishes a secure, real-time WebSocket tunnel to forward webhooks directly to your `localhost` with zero friction.

## Features

- **Zero-Config Forwarding:** Start receiving webhooks instantly. No account or configuration required.
- **Real-Time Web Viewer:** Inspect headers, payloads, and request paths instantly via a sleek web UI.
- **Persistent Subdomains:** Claim your own custom, persistent URLs (e.g., `my-startup.tifybe.com`) so you never have to update your Stripe or GitHub webhook settings again.
- **High Performance:** Written in Go, ensuring minimal memory footprint and sub-millisecond latency tunneling.
- **Cross-Platform:** Available as a standalone binary for macOS, Windows, and Linux.

## Installation

### Using Go (Recommended)
If you have Go 1.22+ installed, you can install the CLI directly:

```bash
go install github.com/emirhannsarial/tifybe-cli@latest
```

### Download Binaries
Pre-compiled binaries for all major platforms are available on the [Releases](https://github.com/emirhannsarial/tifybe-cli/releases) page.

## Quick Start

### 1. Anonymous Mode (No Account Required)
To instantly forward incoming webhooks to a local server running on port `8080`:

```bash
tifybe listen 8080
```

**Output:**
```text
Tifybe CLI v1.0.0
────────────────────────────────────────────────────────
Forwarding:      https://api.tifybe.com/local/req_a1b2c3d4  ->  http://localhost:8080
Web Interface:   https://tifybe.com/local/req_a1b2c3d4
Status:          🟢 Online & Listening

Waiting for webhooks...
────────────────────────────────────────────────────────
```

You can now copy the **Forwarding** URL and paste it into Stripe, GitHub, or any other third-party service. Open the **Web Interface** URL in your browser to inspect the payloads as they arrive.

### 2. Authenticated Mode (Persistent URLs)
Tired of your webhook URL changing every time you restart your terminal? You can claim a persistent subdomain by linking the CLI to your free Tifybe account.

**Login:**
```bash
tifybe login
```
*(You will be prompted to enter your API key, which can be found in your Tifybe Dashboard).*

**Listen on a Custom Subdomain:**
```bash
tifybe listen 8080 --subdomain=my-startup
```
Now, all webhooks sent to `https://api.tifybe.com/local/my-startup` will be reliably routed to your machine.

## How It Works

Under the hood, `tifybe-cli` establishes a multiplexed, outbound WebSocket connection (`wss://`) to the Tifybe Edge infrastructure. 

1. External services send a standard `HTTP POST` request to your unique Tifybe ingress URL.
2. The Edge infrastructure securely serializes the HTTP method, headers, and body.
3. The payload is streamed down your active WebSocket connection.
4. The CLI reconstructs the request and forwards it to your specified `localhost` port.

This architecture ensures that your local environment remains completely shielded from the public internet, mitigating the risk of inbound attacks.

## Contributing

We welcome contributions from the community! If you've found a bug, have a feature request, or want to contribute code:

1. Fork the repository.
2. Create your feature branch (`git checkout -b feature/amazing-feature`).
3. Commit your changes (`git commit -m 'feat: add amazing feature'`).
4. Push to the branch (`git push origin feature/amazing-feature`).
5. Open a Pull Request.

Please ensure all tests pass (`go test ./...`) and that your code adheres to standard Go formatting guidelines (`go fmt`).

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.