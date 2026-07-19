# tifybe-cli

[![Release](https://img.shields.io/github/v/release/emirhannsarial/tifybe-cli)](https://github.com/emirhannsarial/tifybe-cli/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/emirhannsarial/tifybe-cli)](https://goreportcard.com/report/github.com/emirhannsarial/tifybe-cli)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Receive and inspect webhooks on `localhost` through a secure outbound tunnel.

`tifybe` gives you a public HTTPS URL that forwards every request to a port on
your machine — so Stripe, GitHub or any webhook provider can reach your local
dev server without port forwarding, reverse proxies or firewall changes. It is
the open-source companion CLI to [Tifybe](https://tifybe.com), and works
without an account.

```
$ tifybe listen 8080
tifybe v1.1.0

  Forwarding   https://api.tifybe.com/local/req_a1b2c3d4 -> http://localhost:8080
  Inspector    https://tifybe.com/local/req_a1b2c3d4

23:14:02  connected — waiting for webhooks
23:14:31  POST -> http://localhost:8080  200 (12ms)
23:15:07  POST -> http://localhost:8080  500 (3ms)
```

Every request also appears in a live web inspector (the `Inspector` URL), where
you can read headers and payloads as they arrive.

## Install

**Go 1.22+**

```bash
go install github.com/emirhannsarial/tifybe-cli/cmd/tifybe@latest
```

**Binaries** — pre-built for macOS, Windows and Linux (amd64/arm64) on the
[releases page](https://github.com/emirhannsarial/tifybe-cli/releases).

## Usage

### Anonymous session

```bash
tifybe listen 8080
```

Prints a randomly generated public URL. Paste it wherever a provider asks for a
webhook URL. The URL lives as long as the session; no account needed.

### Persistent URL

A random URL changes on every restart. With a free Tifybe account you can claim
a subdomain that stays yours:

```bash
tifybe login                                # paste your API key (tfy_…)
tifybe listen 8080 --subdomain=my-startup   # → …/local/my-startup, forever
```

Configure it once in your provider's dashboard and never touch it again.

### Commands & flags

| Command | Description |
|---|---|
| `tifybe listen <port>` | Start a tunnel to `localhost:<port>` |
| `tifybe login` | Store your API key in `~/.tifybe/credentials.json` (mode `0600`) |
| `tifybe logout` | Delete stored credentials |
| `tifybe --version` | Print the version |

`listen` flags: `--subdomain` (persistent URL, requires login),
`--backend-url`, `--frontend-url` (self-hosted / testing overrides).

## How it works

1. The CLI opens an **outbound** WebSocket connection to the Tifybe edge —
   nothing on your machine listens publicly.
2. A provider POSTs to your public URL; the edge serializes the method, headers
   and body and streams the frame down the WebSocket.
3. The CLI replays the request against `localhost:<port>` and logs the local
   server's status code and response time.

If the connection drops, the CLI reconnects automatically with exponential
backoff (1s → 30s). `Ctrl-C` closes the session cleanly so the URL is released
immediately.

## Security notes

- The tunnel is outbound-only; your machine accepts no inbound connections.
- Credentials are stored locally in `~/.tifybe/credentials.json` with `0600`
  permissions, and only ever sent as a bearer token over TLS.
- Anonymous URLs contain 64 bits of randomness and are not enumerable.

## Development

```bash
go build ./...   # build
go test ./...    # run tests
```

Bug reports and PRs are welcome — please run `go fmt` and make sure tests pass.

## License

MIT — see [LICENSE](LICENSE).
