# Jrpc2 NATS Channel

![Build](https://github.com/41north/go-natschannel/actions/workflows/ci.yml/badge.svg)
[![Coverage Status](https://coveralls.io/repos/github/41north/go-natschannel/badge.svg)](https://coveralls.io/github/41north/go-natschannel)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Status: _EXPERIMENTAL_

Package `natschannel` provides an implementation of the
[`jrpc2`](https://godoc.org/github.com/creachadair/jrpc2) module's
[`Channel`](https://godoc.org/github.com/creachadair/jrpc2/channel#Channel)
interface using [NATS](https://nats.io) as the transport.

## Documentation

[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/41north/go-natschannel)

Full `go doc` style documentation for the project can be viewed online without
installing this package by using the excellent GoDoc site here:
http://godoc.org/github.com/41north/go-natschannel

You can also view the documentation locally once the package is installed with
the `godoc` tool by running `godoc -http=":6060"` and pointing your browser to
http://localhost:6060/pkg/github.com/41north/go-natschannel

## Installation

```bash
$ go get -u github.com/41north/go-natschannel
```

Add this import line to the file you're working in:

```Go
import natschannel "github.com/41north/go-natschannel"
```

## Quick Start

```go
// connects to the nats server and binds the channel to the 'foo.bar' subject
channel, err := natschannel.Dial("nats://localhost:4222", "foo.bar")

// wraps an existing nats connection and binds the channel to the 'foo.bar' subject
channel, err := natschannel.New(conn, "foo.bar")
```

## License

go-natschannel is licensed under the [Apache 2.0 License](LICENSE)

## Contact

If you want to get in touch drop us an email at [hello@41north.dev](mailto:hello@41north.dev)
