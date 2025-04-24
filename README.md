# GOAT - Golang Application Tracing

[![Go Report Card](https://goreportcard.com/badge/github.com/monshunter/goat)](https://goreportcard.com/report/github.com/monshunter/goat)
[![GoDoc](https://godoc.org/github.com/monshunter/goat?status.svg)](https://godoc.org/github.com/monshunter/goat)
[![License](https://img.shields.io/github/license/monshunter/goat)](https://github.com/monshunter/goat/blob/main/LICENSE)

[中文文档](README_ZH.md)

## 📖 Introduction

`GOAT` (Golang Application Tracing) is a high-performance code tracing tool for gray releases, designed specifically for Go applications. It automatically identifies and tracks the execution of incremental code, helping developers make more reliable decisions during the gray release process. Through automated instrumentation and real-time tracking, GOAT provides internal evidence to ensure that incremental features are thoroughly tested during the gray release process.

## 🚀 Features

* Automatically identifies effective incremental code, precisely locating modification points
* Intelligent instrumentation system, supporting tracking of explicit and implicit branches
* Provides code tracking capabilities at multiple granularity levels (line, patch, scope, function)
* Supports multiple precision modes for differences, adapting to code changes of varying complexity
* Embedded HTTP service, displaying instrumentation coverage status in real-time
* Efficient resource utilization, minimizing impact on application performance
* Simple and easy-to-use command-line tools and API interfaces
* Multi-threading support to improve processing speed
* Support for custom instrumentation strategies

## 💡 How GOAT Works

### Workflow

1. **Initialization**: Configure project parameters and generate a configuration file
2. **Difference Analysis**: Analyze code differences between stable and release branches
3. **Intelligent Instrumentation**: Automatically insert tracking code into incremental code
4. **Runtime Monitoring**: Collect instrumentation execution data during application runtime
5. **Status Display**: Display instrumentation coverage status through HTTP interface

## 🧰 Installation

### Method 1: Install using Go Install (Recommended):

```bash
go install github.com/monshunter/goat/cmd/goat@latest
```

Make sure your `$GOPATH/bin` directory is added to your system PATH.

### Method 2: Build and install from source:

```bash
git clone https://github.com/monshunter/goat.git
cd goat
make install
```

This will compile the binary and install it in the `$GOPATH/bin` directory.

### Method 3: Build without installing:

```bash
git clone https://github.com/monshunter/goat.git
cd goat
make build
```

The built binary will be in the `bin` directory.

## 🛠 Usage

### Initialize Project

Execute in the root directory of your Go project:

```bash
goat init
```

This will generate the default configuration file `goat.yaml`. You can customize configuration options:

```bash
goat init --old main --new HEAD --app-name "my-app" --granularity func
```

### Configuration Options

Customize configuration by using various options when calling `goat init`:

```bash
goat init --help
```

Common options include:
- `--old <oldBranch>`: Stable branch (default: "main")
- `--new <newBranch>`: Release branch (default: "HEAD")
- `--app-name <appName>`: Application name (default: "example-app")
- `--granularity <granularity>`: Granularity (line, patch, scope, func) (default: "patch")
- `--diff-precision <diffPrecision>`: Difference precision (1~3) (default: 1)
- `--threads <threads>`: Number of threads (default: 1)
- `--ignores <ignores>`: List of files/directories to ignore, comma-separated

### Insert Tracking Code

```bash
goat track
```

This will analyze incremental code in the project and automatically insert tracking instrumentation. After running this command, you can:
- Use git diff or other tools to view changes
- Build and test your application to verify instrumentation
- If the project already has tracking code, run `goat clean` first

### Handle Manual Instrumentation Markers

```bash
goat patch
```

This command is used to process manual instrumentation markers in the project, mainly handling:
- `// +goat:delete` markers - Delete code segments marked for deletion
- `// +goat:insert` markers - Insert code at marked positions
If you have manually added or removed instrumentation, you can run this command to update the instrumentation implementation.

### Clean Tracking Code

```bash
goat clean
```

Remove all inserted tracking code.

### View Version Information

```bash
goat --version
```

## 📚 Examples

GOAT provides several detailed usage examples to help you better understand its features:

1. [Track Command Example](examples/track_example.md) - How to track code changes and insert tracing code
2. [Patch Command Example](examples/patch_example.md) - How to process manual tracing markers
3. [Clean Command Example](examples/clean_example.md) - How to clean up tracing code
4. [Granularity Examples](examples/granularity_example.md) - Demonstrates tracing at different granularities

For more examples, check the [examples directory](examples/).

## 🖥 Application Scenarios

- Code coverage tracking in gray releases (blue-green deployment, canary release)
- Execution path monitoring for new features
- Validation testing for refactored code
- Impact analysis of performance changes
- Service upgrade tracking in microservice architectures

## 🔋 Development Environment Requirements

- Go 1.21+
- Git

## 📊 Instrumentation Data Monitoring

### HTTP Service

After inserting instrumentation code with GOAT, an HTTP service will automatically start when your application runs, providing real-time instrumentation coverage status. By default, this service runs on port `57005`.

You can customize the port by setting the environment variable `GOAT_PORT`:

```bash
export GOAT_PORT=8080
```

### API Endpoints

GOAT provides the following API endpoints for querying instrumentation coverage status:

#### 1. Get metrics in Prometheus format

```
GET http://127.0.0.1:57005/metrics
```

#### 2. Get Instrumentation Status for All Components

```
GET http://localhost:57005/track
```

#### 3. Get Instrumentation Status for a Specific Component

```
GET http://localhost:57005/track?component=COMPONENT_ID
```

Where `COMPONENT_ID` is the component's ID (usually an integer starting from 0) or the component name.

#### 4. Get Instrumentation Status for Multiple Components

```
GET http://localhost:57005/track?component=COMPONENT_ID1,COMPONENT_ID2
```

#### 5. Sort Results in Different Orders

```
# Sort by execution count in ascending order
GET http://localhost:57005/track?component=COMPONENT_ID&order=0

# Sort by execution count in descending order
GET http://localhost:57005/track?component=COMPONENT_ID&order=1

# Sort by ID in ascending order
GET http://localhost:57005/track?component=COMPONENT_ID&order=2

# Sort by ID in descending order
GET http://localhost:57005/track?component=COMPONENT_ID&order=3
```

### Response Format

The /metrics API returns the standard format of Prometheus:

```
# HELP goat_track_total goat track total
# TYPE goat_track_total gauge
goat_track_total{app="calculator",version="cadafce",component="."} 16
# HELP goat_track_covered goat track covered
# TYPE goat_track_covered gauge
goat_track_covered{app="calculator",version="cadafce",component="."} 5
# HELP goat_track_coverage_ratio goat track coverage ratio
# TYPE goat_track_coverage_ratio gauge
goat_track_coverage_ratio{app="calculator",version="cadafce",component="."} 31
```

The /track API returns responses in JSON format, containing the following information:

```json
{
  "name": "example-app", 
  "version": "1.0.0",
  "results": [
    {
      "id": 0, // component id
      "name": "ComponentName", // component name
      "metrics": {
        "version":"d6f985a397eb7ba24b877cc67a14b663",  // version of this metrics
        "total": 10,         // Total number of instrumentation points
        "covered": 5,         // Number of covered instrumentation points
        "coveredRate": 50,    // Coverage rate (percentage)
        "items": [
          {
            "id": 1,          // Instrumentation ID
            "name": "TRACK_ID_1", // Instrumentation name
            "count": 3        // Execution count
          }
          // More instrumentation points...
        ]
      }
    }
    // More components...
  ]
}
```

### Usage Examples

1. View all instrumentation status using curl:

```bash
curl http://localhost:57005/track | jq
```

2. View instrumentation status for a specific component using curl:

```bash
curl http://localhost:57005/track?component=0 | jq
```

### Observation and Analysis

1. **Real-time Monitoring**: View instrumentation coverage at any time while the application is running
2. **Gray Release Decision**: Evaluate whether to proceed with the next step of gray release based on instrumentation coverage
3. **Problem Analysis**: Identify code paths that have not been executed, locate potential issues
4. **Coverage Reporting**: Generate coverage reports for team review and quality assurance

## 🌐 Environment Variables

The GOAT project supports configuration through environment variables. The following table lists all available environment variables and their functions:

| Environment Variable | Description | Default Value | Use Case |
| --- | --- | --- | --- |
| `GOAT_PORT` | Sets the port for the instrumentation HTTP service | `57005` | When the default port is occupied or a custom port is needed |
| `GOAT_METRICS_IP` | Sets the IP address that the instrumentation HTTP service binds to | `127.0.0.1` | When access from non-local machines is needed, can be set to `0.0.0.0` |
| `GOAT_CONFIG` | Specifies the path to the configuration file | `goat.yaml` | When a non-default location for the configuration file is needed |
| `GOAT_CURRENT_COMPONENT` | Specifies the name of the current component | `""` | If not specified, `/metrics` will return the metrics of all components |
| `GOAT_STACK_TRACE` | Whether to display stack traces when fatal errors occur | `false` | Set to `1` or `true` or `yes` when debugging issues |

### Examples of Using Environment Variables

1. Modify HTTP service port:

```bash
export GOAT_PORT=8080
```

2. Allow access to the instrumentation service from other machines:

```bash
export GOAT_METRICS_IP=0.0.0.0
```

3. Use a custom configuration file path:

```bash
export GOAT_CONFIG=/path/to/custom-goat.yaml
```

4. Enable error stack tracing:

```bash
export GOAT_STACK_TRACE=1
```

5. Component name for the /metrics data:

```bash
export GOAT_CURRENT_COMPONENT="cmd/echo"
```

## 🏷️ Marker Explanation

GOAT uses special code comment markers (markers starting with `// +goat:`) to control code insertion and deletion. There are two types of markers: user-available markers and internal markers.

### User-Available Markers

The following markers can be used by developers:

| Marker | Description | Use Case | Status |
| --- | --- | --- | --- |
| `// +goat:delete` | Marks the beginning of a code block that needs to be deleted | Used when a segment of code needs to be deleted | Enabled |
| `// +goat:insert` | Marks a position where code needs to be inserted | Manually specify instrumentation insertion position | Enabled |

### Internal Markers

The following markers are used internally by GOAT and should not be manually added by users:

| Marker | Description | Use Case | Status |
| --- | --- | --- | --- |
| `// +goat:generate` | Marks the beginning of instrumentation code generation | Beginning marker for automatically generated instrumentation code blocks | Enabled |
| `// +goat:tips: ...` | Tip information | Provides tips to developers about the code block | Enabled |
| `// +goat:main` | Marks the main function entry point instrumentation | Adds HTTP service startup code in the main function | Enabled |
| `// +goat:end` | Marks the end of a code block | End marker for all `+goat:` marked blocks | Enabled |
| `// +goat:import` | Marks the import section | Used to mark import statements related to instrumentation | Not Enabled |
| `// +goat:user` | Marks user-defined instrumentation | User-defined instrumentation code | Not yet supported |

### Notes

1. **Deleting Code Blocks**: If you change `// +goat:generate` to `// +goat:delete`, and then execute the `goat patch` command, the code between `// +goat:delete` and `// +goat:end` will be deleted.

   ```go
   // +goat:delete
   // +goat:tips: do not edit the block between the +goat comments
   goat.Track(goat.TRACK_ID_1)
   // +goat:end
   ```

2. **Inserting Instrumentation**: Add the following at the position where you want to manually insert instrumentation:

   ```go
   // +goat:insert
   ```

   After executing `goat patch`, instrumentation code will be inserted at that position.

3. **When Markers Take Effect**: These markers are processed when executing the `goat patch` command, not the `goat track` command.

4. **Marker Nesting**: Markers do not support nesting. Each marker block must be completely ended with `// +goat:end`.

## 📄 License

The source code of `GOAT` is open-sourced under the [MIT License](LICENSE).

## 💎 Contribution

Contributions of code or suggestions are welcome! Please check the [Contribution Guidelines](CONTRIBUTING.md) for more information.

## ☕️ Support

If you find GOAT helpful, you can support the project in the following ways:

- Star the project on GitHub
- Submit Pull Requests to add new features or fix bugs
- Recommend this project to others
