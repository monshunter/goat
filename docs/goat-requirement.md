# GOAT - Golang Application Gray Release Tracing Tool

## 1. Project Overview

### 1.1 Background

During gray release processes (such as red-blue deployment or canary release), developers and operations personnel typically rely on external metrics (like error rates, business metrics, resource consumption, etc.) to decide whether to proceed with the release. However, from first principles, robust decision-making should be based on internal evidence: ensuring all incremental features are tested during the gray release and external metrics remain within expected ranges.

GOAT (Golang Application Tracing) project aims to provide reliable internal evidence through automated instrumentation, helping developers evaluate the coverage of gray releases and make safer, more reliable decisions about release progression.

### 1.2 Project Goals

- Develop an automated instrumentation tool (command-line tool) providing incremental code execution tracing for Go projects
- Provide easy-to-use APIs for application integration
- Display instrumentation coverage status in real-time through an embedded HTTP service
- Support developer-customized instrumentation strategies
- Minimize performance impact of instrumentation code on applications

## 2. Functional Requirements

### 2.1 Core Features

#### 2.1.1 Identification of Effective Incremental Code

**Effective incremental code** is defined as: new or modified code within function/method bodies that can be executed in this release compared to the stable version.

The following are **not** considered effective incremental code:
- Deleted code
- Incremental code in non-*.go files
- Incremental code in test files (*_test.go)
- New comments or blank lines
- Comments added at line endings (where the code itself remains unchanged)
- Incremental code outside function/method bodies (e.g., global constants, variables, type declarations, interfaces, function declarations)
- Type declarations within functions/methods
- Moved or renamed files

#### 2.1.2 Logical Branch Identification and Instrumentation

GOAT will identify and instrument the following types of logical branches:

**Explicit branches**:
- if-else branches
- switch-case branches
- select-case branches

**Implicit branches**:
- Continuous non-branch statement blocks within function bodies

#### 2.1.3 Instrumentation Rules

- **Pre-instrumentation principle**: Insert instrumentation code at the beginning of branch code blocks
- **Single instrumentation principle**: Each logical branch is instrumented only once
- **Special handling for condition changes**: When branch conditions (e.g., if condition expressions) change, additional instrumentation is inserted for all affected first-level branches
- **Empty branch handling**: Empty select{} or empty switch{} are treated as regular statements with pre-instrumentation

#### 2.1.4 Instrumentation Coverage Tracking

- Provide instrumentation status query interface through embedded HTTP service
- Support coverage status aggregation by component, file, function, etc.
- Support instrumentation coverage statistics

### 2.2 Command-line Tool

#### 2.2.1 Tool Name

```
goat
```
