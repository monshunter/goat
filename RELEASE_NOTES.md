# GOAT v1.0.0 Release Notes

We're excited to announce the first official release of GOAT (Golang Application Tracing), a specialized tool designed to enhance the reliability of gray releases in Go applications.

## Overview

GOAT is a high-performance gray release code tracing tool specifically designed for Go language applications. It automatically identifies and tracks the execution of incremental code, helping developers make more reliable decisions during gray release processes. Through automated instrumentation and real-time tracking, GOAT provides internal evidence to ensure incremental features are thoroughly tested during the gray release process.

## Key Features

- **Automatic Incremental Code Detection**: Precisely identifies modification points in your code
- **Intelligent Instrumentation System**: Tracks both explicit and implicit branches
- **Multiple Granularity Levels**: Supports line-level, patch-level, scope-level, and function-level code tracking
- **Multiple Precision Modes**: Adapts to different complexity levels of code changes
- **Embedded HTTP Service**: Real-time display of instrumentation coverage status
- **Resource Efficiency**: Minimizes impact on application performance
- **Simple CLI and API**: Easy-to-use command line tools and API interfaces
- **Multi-threading Support**: Enhances processing speed
- **Custom Instrumentation Strategies**: Supports customized tracking approaches

## Installation

### Method 1: Using Go Install (Recommended)

```bash
go install github.com/monshunter/goat/cmd/goat@latest
```

### Method 2: Download Pre-built Binaries

Pre-built binaries for various platforms are available in the GitHub release assets.

## Basic Usage

### Initialize Project

```bash
goat init
```

This generates the default configuration file `goat.yaml`. You can customize configuration options:

```bash
goat init --old main --new HEAD --app-name "my-app" --granularity func
```

### Insert Tracking Code

```bash
goat track
```

This analyzes the differences between branches and inserts tracking code.

### Clean Up Tracking Code

```bash
goat clean
```

This removes all tracking code from your project.

## Use Cases

- Code coverage tracking in gray releases (blue-green deployment, canary release)
- Execution path monitoring for new features
- Validation testing for refactored code
- Impact analysis of performance changes
- Service upgrade tracking in microservice architectures

## Requirements

- Go 1.23+
- Git

## Documentation

For detailed documentation, please refer to:
- [Technical Guide](docs/technical-guide.md)
- [Technical Architecture](docs/technical-architecture.md)

## Known Limitations

- GOAT is specifically designed for Go applications and cannot be used with other languages
- Requires Go 1.23+ for full functionality
- Relies on Git for diff analysis
- Requires a valid Git repository with commit history
- May not correctly handle extremely complex code structures

## Acknowledgements

We would like to thank all contributors who have helped make this release possible.

## License

GOAT is released under the MIT License. See the LICENSE file for details.
