# Example of Code Tracking at Different Granularities

[中文版](zh-cn/granularity_example.md)

This example demonstrates the differences in tracing code using different granularity levels (line, patch, scope, func) of Goat, helping developers understand how to choose the appropriate tracing granularity.

## Scenario Description

We will perform four different granularity tracings on calculator tool to show their respective differences and applicable scenarios. We will modify it and use different granularities to trace these modifications.

## Original Code

The following is the original counter service code (`calculator/main.go`):

```go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
)

// Calculator provides basic arithmetic operations
type Calculator struct{}

// Add returns the sum of two numbers
func (c *Calculator) Add(a, b int) int {
	return a + b
}

// Subtract returns the difference of two numbers
func (c *Calculator) Subtract(a, b int) int {
	return a - b
}

// Multiply returns the product of two numbers
func (c *Calculator) Multiply(a, b int) int {
	return a * b
}

// Divide returns the quotient of two numbers
func (c *Calculator) Divide(a, b int) int {
	if b == 0 {
		fmt.Println("Error: Division by zero")
		return 0
	}
	return a / b
}

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: calculator <operation> <num1> <num2>")
		fmt.Println("Operations: add, subtract, multiply, divide")
		os.Exit(1)
	}

	operation := os.Args[1]
	num1, err1 := strconv.Atoi(os.Args[2])
	num2, err2 := strconv.Atoi(os.Args[3])

	if err1 != nil || err2 != nil {
		fmt.Println("Error: Arguments must be integers")
		os.Exit(1)
	}

	calc := &Calculator{}
	var result int

	switch operation {
	case "add":
		result = calc.Add(num1, num2)
	case "subtract":
		result = calc.Subtract(num1, num2)
	case "multiply":
		result = calc.Multiply(num1, num2)
	case "divide":
		result = calc.Divide(num1, num2)
	default:
		fmt.Println("Error: Invalid operation")
	}

	fmt.Printf("Result: %d\n", result)
	fmt.Println("Click http://localhost:57005/metrics to see the metrics")
	fmt.Println("Ctrl+C to stop the calculator")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
	fmt.Println("Calculator stopped")
}

```

## Modified Code

We made the following modifications to the original code:
1. Add a Power method
2. Each Operation method prints the input
3. main prints the externally input parameters and modifies the result display copy

The modified code is as follows:

```go
package main

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"strconv"
)

// Calculator provides basic arithmetic operations
type Calculator struct{}

// Add returns the sum of two numbers
func (c *Calculator) Add(a, b int) int {
	fmt.Println("Adding ", a, " and ", b)
	if a > b {
		fmt.Println("a is greater than b")
	}
	return a + b
}

// Subtract returns the difference of two numbers
func (c *Calculator) Subtract(a, b int) int {
	fmt.Println("Subtracting ", a, " and ", b)
	if a < b {
		fmt.Println("a is less than b")
	}
	return a - b
}

// Multiply returns the product of two numbers
func (c *Calculator) Multiply(a, b int) int {
	fmt.Println("Multiplying ", a, " and ", b)
	if a == 0 || b == 0 {
		fmt.Println("One of the numbers is 0")
	}
	return a * b
}

// Divide returns the quotient of two numbers
func (c *Calculator) Divide(a, b int) int {
	fmt.Println("Dividing ", a, " by ", b)
	if b == 0 {
		fmt.Println("Error: Division by zero")
		return 0
	}
	fmt.Println("Quotient: ", a/b)
	return a / b
}

// Power returns a raised to the power of b
func (c *Calculator) Power(a, b int) int {
	fmt.Println("Raising ", a, " to the power of ", b)
	if a == 0 && b == 0 {
		fmt.Println("Error: 0^0 is undefined")
		return 0
	}
	return int(math.Pow(float64(a), float64(b)))
}

func main() {
	fmt.Println("This is a calculator")
	if len(os.Args) != 4 {
		fmt.Println("Usage: calculator <operation> <num1> <num2>")
		fmt.Println("Operations: add, subtract, multiply, divide")
		os.Exit(1)
	}
	fmt.Println("Args: ", os.Args)
	operation := os.Args[1]
	num1, err1 := strconv.Atoi(os.Args[2])
	num2, err2 := strconv.Atoi(os.Args[3])
	fmt.Println("Operation: ", operation)
	fmt.Println("Num1: ", num1)
	fmt.Println("Num2: ", num2)
	if err1 != nil || err2 != nil {
		fmt.Println("Err1", err1)
		fmt.Println("Err2", err2)
		fmt.Println("Error: Arguments must be integers")
		os.Exit(1)
	}

	calc := &Calculator{}
	var result int

	switch operation {
	case "add":
		result = calc.Add(num1, num2)
	case "subtract":
		result = calc.Subtract(num1, num2)
	case "multiply":
		result = calc.Multiply(num1, num2)
	case "divide":
		result = calc.Divide(num1, num2)
	default:
		fmt.Println("Error: Invalid operation")
	}

	fmt.Printf("Calculator Result: %d\n", result)
	fmt.Println("Click http://localhost:57005/metrics to see the metrics")
	fmt.Println("Ctrl+C to stop the calculator")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
	fmt.Println("Calculator stopped")
}

```

## Tracking at Different Granularities

Next, we will show the tracking results at four different granularities:

### 1. Line-level Tracking

Line-level is the finest-grained tracking, which tracks every line modification.

Configure goat.yaml:
```yaml
granularity: line
```

Execute the `goat track` command:
```bash
goat init --old main --new feature/logging-function --app-name "calculator" --granularity line --force
goat track
```

Line-level execution process information output:
```
[2025/04/24 00:00:22] INFO: Start to run track command
[2025/04/24 00:00:22] INFO: Track applied successfully with 20 tracking points
[2025/04/24 00:00:22] INFO: ----------------------------------------------------------
[2025/04/24 00:00:22] INFO: ✅ Track completed successfully!
```

Example of Line-level tracking results:

```go
package main

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"strconv"
	goat "calculator/goat"
)

// Calculator provides basic arithmetic operations
type Calculator struct{}

// Add returns the sum of two numbers
func (c *Calculator) Add(a, b int) int {
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_1)
	// +goat:end
	fmt.Println("Adding ", a, " and ", b)
	if a > b {
		// +goat:generate
		// +goat:tips: do not edit the block between the +goat comments
		goat.Track(goat.TRACK_ID_2)
		// +goat:end
		fmt.Println("a is greater than b")
	}
	return a + b
}

// Subtract returns the difference of two numbers
func (c *Calculator) Subtract(a, b int) int {
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_3)
	// +goat:end
	fmt.Println("Subtracting ", a, " and ", b)
	if a < b {
		// +goat:generate
		// +goat:tips: do not edit the block between the +goat comments
		goat.Track(goat.TRACK_ID_4)
		// +goat:end
		fmt.Println("a is less than b")
	}
	return a - b
}

// Multiply returns the product of two numbers
func (c *Calculator) Multiply(a, b int) int {
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_5)
	// +goat:end
	fmt.Println("Multiplying ", a, " and ", b)
	if a == 0 || b == 0 {
		// +goat:generate
		// +goat:tips: do not edit the block between the +goat comments
		goat.Track(goat.TRACK_ID_6)
		// +goat:end
		fmt.Println("One of the numbers is 0")
	}
	return a * b
}

// Divide returns the quotient of two numbers
func (c *Calculator) Divide(a, b int) int {
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_7)
	// +goat:end
	fmt.Println("Dividing ", a, " by ", b)
	if b == 0 {
		fmt.Println("Error: Division by zero")
		return 0
	}
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_8)
	// +goat:end
	fmt.Println("Quotient: ", a/b)
	return a / b
}

// Power returns a raised to the power of b
func (c *Calculator) Power(a, b int) int {
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_9)
	// +goat:end
	fmt.Println("Raising ", a, " to the power of ", b)
	if a == 0 && b == 0 {
		// +goat:generate
		// +goat:tips: do not edit the block between the +goat comments
		goat.Track(goat.TRACK_ID_10)
		// +goat:end
		fmt.Println("Error: 0^0 is undefined")
		// +goat:generate
		// +goat:tips: do not edit the block between the +goat comments
		goat.Track(goat.TRACK_ID_11)
		// +goat:end
		return 0
	}
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_12)
	// +goat:end
	return int(math.Pow(float64(a), float64(b)))
}

func main() {
	// +goat:main
	// +goat:tips: do not edit the block between the +goat comments
	goat.ServeHTTP(goat.COMPONENT_0)
	// +goat:end
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_13)
	// +goat:end
	fmt.Println("This is a calculator")
	if len(os.Args) != 4 {
		fmt.Println("Usage: calculator <operation> <num1> <num2>")
		fmt.Println("Operations: add, subtract, multiply, divide")
		os.Exit(1)
	}
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_14)
	// +goat:end
	fmt.Println("Args: ", os.Args)
	operation := os.Args[1]
	num1, err1 := strconv.Atoi(os.Args[2])
	num2, err2 := strconv.Atoi(os.Args[3])
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_15)
	// +goat:end
	fmt.Println("Operation: ", operation)
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_16)
	// +goat:end
	fmt.Println("Num1: ", num1)
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_17)
	// +goat:end
	fmt.Println("Num2: ", num2)
	if err1 != nil || err2 != nil {
		// +goat:generate
		// +goat:tips: do not edit the block between the +goat comments
		goat.Track(goat.TRACK_ID_18)
		// +goat:end
		fmt.Println("Err1", err1)
		// +goat:generate
		// +goat:tips: do not edit the block between the +goat comments
		goat.Track(goat.TRACK_ID_19)
		// +goat:end
		fmt.Println("Err2", err2)
		fmt.Println("Error: Arguments must be integers")
		os.Exit(1)
	}

	calc := &Calculator{}
	var result int

	switch operation {
	case "add":
		result = calc.Add(num1, num2)
	case "subtract":
		result = calc.Subtract(num1, num2)
	case "multiply":
		result = calc.Multiply(num1, num2)
	case "divide":
		result = calc.Divide(num1, num2)
	default:
		fmt.Println("Error: Invalid operation")
	}

	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_20)
	// +goat:end
	fmt.Printf("Calculator Result: %d\n", result)
	fmt.Println("Click http://localhost:57005/metrics to see the metrics")
	fmt.Println("Ctrl+C to stop the calculator")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
	fmt.Println("Calculator stopped")
}


```

**Features**:
- The finest tracking granularity, with logging for each line modification, totaling 20.
- Can accurately locate the specific line where the code is executed.
- A large number of logging points may have a greater impact on performance.
- Suitable for scenarios that require very detailed tracking.

### 2. Patch-level tracking

Patch-level tracking will track continuous modification blocks and is a more balanced choice.

Configure goat.yaml:
```yaml
granularity: patch
```

Execute the `goat track` command:
```bash
goat init --old main --new feature/logging-function --app-name "calculator" --granularity patch --force
goat track
```

Patch level execution process information output:

```
[2025/04/24 00:04:45] INFO: Start to run track command
[2025/04/24 00:04:45] INFO: Track applied successfully with 16 tracking points
[2025/04/24 00:04:45] INFO: ----------------------------------------------------------
[2025/04/24 00:04:45] INFO: ✅ Track completed successfully!
```

Patch level tracking result example:

```go
package main

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"strconv"
	goat "calculator/goat"
)

// Calculator provides basic arithmetic operations
type Calculator struct{}

// Add returns the sum of two numbers
func (c *Calculator) Add(a, b int) int {
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_1)
	// +goat:end
	fmt.Println("Adding ", a, " and ", b)
	if a > b {
		// +goat:generate
		// +goat:tips: do not edit the block between the +goat comments
		goat.Track(goat.TRACK_ID_2)
		// +goat:end
		fmt.Println("a is greater than b")
	}
	return a + b
}

// Subtract returns the difference of two numbers
func (c *Calculator) Subtract(a, b int) int {
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_3)
	// +goat:end
	fmt.Println("Subtracting ", a, " and ", b)
	if a < b {
		// +goat:generate
		// +goat:tips: do not edit the block between the +goat comments
		goat.Track(goat.TRACK_ID_4)
		// +goat:end
		fmt.Println("a is less than b")
	}
	return a - b
}

// Multiply returns the product of two numbers
func (c *Calculator) Multiply(a, b int) int {
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_5)
	// +goat:end
	fmt.Println("Multiplying ", a, " and ", b)
	if a == 0 || b == 0 {
		// +goat:generate
		// +goat:tips: do not edit the block between the +goat comments
		goat.Track(goat.TRACK_ID_6)
		// +goat:end
		fmt.Println("One of the numbers is 0")
	}
	return a * b
}

// Divide returns the quotient of two numbers
func (c *Calculator) Divide(a, b int) int {
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_7)
	// +goat:end
	fmt.Println("Dividing ", a, " by ", b)
	if b == 0 {
		fmt.Println("Error: Division by zero")
		return 0
	}
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_8)
	// +goat:end
	fmt.Println("Quotient: ", a/b)
	return a / b
}

// Power returns a raised to the power of b
func (c *Calculator) Power(a, b int) int {
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_9)
	// +goat:end
	fmt.Println("Raising ", a, " to the power of ", b)
	if a == 0 && b == 0 {
		// +goat:generate
		// +goat:tips: do not edit the block between the +goat comments
		goat.Track(goat.TRACK_ID_10)
		// +goat:end
		fmt.Println("Error: 0^0 is undefined")
		return 0
	}
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_11)
	// +goat:end
	return int(math.Pow(float64(a), float64(b)))
}

func main() {
	// +goat:main
	// +goat:tips: do not edit the block between the +goat comments
	goat.ServeHTTP(goat.COMPONENT_0)
	// +goat:end
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_12)
	// +goat:end
	fmt.Println("This is a calculator")
	if len(os.Args) != 4 {
		fmt.Println("Usage: calculator <operation> <num1> <num2>")
		fmt.Println("Operations: add, subtract, multiply, divide")
		os.Exit(1)
	}
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_13)
	// +goat:end
	fmt.Println("Args: ", os.Args)
	operation := os.Args[1]
	num1, err1 := strconv.Atoi(os.Args[2])
	num2, err2 := strconv.Atoi(os.Args[3])
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_14)
	// +goat:end
	fmt.Println("Operation: ", operation)
	fmt.Println("Num1: ", num1)
	fmt.Println("Num2: ", num2)
	if err1 != nil || err2 != nil {
		// +goat:generate
		// +goat:tips: do not edit the block between the +goat comments
		goat.Track(goat.TRACK_ID_15)
		// +goat:end
		fmt.Println("Err1", err1)
		fmt.Println("Err2", err2)
		fmt.Println("Error: Arguments must be integers")
		os.Exit(1)
	}

	calc := &Calculator{}
	var result int

	switch operation {
	case "add":
		result = calc.Add(num1, num2)
	case "subtract":
		result = calc.Subtract(num1, num2)
	case "multiply":
		result = calc.Multiply(num1, num2)
	case "divide":
		result = calc.Divide(num1, num2)
	default:
		fmt.Println("Error: Invalid operation")
	}

	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_16)
	// +goat:end
	fmt.Printf("Calculator Result: %d\n", result)
	fmt.Println("Click http://localhost:57005/metrics to see the metrics")
	fmt.Println("Ctrl+C to stop the calculator")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
	fmt.Println("Calculator stopped")
}

```
**Features**:
- Moderate tracking granularity, tracking code blocks with consecutive modifications within the scope. Each patch performs instrumentation only once, with a total of 16.
- Reasonable number of instrumentations, balancing tracking granularity and performance impact.
- Suitable for most canary release scenarios and is the default recommended tracking granularity.

### 3. Scope-Level Tracking

Scope-level tracking monitors modifications across an entire scope (such as if blocks, for loops, etc.).

Configure goat.yaml:
```yaml
granularity: scope
```

Execute the `goat track` command:
```bash
goat init --old main --new feature/logging-function --app-name "calculator" --granularity scope --force
goat track
```

Output of execution process information at the Scope level:

```
[2025/04/24 00:11:49] INFO: Start to run track command
[2025/04/24 00:11:49] INFO: Track applied successfully with 15 tracking points
[2025/04/24 00:11:49] INFO: ----------------------------------------------------------
[2025/04/24 00:11:49] INFO: ✅ Track completed successfully!
```

Example of Scope level tracking results:

```go
package main

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"strconv"
	goat "calculator/goat"
)

// Calculator provides basic arithmetic operations
type Calculator struct{}

// Add returns the sum of two numbers
func (c *Calculator) Add(a, b int) int {
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_1)
	// +goat:end
	fmt.Println("Adding ", a, " and ", b)
	if a > b {
		// +goat:generate
		// +goat:tips: do not edit the block between the +goat comments
		goat.Track(goat.TRACK_ID_2)
		// +goat:end
		fmt.Println("a is greater than b")
	}
	return a + b
}

// Subtract returns the difference of two numbers
func (c *Calculator) Subtract(a, b int) int {
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_3)
	// +goat:end
	fmt.Println("Subtracting ", a, " and ", b)
	if a < b {
		// +goat:generate
		// +goat:tips: do not edit the block between the +goat comments
		goat.Track(goat.TRACK_ID_4)
		// +goat:end
		fmt.Println("a is less than b")
	}
	return a - b
}

// Multiply returns the product of two numbers
func (c *Calculator) Multiply(a, b int) int {
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_5)
	// +goat:end
	fmt.Println("Multiplying ", a, " and ", b)
	if a == 0 || b == 0 {
		// +goat:generate
		// +goat:tips: do not edit the block between the +goat comments
		goat.Track(goat.TRACK_ID_6)
		// +goat:end
		fmt.Println("One of the numbers is 0")
	}
	return a * b
}

// Divide returns the quotient of two numbers
func (c *Calculator) Divide(a, b int) int {
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_7)
	// +goat:end
	fmt.Println("Dividing ", a, " by ", b)
	if b == 0 {
		fmt.Println("Error: Division by zero")
		return 0
	}
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_8)
	// +goat:end
	fmt.Println("Quotient: ", a/b)
	return a / b
}

// Power returns a raised to the power of b
func (c *Calculator) Power(a, b int) int {
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_9)
	// +goat:end
	fmt.Println("Raising ", a, " to the power of ", b)
	if a == 0 && b == 0 {
		// +goat:generate
		// +goat:tips: do not edit the block between the +goat comments
		goat.Track(goat.TRACK_ID_10)
		// +goat:end
		fmt.Println("Error: 0^0 is undefined")
		return 0
	}
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_11)
	// +goat:end
	return int(math.Pow(float64(a), float64(b)))
}

func main() {
	// +goat:main
	// +goat:tips: do not edit the block between the +goat comments
	goat.ServeHTTP(goat.COMPONENT_0)
	// +goat:end
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_12)
	// +goat:end
	fmt.Println("This is a calculator")
	if len(os.Args) != 4 {
		fmt.Println("Usage: calculator <operation> <num1> <num2>")
		fmt.Println("Operations: add, subtract, multiply, divide")
		os.Exit(1)
	}
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_13)
	// +goat:end
	fmt.Println("Args: ", os.Args)
	operation := os.Args[1]
	num1, err1 := strconv.Atoi(os.Args[2])
	num2, err2 := strconv.Atoi(os.Args[3])
	fmt.Println("Operation: ", operation)
	fmt.Println("Num1: ", num1)
	fmt.Println("Num2: ", num2)
	if err1 != nil || err2 != nil {
		// +goat:generate
		// +goat:tips: do not edit the block between the +goat comments
		goat.Track(goat.TRACK_ID_14)
		// +goat:end
		fmt.Println("Err1", err1)
		fmt.Println("Err2", err2)
		fmt.Println("Error: Arguments must be integers")
		os.Exit(1)
	}

	calc := &Calculator{}
	var result int

	switch operation {
	case "add":
		result = calc.Add(num1, num2)
	case "subtract":
		result = calc.Subtract(num1, num2)
	case "multiply":
		result = calc.Multiply(num1, num2)
	case "divide":
		result = calc.Divide(num1, num2)
	default:
		fmt.Println("Error: Invalid operation")
	}

	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_15)
	// +goat:end
	fmt.Printf("Calculator Result: %d\n", result)
	fmt.Println("Click http://localhost:57005/metrics to see the metrics")
	fmt.Println("Ctrl+C to stop the calculator")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
	fmt.Println("Calculator stopped")
}

```

**Features**:
- Coarse tracking granularity, tracking for the entire scope. One instrumentation is performed within the same scope, for a total of 15 times, 1 less than patch (`fmt.Println("Operation: ", operation)` is no longer instrumented because it belongs to the same scope as `fmt.Println("Args: ", os.Args)`).
- Fewer instrumentation points, with little impact on performance.
- Can track the execution of the entire statement block.
- Suitable for modifying code that is scattered but logically related.

### 4. Func-level Tracking

Func-level is the coarsest level of tracking, only tracking modifications at the function level.

Configure goat.yaml:
```yaml
granularity: func
```

Execute the `goat track` command:
```bash
goat init --old main --new feature/logging-function --app-name "calculator" --granularity func --force
goat track
```

Output of execution process information at the func level:

```
[2025/04/24 00:18:03] INFO: Start to run track command
[2025/04/24 00:18:03] INFO: Track applied successfully with 6 tracking points
[2025/04/24 00:18:03] INFO: ----------------------------------------------------------
[2025/04/24 00:18:03] INFO: ✅ Track completed successfully!
```

Example of Func level tracing results:

```go
package main

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"strconv"
	goat "calculator/goat"
)

// Calculator provides basic arithmetic operations
type Calculator struct{}

// Add returns the sum of two numbers
func (c *Calculator) Add(a, b int) int {
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_1)
	// +goat:end
	fmt.Println("Adding ", a, " and ", b)
	if a > b {
		fmt.Println("a is greater than b")
	}
	return a + b
}

// Subtract returns the difference of two numbers
func (c *Calculator) Subtract(a, b int) int {
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_2)
	// +goat:end
	fmt.Println("Subtracting ", a, " and ", b)
	if a < b {
		fmt.Println("a is less than b")
	}
	return a - b
}

// Multiply returns the product of two numbers
func (c *Calculator) Multiply(a, b int) int {
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_3)
	// +goat:end
	fmt.Println("Multiplying ", a, " and ", b)
	if a == 0 || b == 0 {
		fmt.Println("One of the numbers is 0")
	}
	return a * b
}

// Divide returns the quotient of two numbers
func (c *Calculator) Divide(a, b int) int {
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_4)
	// +goat:end
	fmt.Println("Dividing ", a, " by ", b)
	if b == 0 {
		fmt.Println("Error: Division by zero")
		return 0
	}
	fmt.Println("Quotient: ", a/b)
	return a / b
}

// Power returns a raised to the power of b
func (c *Calculator) Power(a, b int) int {
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_5)
	// +goat:end
	fmt.Println("Raising ", a, " to the power of ", b)
	if a == 0 && b == 0 {
		fmt.Println("Error: 0^0 is undefined")
		return 0
	}
	return int(math.Pow(float64(a), float64(b)))
}

func main() {
	// +goat:main
	// +goat:tips: do not edit the block between the +goat comments
	goat.ServeHTTP(goat.COMPONENT_0)
	// +goat:end
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_6)
	// +goat:end
	fmt.Println("This is a calculator")
	if len(os.Args) != 4 {
		fmt.Println("Usage: calculator <operation> <num1> <num2>")
		fmt.Println("Operations: add, subtract, multiply, divide")
		os.Exit(1)
	}
	fmt.Println("Args: ", os.Args)
	operation := os.Args[1]
	num1, err1 := strconv.Atoi(os.Args[2])
	num2, err2 := strconv.Atoi(os.Args[3])
	fmt.Println("Operation: ", operation)
	fmt.Println("Num1: ", num1)
	fmt.Println("Num2: ", num2)
	if err1 != nil || err2 != nil {
		fmt.Println("Err1", err1)
		fmt.Println("Err2", err2)
		fmt.Println("Error: Arguments must be integers")
		os.Exit(1)
	}

	calc := &Calculator{}
	var result int

	switch operation {
	case "add":
		result = calc.Add(num1, num2)
	case "subtract":
		result = calc.Subtract(num1, num2)
	case "multiply":
		result = calc.Multiply(num1, num2)
	case "divide":
		result = calc.Divide(num1, num2)
	default:
		fmt.Println("Error: Invalid operation")
	}

	fmt.Printf("Calculator Result: %d\n", result)
	fmt.Println("Click http://localhost:57005/metrics to see the metrics")
	fmt.Println("Ctrl+C to stop the calculator")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
	fmt.Println("Calculator stopped")
}

```

**Features**:
- The coarsest tracking granularity, only tracking to the function level (no instrumentation will be added if there are no changes inside the function), with a total of 6 instrumentation points.
- The least number of instrumentation points, with the least impact on performance.
- Can only determine whether a function is called, unable to track the internal execution path of the function.
- Suitable for scenarios where the entire function is newly added or modified, or scenarios where only whether the function is called is of concern.

## Suggestions for Selection at Different Granularities

| Granularity Level | Applicable Scenarios | Advantages | Disadvantages |
|---------|---------|------|------|
| line    | Need to trace the code execution path very detailedly<br>The code modifications are highly scattered<br>Need to accurately locate the performance bottleneck | The most accurate tracing<br>Can discover all execution paths | Large number of instrumentation points<br>Great impact on performance<br>Large amount of tracing data generated |
| patch   | Most gray release scenarios<br>The code modifications are concentrated in several blocks<br>Need to balance the tracing granularity and performance | Moderate tracing granularity<br>Acceptable performance impact<br>Reasonable amount of tracing data | Unable to trace the execution at the single-line level<br>Different logical blocks may be merged |
| scope   | The code modifications are distributed in multiple code blocks<br>Mainly concerned about whether the conditional branches are executed<br>Performance-sensitive applications | Fewer instrumentation points<br>Small impact on performance<br>Can trace the main logical blocks | Coarse granularity<br>Unable to trace the detailed execution path |
| func    | Mainly add entire functions<br>Only care about whether the functions are called<br>Applications that are very concerned about performance | The fewest instrumentation points<br>The least impact on performance | The coarsest tracing granularity<br>Unable to trace the execution path inside the functions |

## Summary

Through this example, we have demonstrated four different tracing granularities supported by the goat tool and analyzed their respective characteristics, advantages and disadvantages, as well as applicable scenarios. In actual use, developers can choose the most suitable tracing granularity according to their own needs:

1. If very detailed execution path information is required, the line level can be selected.
2. If you want to balance the tracing granularity and performance impact, the patch level (default recommended) can be selected.
3. If you are concerned about the execution of the main logical blocks, the scope level can be selected.
4. If you only care about whether a function is called, the func level can be selected.

Different granularities are applicable to different scenarios, and developers should make a choice based on specific requirements and performance considerations.
