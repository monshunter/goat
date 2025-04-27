# GOAT - Technical Architecture and Principles

## 1. Introduction

GOAT (Golang Application Tracing) is a high-performance code tracing tool designed specifically for Go applications in gray release scenarios. This document provides a detailed explanation of GOAT's technical architecture, core components, and underlying principles.

## 2. System Architecture

GOAT follows a modular architecture with several key components working together to provide code tracing capabilities:

```
┌─────────────────────────────────────────────────────────────┐
│                      GOAT Architecture                      │
├─────────────┬─────────────────────────────┬─────────────────┤
│             │                             │                 │
│  CLI Layer  │     Core Processing         │  Runtime        │
│             │                             │  Components     │
│ ┌─────────┐ │  ┌─────────┐   ┌──────────┐ │  ┌─────────┐    │
│ │ Command │ │  │  Diff   │   │Tracking  │ │  │ HTTP    │    │
│ │ Parser  │ │  │ Analysis│──▶│ System   │ │  │ Service │    │
│ └─────────┘ │  └─────────┘   └──────────┘ │  └─────────┘    │
│             │        │            │       │       ▲         │
│ ┌─────────┐ │        │            │       │       │         │
│ │ Config  │ │        │            │       │       │         │
│ │ Manager │ │        ▼            ▼       │       │         │
│ └─────────┘ │  ┌─────────────────────┐    │       │         │
│             │  │    Code Insertion   │    │       │         │
│             │  │      System         │────┼───────┘         │
│             │  └─────────────────────┘    │                 │
└─────────────┴─────────────────────────────┴─────────────────┘
```

### 2.1 Key Components

1. **CLI Layer**: Handles user commands and configuration management
2. **Diff Analysis**: Analyzes code differences between branches
3. **Tracking System**: Manages the tracking logic and instrumentation
4. **Code Insertion System**: Inserts tracking code into the application
5. **HTTP Service**: Provides runtime monitoring of instrumentation coverage

## 3. Core Technical Principles

### 3.1 Diff Analysis System

The diff analysis system is responsible for identifying code changes between two branches (typically stable and release branches). GOAT supports three different precision modes for diff analysis:

1. **Precision Level 1 (High Precision)**:
   - Uses `git blame` to get file change history
   - Highest precision but worst performance
   - Implemented in `DifferV1`

2. **Precision Level 2 (Medium Precision)**:
   - Uses `git diff` to get file change history
   - Tracks file renames or moves (some renamed files may be treated as new files)
   - Medium precision with normal performance (approximately 100 times faster than Level 1)
   - Implemented in `DifferV2`

3. **Precision Level 3 (Low Precision)**:
   - Uses `git diff` to get file change history
   - Cannot track file renames or moves (all renamed files are treated as new files)
   - Lowest precision but best performance (1-10 times faster than Level 2)
   - Implemented in `DifferV3`

The diff analysis process works as follows:

1. Resolve the references to the old and new branches
2. Verify that the old branch is an ancestor of the new branch
3. Compare the trees of both commits to identify changes
4. Process changes concurrently using a worker pool for better performance
5. Generate `FileChange` objects containing information about modified files and line changes

### 3.2 Tracking Granularity System

GOAT supports four levels of tracking granularity, allowing developers to choose the appropriate level of detail for their specific use case:

1. **Line Granularity (`line`)**:
   - Tracks changes at the line level
   - Most detailed tracking with instrumentation for each modified line
   - Highest number of tracking points
   - Suitable for scenarios requiring very detailed tracking
   - May have a greater impact on performance due to the large number of tracking points

2. **Patch Granularity (`patch`)**:
   - Tracks changes at the patch level (continuous modification blocks within the same scope)
   - Balanced approach with moderate number of tracking points
   - Default granularity level
   - Suitable for most gray release scenarios

3. **Scope Granularity (`scope`)**:
   - Tracks changes at the scope level (if blocks, for loops, etc.)
   - Coarser tracking with fewer tracking points
   - Tracks entire statement blocks
   - Suitable for code with scattered but logically related modifications

4. **Function Granularity (`func`)**:
   - Tracks changes at the function level
   - Coarsest tracking with minimal tracking points
   - Minimal performance impact
   - Suitable for high-level tracking of function execution

The granularity system uses AST (Abstract Syntax Tree) analysis to identify the appropriate insertion points based on the selected granularity level.

### 3.3 Instrumentation System

The instrumentation system is responsible for inserting tracking code into the application. It uses a template-based approach to generate the necessary code:

1. **Code Generation**:
   - Uses Go templates to generate tracking code
   - Supports customization of tracking code format
   - Generates unique tracking IDs for each instrumentation point

2. **Code Insertion**:
   - Inserts tracking code at the appropriate locations based on granularity
   - Uses special comment markers to identify instrumentation blocks
   - Ensures proper import statements are added

3. **Main Entry Point Instrumentation**:
   - Automatically identifies main functions in the application
   - Inserts HTTP service startup code in the main function

The instrumentation system uses the following special comment markers:

| Marker | Description | Use Case |
| --- | --- | --- |
| `// +goat:generate` | Marks the beginning of instrumentation code generation | Beginning marker for automatically generated instrumentation code blocks |
| `// +goat:tips: ...` | Tip information | Provides tips to developers about the code block |
| `// +goat:main` | Marks the main function entry point instrumentation | Adds HTTP service startup code in the main function |
| `// +goat:end` | Marks the end of a code block | End marker for all `+goat:` marked blocks |
| `// +goat:delete` | Marks code to be deleted | Used when inserted code needs to be removed |
| `// +goat:insert` | Marks insertion points | Used to manually specify instrumentation insertion points |

### 3.4 Runtime Monitoring System

The runtime monitoring system provides real-time visibility into the execution of instrumented code:

1. **HTTP Service**:
   - Automatically starts when the instrumented application runs
   - Runs on port 57005 by default (configurable via `GOAT_PORT` environment variable)
   - Provides endpoints for querying instrumentation coverage status

2. **Tracking Data Collection**:
   - Collects data on which instrumentation points have been executed
   - Supports atomic operations for thread safety in concurrent environments
   - Minimal performance impact on the application

3. **API Endpoints**:
   - `/metrics`: Returns metrics in Prometheus format
   - `/track`: Returns instrumentation status for all components
   - `/track?component=COMPONENT_ID`: Returns status for a specific component
   - `/track?component=COMPONENT_ID1,COMPONENT_ID2`: Returns status for multiple components
   - Supports various sorting options for results

## 4. Implementation Details

### 4.1 Core Data Structures

1. **Values Structure**:
   ```go
   type Values struct {
       PackageName string      // Generated code package name
       Version     string      // Application version
       Name        string      // Application name
       Components  []Component // Component list
       TrackIds    []int       // Tracking ID list
       Race        bool        // Whether to enable race condition protection
   }
   ```

2. **Component Structure**:
   ```go
   type Component struct {
       ID       int    // Component ID
       Name     string // Component name
       TrackIds []int  // Tracking IDs associated with the component
   }
   ```

3. **FileChange Structure**:
   ```go
   type FileChange struct {
       Path        string      // File path
       LineChanges LineChanges // Line-level change information
   }
   ```

4. **TrackScope Structure**:
   ```go
   type TrackScope struct {
       StartLine int
       EndLine   int
       node      *ast.BlockStmt
       Children  TrackScopes
   }
   ```

### 4.2 Code Generation and Insertion

The code generation process uses Go templates to create the necessary tracking code. The generated code includes:

1. **Tracking Function**:
   ```go
   func Track(id trackId) {
       if id > 0 && id < TRACK_ID_END {
           atomic.StoreUint32(&trackIdStatus[id], 1) // or atomic.AddUint32(&trackIdStatus[id], 1)
       }
   }
   ```

2. **HTTP Service**:
   ```go
   func ServeHTTP(component Component) {
       go func() {
           system := http.NewServeMux()
           system.HandleFunc("/metrics", metricsHandler)
           system.HandleFunc("/track", trackHandler)
           port := "57005"
           if os.Getenv("GOAT_PORT") != "" {
               port = os.Getenv("GOAT_PORT")
           }
           expose := os.Getenv("GOAT_METRICS_IP")
           if expose == "" {
               expose = "127.0.0.1"
           }
           addr := fmt.Sprintf("%s:%s", expose, port)
           log.Printf("Goat track service started: http://%s\n", addr)
           log.Fatal(http.ListenAndServe(addr, system))
       }()
   }
   ```

### 4.3 Concurrency and Performance Optimization

GOAT employs several techniques to optimize performance:

1. **Worker Pool**:
   - Uses a configurable number of worker goroutines for parallel processing
   - Controlled by the `threads` configuration parameter

2. **Efficient Diff Analysis**:
   - Multiple precision levels to balance accuracy and performance
   - Optimized git operations to minimize processing time

3. **Minimal Runtime Overhead**:
   - Efficient tracking code with minimal impact on application performance
   - Optional atomic operations for thread safety in concurrent environments

## 5. Configuration System

GOAT uses a YAML-based configuration system with the following key parameters:

```yaml
# App name
appName: example-app

# App version
appVersion: 1.0.0

# Old branch name
oldBranch: main

# New branch name
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

## 6. Workflow Integration

GOAT is designed to integrate seamlessly into the gray release workflow:

1. **Development Phase**:
   - Developers make code changes in a feature branch
   - GOAT is used to analyze differences between the stable branch and the feature branch

2. **Pre-Release Phase**:
   - GOAT automatically instruments the code with tracking points
   - Developers can review the instrumented code before deployment

3. **Gray Release Phase**:
   - The instrumented application is deployed to a subset of users
   - The HTTP service provides real-time visibility into which code paths are being executed

4. **Post-Release Phase**:
   - GOAT's tracking data helps verify that all incremental code has been properly tested
   - The `goat clean` command can be used to remove instrumentation code after successful deployment

## 7. Technical Limitations and Considerations

1. **Go Language Specificity**:
   - GOAT is specifically designed for Go applications and cannot be used with other languages
   - Requires Go 1.23+ for full functionality

2. **Git Dependency**:
   - Relies on Git for diff analysis
   - Requires a valid Git repository with commit history

3. **AST Analysis Limitations**:
   - May not correctly handle extremely complex code structures
   - Code with unusual formatting or non-standard patterns might require manual adjustments

4. **Performance Impact**:
   - Higher granularity levels (especially line-level) may have a noticeable performance impact
   - Consider using coarser granularity for performance-critical applications

## 8. Future Technical Directions

1. **Enhanced Diff Analysis**:
   - Improved handling of complex code refactorings
   - Better support for moved code blocks

2. **Advanced Instrumentation**:
   - Support for more sophisticated tracking patterns
   - Custom instrumentation strategies based on code patterns

3. **Extended Monitoring**:
   - Integration with more observability platforms
   - Enhanced visualization of tracking data

4. **Performance Optimization**:
   - Further reduction of runtime overhead
   - More efficient instrumentation techniques

## 9. Conclusion

GOAT provides a comprehensive solution for tracking code execution in gray release scenarios. Its modular architecture, flexible granularity system, and efficient implementation make it a powerful tool for ensuring the reliability of incremental code changes in Go applications.

By automatically identifying and instrumenting incremental code, GOAT helps developers make more informed decisions during the gray release process, ultimately leading to more reliable software deployments.
