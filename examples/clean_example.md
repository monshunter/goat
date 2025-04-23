# goat clean Usage Example

[中文版](zh-cn/clean_example.md)

This example demonstrates how to use the `goat clean` command to remove the tracking code inserted previously by `goat track`.

## Scenario Description

After the end of the gray release phase, or when it is necessary to switch to a different tracking strategy, we need to clean up the inserted tracking code. The `goat clean` command can easily accomplish this task.

## Original Code (Including Tracking Code)

The following is the calculator application code after inserting the tracking code by the `goat track` command:

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
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_1)
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
		// +goat:generate
		// +goat:tips: do not edit the block between the +goat comments
		goat.Track(goat.TRACK_ID_2)
		// +goat:end
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
Note that the above code contains the following tracing code:
```
Import the goat package
`goat "calculator/goat"`
```
```
Start an HTTP Service at the entry point of the main function
`// +goat:main
 // +goat:tips: do not edit the block between the +goat comments
 goat.ServeHTTP(goat.COMPONENT_0)
 // +goat:end`
```

```
Tracing call in the Power method
`// +goat:generate
 // +goat:tips: do not edit the block between the +goat comments
 goat.Track(goat.TRACK_ID_1)
 // +goat:end`
```

```
Tracing call in the main function
`// +goat:generate
 // +goat:tips: do not edit the block between the +goat comments
 goat.Track(goat.TRACK_ID_2)
 // +goat:end`
```

## Step 1: Run the goat clean command

When we have completed gray-box tracing and no longer need this code, we can use the `goat clean` command to remove it:

```bash
# Ensure that you are currently in the project root directory
cd calculator
```

```bash
# Confirm the files in the current path
ls
```

The files in the current directory are as follows:
```
go.mod  goat  goat.yaml  main.go
```

```bash
# Run the goat clean command
goat clean
```

Example command output:

```
[2025/04/23 22:40:45] INFO: Start to run clean command
[2025/04/23 22:40:45] INFO: Total cleaned files: 1
[2025/04/23 22:40:45] INFO: Cleaned project
[2025/04/23 22:40:45] INFO: ----------------------------------------------------------
[2025/04/23 22:40:45] INFO: ✅ Clean completed successfully!
```

## Step 2: Check the Cleaned Code

After running `goat clean`, all the tracing code is removed. Check the cleaned code:

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

```bash
# Reconfirm the files in the current path
ls
```

The files in the current directory are as follows:
```
go.mod  goat.yaml  main.go
```

As can be seen, all the tracing code has been completely removed, including:
- The import of the `goat` package
- The declaration of the tracing component variables
- All calls to the `Track` method
- The `goat` directory has been deleted

## Step 3: Submit the Cleaned Code

After cleaning, we can submit the code and merge it into the main branch:

```bash
# Add changes
git add.

# Commit changes
git commit -m "Remove tracking code with goat clean"

# Switch to the main branch
git checkout main

# Merge the feature branch
git merge feature/power-function

# Push to the remote repository
git push origin main
```

## When to Use goat clean

The `goat clean` command is particularly useful in the following situations:

1. **After canary release**: When the new feature passes the canary test, the tracing code can be removed.
2. **Changing tracing strategy**: When adjusting the tracing granularity or method, clean the old tracing code first.
3. **Preparing for production deployment**: Remove the tracing code in the development or testing phase before deploying to the production environment.
4. **Resetting tracing status**: When it is found that there is a problem with the tracing code and it needs to be regenerated.

## Combined Use with Other Commands

`goat clean` is usually combined with other goat commands to form a complete workflow:

1. Use `goat init` to initialize the configuration.
2. Use `goat track` to insert the tracing code.
3. Run the application and collect the buried point data.
4. Use `goat clean` to clean the tracing code.
5. Repeat the above process after adjusting the configuration as needed.

## Summary

Through this example, we demonstrated how to use the `goat clean` command to remove the tracing code and keep the code clean. This command provides a convenient way to manage the lifecycle of the tracing code during the canary release process, ensuring that the tracing code can be easily added or removed when needed.
