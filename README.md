## Fleetlock client

[![Go Reference](https://pkg.go.dev/badge/github.com/flatcar-linux/fleetlock.svg)](https://pkg.go.dev/github.com/flatcar-linux/fleetlock)
[![Go](https://github.com/flatcar-linux/fleetlock/actions/workflows/go.yml/badge.svg)](https://github.com/flatcar-linux/fleetlock/actions/workflows/go.yml)

Go implementation of `FleetLock` protocol.

### Example

```
$ fleetlockctl --help
Usage:
  fleetlockctl  [flags]

Flags:
  -h, --help   help for fleetlockctl
```

### Build

requirements:
  * `go` in the path

```
$ make
```

ref: https://coreos.github.io/zincati/development/fleetlock/protocol/
