# goat patch Usage Example

[中文版](zh-cn/patch_example.md)

This example demonstrates how to use the `goat patch` command to handle manually added markers and achieve flexible code insertion and deletion operations.

## Scenario Description

This is a simple calculator and uses the `goat patch` command to process these tokens.

## Creating Example Code

The following is a simple calculator (`calculator/main.go`):

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

// Power returns a raised to the power of b
func (c *Calculator) Power(a, b int) int {
	return int(math.Pow(float64(a), float64(b)))
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
	case "power":
		result = calc.Power(num1, num2)
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

## Step 1: Add Manual Logging Markers

We need to observe whether each operator has been correctly tested, so we manually added a logging marker: `// +goat:insert`.

Modified code (with manual insert markers added):

```go
// Existing code ...
// Add returns the sum of two numbers
func (c *Calculator) Add(a, b int) int {
	// +goat:insert
	return a + b
}

// Subtract returns the difference of two numbers
func (c *Calculator) Subtract(a, b int) int {
	// +goat:insert
	return a - b
}

// Multiply returns the product of two numbers
func (c *Calculator) Multiply(a, b int) int {
	// +goat:insert
	return a * b
}

// Divide returns the quotient of two numbers
func (c *Calculator) Divide(a, b int) int {
	if b == 0 {
		fmt.Println("Error: Division by zero")
		return 0
	}
	// +goat:insert
	return a / b
}

// Power returns a raised to the power of b
func (c *Calculator) Power(a, b int) int {
	// +goat:insert
	return int(math.Pow(float64(a), float64(b)))
}
// Existing code ...
```

Note in the above code, we added multiple `// +goat:insert` markers, and each marker is followed by a return statement.

## Step 2: Run the goat patch command to process the markers

```bash
# Initialize the goat configuration (if not already initialized)
goat init --app-name "calculator"

# Run the goat patch command to process the markers
goat patch
```

Example of command output:

```
[2025/04/23 23:17:41] INFO: Start to run patch command
[2025/04/23 23:17:41] INFO: Total replaced tracks: 5
[2025/04/23 23:17:41] INFO: Patch applied
[2025/04/23 23:17:41] INFO: ----------------------------------------------------------
[2025/04/23 23:17:41] INFO: ✅ Patch applied successfully!
```

## Step 3: Review the Processed Code

After running `goat patch`, all `+goat:insert` markers are processed, and the commented code is uncommented
 and inserted at the marker:

```go

package main

import (
	"fmt"
	goat "calculator/goat"
	"math"
	"os"
	"os/signal"
	"strconv"
)

// Calculator provides basic arithmetic operations
type Calculator struct{}

// Add returns the sum of two numbers
func (c *Calculator) Add(a, b int) int {
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_1)
	// +goat:end

	return a + b
}

// Subtract returns the difference of two numbers
func (c *Calculator) Subtract(a, b int) int {
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_2)
	// +goat:end

	return a - b
}

// Multiply returns the product of two numbers
func (c *Calculator) Multiply(a, b int) int {
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_3)
	// +goat:end

	return a * b
}

// Divide returns the quotient of two numbers
func (c *Calculator) Divide(a, b int) int {
	if b == 0 {
		fmt.Println("Error: Division by zero")
		return 0
	}
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_4)
	// +goat:end

	return a / b
}

// Power returns a raised to the power of b
func (c *Calculator) Power(a, b int) int {
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_5)
	// +goat:end

	return int(math.Pow(float64(a), float64(b)))
}

func main() {
	// +goat:main
	// +goat:tips: do not edit the block between the +goat comments
	goat.ServeHTTP(goat.COMPONENT_0)
	// +goat:end
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
	case "power":
		result = calc.Power(num1, num2)
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

## Step 4: Use the goat:delete Tag

Suppose during testing, we find that addition and subtraction are stable enough and no longer need to be observed.
We can use the `+goat:delete` tag to mark the data points to be deleted:

```go
// Add returns the sum of two numbers
func (c *Calculator) Add(a, b int) int {
	// +goat:delete
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_1)
	// +goat:end

	return a + b
}

// Subtract returns the difference of two numbers
func (c *Calculator) Subtract(a, b int) int {
	// +goat:delete
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_2)
	// +goat:end

	return a - b
}
```

Run the `goat patch` command again:

```bash
goat patch
```

The processed code:

```go

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
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_1)
	// +goat:end

	return a * b
}

// Divide returns the quotient of two numbers
func (c *Calculator) Divide(a, b int) int {
	if b == 0 {
		fmt.Println("Error: Division by zero")
		return 0
	}
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_2)
	// +goat:end

	return a / b
}

// Power returns a raised to the power of b
func (c *Calculator) Power(a, b int) int {
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_3)
	// +goat:end

	return int(math.Pow(float64(a), float64(b)))
}

func main() {
	// +goat:main
	// +goat:tips: do not edit the block between the +goat comments
	goat.ServeHTTP(goat.COMPONENT_0)
	// +goat:end
	if len(os.Args) != 4 {
		fmt.Println("Usage: calculator <operation> <num1> <num2>")
		fmt.Println("Operations: add, subtract, multiply, divide")
		os.Exit(1)
	}
	// Existing code ...
}
```

## Summary

Through this example, we demonstrated how to use the `goat patch` command to handle manual tracing markers and achieve the following functions:

1. Use the `+goat:insert` marker to insert code at specific positions, which can be used for gray release of new features.
2. Use the `+goat:delete` marker to delete the inserted tracing points, which can be used to correct tracing points.
3. Add or delete tracing points through simple markers without modifying the existing code.

This manual marking method provides more refined control for gray release, allowing developers to flexibly enable or disable specific functions without modifying the core code logic.
