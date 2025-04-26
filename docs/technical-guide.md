# GOAT Technical Guide

## Overview

GOAT (Golang Application Tracing) is a specialized tool designed to enhance the reliability of gray releases in Go applications. This technical guide explains how to use GOAT effectively, covering installation, configuration, and practical usage scenarios.

## Technical Concepts

### Gray Release

A gray release (also known as canary release or phased rollout) is a deployment strategy where new code is gradually rolled out to a subset of users before being deployed to the entire user base. This approach helps identify potential issues before they affect all users.

### Code Tracing

Code tracing in the context of GOAT refers to the process of instrumenting code to track its execution. This instrumentation allows developers to verify that new or modified code paths are being executed as expected during the gray release process.

### Instrumentation

Instrumentation is the process of adding tracking code to an application. GOAT automatically inserts tracking statements at strategic points in the code, allowing it to monitor which parts of the code are executed at runtime.

## Technical Architecture

GOAT consists of several key components:

1. **Diff Analysis Engine**: Identifies code changes between branches using Git
2. **Instrumentation System**: Inserts tracking code into the application
3. **Runtime Monitoring**: Collects and displays execution data during runtime
4. **HTTP Service**: Provides real-time visibility into instrumentation coverage

## Installation

### Prerequisites

- Go 1.21 or higher
- Git
- A Go project with a valid Git repository

### Installation Methods

#### Method 1: Using Go Install (Recommended)

```bash
go install github.com/monshunter/goat/cmd/goat@latest
```

Ensure your `$GOPATH/bin` directory is in your system PATH.

#### Method 2: Build and Install from Source

```bash
git clone https://github.com/monshunter/goat.git
cd goat
make install
```

#### Method 3: Build Without Installing

```bash
git clone https://github.com/monshunter/goat.git
cd goat
make build
```

The built binary will be in the `bin` directory.

## Technical Configuration

### Configuration File

GOAT uses a YAML configuration file (`goat.yaml`) to control its behavior. The configuration file includes the following key parameters:

```yaml
# App name
appName: example-app

# App version
appVersion: 1.0.0

# Old branch name (stable branch)
oldBranch: main

# New branch name (release branch)
newBranch: HEAD

# Files or directories to ignore
ignores:
  - .git
  - vendor
  - testdata

# Goat package name
goatPackageName: goat

# Goat package alias
goatPackageAlias: goat

# Goat package path
goatPackagePath: goat

# Granularity (line, patch, scope, func)
granularity: patch

# Diff precision (1~3)
diffPrecision: 1

# Threads
threads: 4

# Race condition protection
race: true

# Main packages to track
mainEntries:
  - "*"
```

### Granularity Levels

GOAT supports four levels of tracking granularity:

1. **Line Granularity (`line`)**: Tracks changes at the line level, providing the most detailed tracking but with the highest number of tracking points.

2. **Patch Granularity (`patch`)**: Tracks changes at the patch level (continuous modification blocks within the same scope). This is the default granularity level and provides a good balance between detail and performance.

3. **Scope Granularity (`scope`)**: Tracks changes at the scope level (if blocks, for loops, etc.), resulting in fewer tracking points and less performance impact.

4. **Function Granularity (`func`)**: Tracks changes at the function level, providing the coarsest tracking with minimal performance impact.

### Diff Precision Modes

GOAT offers three precision modes for diff analysis:

1. **Precision Level 1**: Uses `git blame` for high-precision analysis but with slower performance.

2. **Precision Level 2**: Uses `git diff` with the ability to track file renames (some renamed files may be treated as new files), offering a balance between precision and performance.

3. **Precision Level 3**: Uses `git diff` without tracking file renames (all renamed files are treated as new files), providing the fastest performance but with lower precision.

## Technical Usage

### Workflow

The typical workflow for using GOAT consists of the following steps:

1. **Initialize**: Configure project parameters and generate a configuration file
2. **Analyze Differences**: Analyze code differences between stable and release branches
3. **Insert Instrumentation**: Automatically insert tracking code into incremental code
4. **Monitor Execution**: Collect instrumentation execution data during application runtime
5. **View Coverage**: Display instrumentation coverage status through the HTTP interface

### Commands

#### Initialize Project

```bash
goat init
```

This generates the default configuration file `goat.yaml`. You can customize configuration options:

```bash
goat init --old main --new HEAD --app-name "my-app" --granularity func
```

#### Insert Tracking Code

```bash
goat track
```

This analyzes the project's incremental code and automatically inserts tracking points.

#### Process Manual Tracking Markers

```bash
goat patch
```

This processes any manual tracking markers in the code.

#### Clean Up Tracking Code

```bash
goat clean
```

This removes all inserted tracking code from the project.

### Runtime Monitoring

After inserting instrumentation code with GOAT, an HTTP service will automatically start when your application runs, providing real-time instrumentation coverage status. By default, this service runs on port `57005`.

You can customize the port by setting the environment variable `GOAT_PORT`:

```bash
export GOAT_PORT=8080
```

#### API Endpoints

GOAT provides the following API endpoints for querying instrumentation coverage status:

1. **Get metrics in Prometheus format**:
   ```
   GET http://127.0.0.1:57005/metrics
   ```

2. **Get Instrumentation Status for All Components**:
   ```
   GET http://localhost:57005/track
   ```

3. **Get Instrumentation Status for a Specific Component**:
   ```
   GET http://localhost:57005/track?component=COMPONENT_ID
   ```

4. **Get Instrumentation Status for Multiple Components**:
   ```
   GET http://localhost:57005/track?component=COMPONENT_ID1,COMPONENT_ID2
   ```

5. **Sort Results in Different Orders**:
   ```
   # Sort by execution count (ascending)
   GET http://localhost:57005/track?component=COMPONENT_ID&order=0

   # Sort by execution count (descending)
   GET http://localhost:57005/track?component=COMPONENT_ID&order=1

   # Sort by ID (ascending)
   GET http://localhost:57005/track?component=COMPONENT_ID&order=2

   # Sort by ID (descending)
   GET http://localhost:57005/track?component=COMPONENT_ID&order=3
   ```

## Technical Implementation Details

### Tracking Code Structure

The tracking code inserted by GOAT consists of the following components:

1. **Import Statement**:
   ```go
   import goat "go-module/goat" 
   ```

2. **Tracking Call**:
   ```go
   // +goat:generate
   // +goat:tips: do not edit the block between the +goat comments
   goat.Track(goat.TRACK_ID_X)
   // +goat:end
   ```

3. **HTTP Service Initialization** (in main function):
   ```go
   // +goat:main
   // +goat:tips: do not edit the block between the +goat comments
   goat.ServeHTTP(goat.COMPONENT_Y)
   // +goat:end
   ```

### Special Comment Markers

GOAT uses special comment markers to control code insertion and deletion:

| Marker | Description | Use Case |
| --- | --- | --- |
| `// +goat:generate` | Marks the beginning of instrumentation code generation | Beginning marker for automatically generated instrumentation code blocks |
| `// +goat:tips: ...` | Tip information | Provides tips to developers about the code block |
| `// +goat:main` | Marks the main function entry point instrumentation | Adds HTTP service startup code in the main function |
| `// +goat:end` | Marks the end of a code block | End marker for all `+goat:` marked blocks |
| `// +goat:delete` | Marks code to be deleted | Used when code needs to be removed |
| `// +goat:insert` | Marks insertion points | Used to manually specify instrumentation insertion points |

## Technical Best Practices

### Choosing the Right Granularity

- Use **line granularity** when you need extremely detailed tracking and performance is not a concern
- Use **patch granularity** (default) for most gray release scenarios
- Use **scope granularity** when you have scattered but logically related code changes
- Use **function granularity** when you only need high-level tracking of function execution

### Optimizing Performance

- Use a higher diff precision level (2 or 3) for large codebases to improve analysis speed
- Increase the `threads` parameter on multi-core systems to parallelize processing
- Use coarser granularity levels (scope or func) for performance-critical applications

### Integration with CI/CD Pipelines

GOAT can be integrated into CI/CD pipelines to automate the instrumentation process:

1. Add a step to run `goat track` after code changes are merged to the release branch
2. Deploy the instrumented application to the gray release environment
3. Monitor the instrumentation coverage during the gray release period
4. Or run `goat clean` before fully deploying to production

## Technical Troubleshooting

### Common Issues

1. **No Main Package Found**:
   - Ensure your project has at least one `main` package
   - Check that the main package is not in an ignored directory

2. **Instrumentation Not Working**:
   - Verify that the tracking code has been properly inserted
   - Check that the HTTP service is running on the expected port
   - Ensure the application has permission to bind to the specified port

3. **Performance Issues**:
   - Try using a coarser granularity level
   - Reduce the diff precision level
   - Increase the number of threads for parallel processing

### Debugging

GOAT provides verbose output to help diagnose issues:

```bash
goat track --verbose
```

This will display detailed information about the tracking process, including:
- Files being analyzed
- Changes detected
- Instrumentation points inserted

## Advanced Technical Topics

### Custom Instrumentation

GOAT supports custom instrumentation through manual markers:

1. Add a `// +goat:insert` comment at the location where you want to insert tracking code
2. Add a `// +goat:delete` comment to remove existing tracking code
3. Run `goat patch` to process the manual markers

### Multiple Main Entries

For projects with multiple main packages, you can specify which ones to instrument:

```yaml
mainEntries:
  - "cmd/server"
  - "cmd/client"
```

### Race Condition Protection

GOAT can use atomic operations to ensure thread safety in concurrent environments, tipically when `dataType: 2` is chosen:

```yaml
race: true
```

This is particularly important for applications with high concurrency, but may have a small performance impact.

## Conclusion

GOAT provides a powerful solution for tracking code execution in gray release scenarios. By understanding its technical principles and using it effectively, developers can ensure that incremental code changes are thoroughly tested before being deployed to all users.
