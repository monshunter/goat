# 不同粒度下的代码跟踪示例

[English](../granularity_example.md)

本示例演示了使用不同粒度级别（行、块、作用域、函数）对代码进行追踪的差异，帮助开发者理解如何选择合适的追踪粒度。

## 场景描述

我们将对一个用户管理服务进行四种不同粒度的追踪，我们将修改它并使用不同粒度来追踪这些修改。

## 原始代码

以下是原始计数器服务代码（`calculator/main.go`）：

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

## 修改后的代码

我们对原始代码进行了以下修改：
1. 添加了一个Power方法
2. 每个Operation方法都会打印输入
3. main方法打印外部输入参数并修改结果显示副本

修改后的代码如下：

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

## 不同粒度的跟踪

接下来，我们将展示四种不同粒度的跟踪结果：

### 1. 行级跟踪

Line 是最细粒度的跟踪，它会跟踪每一行的修改。

配置goat.yaml：
```yaml
granularity: line
```

执行`goat track`命令：
```bash
goat init --old main --new feature/logging-function --app-name "calculator" --granularity line --force
goat track
```

Line 级别执行过程信息输出：
```
[2025/04/24 00:00:22] INFO: Start to run track command
[2025/04/24 00:00:22] INFO: Track applied successfully with 20 tracking points
[2025/04/24 00:00:22] INFO: ----------------------------------------------------------
[2025/04/24 00:00:22] INFO: ✅ Track completed successfully!
```

Line 级别跟踪结果示例：

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

**功能特性**：
- 最精细的跟踪粒度，记录每一行修改，共20条。
- 能够精确地定位代码执行的具体行。
- 大量的日志记录点可能会对性能产生较大影响。
- 适用于需要非常详细跟踪的场景。

### 2. 补丁级跟踪

补丁级跟踪将跟踪连续的修改块，是一种更为平衡的选择。

配置goat.yaml：
```yaml
granularity: patch
```

执行`goat track`命令：
```bash
goat init --old main --new feature/logging-function --app-name "calculator" --granularity patch --force
goat track
```

Patch 级别执行过程信息输出：

```
[2025/04/24 00:04:45] INFO: Start to run track command
[2025/04/24 00:04:45] INFO: Track applied successfully with 16 tracking points
[2025/04/24 00:04:45] INFO: ----------------------------------------------------------
[2025/04/24 00:04:45] INFO: ✅ Track completed successfully!
```

Patch 级别跟踪结果示例：

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
**功能特性**：
- 适度的跟踪粒度，跟踪范围内连续修改的代码块。每个补丁仅执行一次插桩，总共16次。
- 插桩数量合理，平衡了跟踪粒度和性能影响。
- 适用于大多数金丝雀发布场景，是默认推荐的跟踪粒度。

### 3. 范围级跟踪

范围级跟踪会监控整个范围内的修改（例如if块、for循环等）。

配置goat.yaml：
```yaml
granularity: scope
```
执行"goat track"命令：
```bash
goat init --old main --new feature/logging-function --app-name "calculator" --granularity scope --force
goat track
```

Scope 级别的执行过程信息输出：

```
[2025/04/24 00:11:49] INFO: Start to run track command
[2025/04/24 00:11:49] INFO: Track applied successfully with 15 tracking points
[2025/04/24 00:11:49] INFO: ----------------------------------------------------------
[2025/04/24 00:11:49] INFO: ✅ Track completed successfully!
```

Scope 级别跟踪结果示例：

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

**特性**：
- 粗略的跟踪粒度，对整个作用域进行跟踪。在同一作用域内进行一次插桩，总共插桩15次，比补丁少1次（`fmt.Println("Operation: ", operation)`不再插桩，因为它与`fmt.Println("Args: ", os.Args)`属于同一作用域）。
- 插桩点较少，对性能影响小。
- 可以跟踪整个语句块的执行。
- 适用于修改分散但逻辑相关的代码。

### 4. 函数级跟踪

Func 级是最粗略的跟踪级别，仅在函数级别跟踪修改。

配置goat.yaml：
```yaml
granularity: func
```

执行“goat track”命令：
```bash
goat init --old main --new feature/logging-function --app-name "calculator" --granularity func --force
goat track
```

Func 级别执行过程信息的输出：

```
[2025/04/24 00:18:03] INFO: Start to run track command
[2025/04/24 00:18:03] INFO: Track applied successfully with 6 tracking points
[2025/04/24 00:18:03] INFO: ----------------------------------------------------------
[2025/04/24 00:18:03] INFO: ✅ Track completed successfully!
```

Func 级跟踪结果示例：

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

**功能特性**：
- 最粗粒度的跟踪，仅跟踪到函数级别（如果函数内部没有变化则不会添加插桩），总共6个插桩点。
- 插桩点数量最少，对性能影响最小。
- 只能确定函数是否被调用，无法跟踪函数内部执行路径。
- 适用于整个函数被新增或修改的场景，或者仅关注函数是否被调用的场景。

## 不同粒度下的选择建议

| 粒度级别 | 适用场景 | 优点 | 缺点 |
| ---- | ---- | ---- | ---- |
| 行 | 需要非常详细地追踪代码执行路径<br>代码修改非常分散<br>需要精确定位性能瓶颈 | 最精确的追踪<br>可以发现所有执行路径 | 插桩点数量众多<br>对性能影响大<br>生成的追踪数据量巨大 |
| 补丁 | 大多数灰度发布场景<br>代码修改集中在几个代码块<br>需要平衡追踪粒度和性能 | 追踪粒度适中<br>性能影响可接受<br>追踪数据量合理 | 无法在单行级别追踪执行情况<br>不同逻辑块可能会合并 |
| 范围 | 代码修改分布在多个代码块<br>主要关注条件分支是否被执行<br>对性能敏感的应用 | 插桩点较少<br>对性能影响小<br>可以追踪主要逻辑块 | 粒度较粗<br>无法追踪详细的执行路径 |
| 函数 | 主要是添加整个函数<br>只关心函数是否被调用<br>非常关注性能的应用 | 插桩点最少<br>对性能影响最小 | 追踪粒度最粗<br>无法追踪函数内部的执行路径 |

## 总结

通过这个示例，我们展示了goat工具支持的四种不同的追踪粒度，并分析了它们各自的特点、优缺点以及适用场景。在实际使用中，开发者可以根据自身需求选择最合适的追踪粒度：

1. 如果需要非常详细的执行路径信息，可以选择行级别。
2. 如果想要平衡追踪粒度和性能影响，可以选择补丁级别（默认推荐）。
3. 如果关注主要逻辑块的执行情况，可以选择范围级别。
4. 如果只关心函数是否被调用，可以选择函数级别。

不同的粒度适用于不同的场景，开发者应根据具体需求和性能考虑做出选择。
