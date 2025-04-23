# goat patch 使用示例

[English](../patch_example.md)

本示例演示如何使用 `goat patch` 命令处理手动添加的标记并实现灵活的代码插入和删除操作。

## 场景描述

这是一个简单的计算器，使用 `goat patch` 命令处理这些标记。

## 创建示例代码

以下是一个简单的计算器（`calculator/main.go`）：

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

## 步骤 1：添加手动日志标记

我们需要观察每个操作符是否已被正确测试，所以我们手动添加了日志标记：`// +goat:insert`。

修改后的代码（添加了手动插入标记）：

```go
// 现有代码 ...
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
// 现有代码 ...
```

注意在上述代码中，我们添加了多个 `// +goat:insert` 标记，每个标记后面都有一个 return 语句。

## 步骤 2：运行 goat patch 命令处理标记

```bash
# 初始化 goat 配置（如果尚未初始化）
goat init --app-name "calculator"

# 运行 goat patch 命令处理标记
goat patch
```

命令输出示例：

```
[2025/04/23 23:17:41] INFO: Start to run patch command
[2025/04/23 23:17:41] INFO: Total replaced tracks: 5
[2025/04/23 23:17:41] INFO: Patch applied
[2025/04/23 23:17:41] INFO: ----------------------------------------------------------
[2025/04/23 23:17:41] INFO: ✅ Patch applied successfully!
```

## 步骤 3：查看处理后的代码

运行 `goat patch` 后，所有 `+goat:insert` 标记都被处理，并在标记处插入了跟踪代码：

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

## 步骤 4：使用 goat:delete 标记

假设在测试过程中，我们发现加法和减法已经足够稳定，不再需要观察。我们可以使用 `+goat:delete` 标记来标记要删除的数据点：

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

再次运行 `goat patch` 命令：

```bash
goat patch
```

处理后的代码：

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
	// 现有代码 ...
}
```

## 总结

通过本示例，我们演示了如何使用 `goat patch` 命令处理手动跟踪标记并实现以下功能：

1. 使用 `+goat:insert` 标记在特定位置插入代码，可用于新功能的灰度发布。
2. 使用 `+goat:delete` 标记删除插入的跟踪点，可用于纠正跟踪点。
3. 通过简单的标记添加或删除跟踪点，而无需修改现有代码。

这种手动标记方法为灰度发布提供了更精细的控制，使开发人员能够灵活地启用或禁用特定功能，而无需修改核心代码逻辑。
