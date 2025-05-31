# GOAT - Golang Application Tracing

[![Go Report Card](https://goreportcard.com/badge/github.com/monshunter/goat)](https://goreportcard.com/report/github.com/monshunter/goat)
[![GoDoc](https://godoc.org/github.com/monshunter/goat?status.svg)](https://godoc.org/github.com/monshunter/goat)
[![License](https://img.shields.io/github/license/monshunter/goat)](https://github.com/monshunter/goat/blob/main/LICENSE)

[中文文档](README_ZH.md) [Wiki](https://deepwiki.com/monshunter/goat)

`GOAT` (Golang Application Tracing) is a high-performance code tracing tool that automatically tracks incremental code execution during gray releases.

## Architecture

![GOAT Architecture](docs/images/goat-architecture.svg)

## Features

- Auto-detects Git branch changes
- Real-time coverage dashboard
- Minimal manual setup, clean removal

## Quick Start

### Step 1: Install
```bash
go install github.com/monshunter/goat/cmd/goat@latest
```

### Step 2: Config In Project

```bash
goat init
```
### Step 3: Insert Tracking Code

```bash
goat track
```

### Step 4: Run and Monitor

- (Opt 1) prometheus format

```
GET http://127.0.0.1:57005/metrics
```

- (Opt 2) json format

```
GET http://localhost:57005/track
```

### At the End: Clean Up

```bash
goat clean
```

## Examples & Documentation

- [examples directory](examples/README.md).
- [architecture document](docs/technical-architecture.md).

## License

[MIT License](LICENSE).

## Support

If you find GOAT helpful, Star & Share & Contribute!
