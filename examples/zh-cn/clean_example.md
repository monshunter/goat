# goat clean 使用示例

[English](../clean_example.md)

本示例演示如何使用 `goat clean` 命令删除先前由 `goat track` 插入的跟踪代码。

## 场景描述

在灰度发布阶段结束后，或者当需要切换到不同的跟踪策略时，我们需要清理插入的跟踪代码。`goat clean` 命令可以轻松完成这项任务。

## 原始代码（包含跟踪代码）

以下是通过 `goat track` 命令插入跟踪代码后的计算器应用程序代码：

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
注意上述代码包含以下跟踪代码：
```
导入 goat 包
`goat "calculator/goat"`
```
在 main 函数的入口点启动 HTTP 服务
`// +goat:main
 // +goat:tips: do not edit the block between the +goat comments
 goat.ServeHTTP(goat.COMPONENT_0)
 // +goat:end`
```

```
Power 方法中的跟踪调用
`// +goat:generate
 // +goat:tips: do not edit the block between the +goat comments
 goat.Track(goat.TRACK_ID_1)
 // +goat:end`
```

```
main 函数中的跟踪调用
`// +goat:generate
 // +goat:tips: do not edit the block between the +goat comments
 goat.Track(goat.TRACK_ID_2)
 // +goat:end`
```

## 步骤 1：运行 goat clean 命令

当我们完成灰度跟踪且不再需要此代码时，可以使用 `goat clean` 命令将其删除：

```bash
# 确保当前位于项目根目录
cd calculator
```

```bash
# 确认当前路径中的文件
ls
```

当前目录中的文件如下：
```
go.mod  goat  goat.yaml  main.go
```

```bash
# 运行 goat clean 命令
goat clean
```

命令输出示例：

```
[2025/04/23 22:40:45] INFO: Start to run clean command
[2025/04/23 22:40:45] INFO: Total cleaned files: 1
[2025/04/23 22:40:45] INFO: Cleaned project
[2025/04/23 22:40:45] INFO: ----------------------------------------------------------
[2025/04/23 22:40:45] INFO: ✅ Clean completed successfully!
```

## 步骤 2：检查清理后的代码

运行 `goat clean` 后，所有跟踪代码都被删除。检查清理后的代码：

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
# 重新确认当前路径中的文件
ls
```

当前目录中的文件如下：
```
go.mod  goat.yaml  main.go
```

如可见，所有跟踪代码已完全删除，包括：
- `goat` 包的导入
- 跟踪组件变量的声明
- 所有对 `Track` 方法的调用
- `goat` 目录已被删除

## 步骤 3：提交清理后的代码

清理后，我们可以提交代码并将其合并到主分支：

```bash
# 添加更改
git add .

# 提交更改
git commit -m "Remove tracking code with goat clean"

# 切换到主分支
git checkout main

# 合并功能分支
git merge feature/power-function

# 推送到远程仓库
git push origin main
```

## 何时使用 goat clean

`goat clean` 命令在以下情况下特别有用：

1. **金丝雀发布后**：当新功能通过金丝雀测试后，可以删除跟踪代码。
2. **更改跟踪策略**：调整跟踪粒度或方法时，先清理旧的跟踪代码。
3. **准备生产部署**：在开发或测试阶段的跟踪代码在部署到生产环境前删除。
4. **重置跟踪状态**：当发现跟踪代码有问题需要重新生成时。

## 与其他命令的组合使用

`goat clean` 通常与其他 goat 命令结合使用，形成完整的工作流：

1. 使用 `goat init` 初始化配置。
2. 使用 `goat track` 插入跟踪代码。
3. 运行应用程序并收集埋点数据。
4. 使用 `goat clean` 清理跟踪代码。
5. 根据需要调整配置后重复上述过程。

## 总结

通过本示例，我们演示了如何使用 `goat clean` 命令删除跟踪代码并保持代码整洁。此命令提供了一种便捷的方式来管理灰度发布过程中跟踪代码的生命周期，确保在需要时可以轻松添加或删除跟踪代码。
