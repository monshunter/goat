# GOAT - Golang应用灰度追踪工具

## 1. 项目概述

### 1.1 背景

在对软件应用进行灰度发布（如红蓝部署、金丝雀发布）过程中，开发和运维人员通常依赖外部指标（如报错率、业务指标、资源消耗等）来决定是否推进灰度流程。然而，从第一性原理考虑，稳健的决策依据应当是由内而外的：确保灰度过程中，所有增量功能都被覆盖测试，且外部指标保持在预期范围内。

GOAT (Golang Application Tracing) 项目旨在通过自动化埋点方式，提供可靠的内部依据，帮助开发人员评估灰度发布的覆盖情况，从而做出更加安全、可靠的灰度推进决策。

### 1.2 项目目标

- 开发一个自动化埋点工具（命令行工具），为Go语言项目提供增量代码执行追踪能力
- 提供简单易用的API，便于应用程序集成埋点功能
- 通过内嵌HTTP服务，实时展示埋点覆盖状态
- 支持开发人员自定义埋点策略
- 最小化埋点代码对应用性能的影响

## 2. 功能需求

### 2.1 核心功能

#### 2.1.1 有效增量代码的识别

**有效增量代码**定义为：本次发布相对于稳定版本的、可以被执行的函数体或方法体内部的新增或修改代码。

以下内容**不**属于有效增量代码：
- 被删除的代码
- 非 *.go 文件的增量代码
- 测试文件（*_test.go）中的增量代码
- 新增的注释、空行
- 行尾新增的注释（代码本身未变）
- 函数/方法体外部的增量代码（如全局常量、变量、类型、接口和函数声明等）
- 函数/方法内部的类型声明
- 移动的文件或者重命名的文件

#### 2.1.2 逻辑分支的识别与埋点

GOAT将识别并对以下类型的逻辑分支进行埋点：

**显式分支**：
- if-else 分支
- switch-case 分支
- select-case 分支

**隐式分支**：
- 函数体中连续的非分支语句块

#### 2.1.3 埋点规则

- **前置埋点原则**：在分支代码块的开始处插入埋点代码
- **单次埋点原则**：每个逻辑分支只进行一次埋点
- **条件变更特殊处理**：当分支条件（如if条件表达式）发生变更时，需对其影响的所有一级分支额外插入埋点
- **空分支处理**：空select{}或空switch{}视为普通语句，使用前置埋点

#### 2.1.4 埋点覆盖状态追踪

- 通过内嵌HTTP服务提供埋点状态查询接口
- 支持按组件、文件、函数等维度聚合埋点覆盖状态
- 支持埋点覆盖率统计

### 2.2 命令行工具

#### 2.2.1 工具名称

```
goat
```

#### 2.2.2 子命令

1. **init** - 插入埋点

```
goat init <project> --stable master --publish "release-1.32"
```

2. **patch** - 插入埋点

```
goat patch <project> 
```

参数说明：
- `<project>`：目标项目的路径，即/path/to/project

3. **fix** - 修复埋点

```
goat fix <project>
```

4. **clean** - 清理埋点

```
goat clean <project>
```

### 2.3 埋点API

GOAT将提供简洁的埋点API，自动生成在项目中的`/goat`目录下（可通过GOAT_DIR环境变量修改）。

#### 2.3.1 API结构

```go
package goat

// 埋点状态记录，使用切片存储以提高性能
var embeddings []bool

// 埋点ID类型
type EmbeddingID int

// 应用组件类型
type Composer int

// 埋点函数，用于记录执行路径
func Track(id EmbeddingID)

// 启动HTTP服务，展示埋点状态
func ServeHTTP(composer Composer)
```

#### 2.3.2 自动生成的配置

工具会自动生成以下内容：
- 埋点ID常量定义
- 组件与埋点ID的映射关系
- HTTP服务端点处理函数

## 3. 埋点示例

### 3.1 简单语句埋点

**原始代码**:
```go
func example1() {
    x, y, z := 0, 1, 2
    x, y, z = x + y, y + z, z + x
}
```

**变更代码**:
```go
func example1() {
    x, y, z := 0, 1, 2
    x, y, z = x + y, y + z, z + x
    fmt.Println("分支1")
    fmt.Println(x, y, z)
}
```

**埋点后代码**:
```go
func example1() {
    x, y, z := 0, 1, 2
    x, y, z = x + y, y + z, z + x
    // +goat
    goat.Track(goat.EmbeddingID_1)
    fmt.Println("分支1")
    fmt.Println(x, y, z)
}
```

### 3.2 条件分支埋点

**原始代码**:
```go
func example2() {
    x, y, z := 0, 1, 2
    x, y, z = x + y, y + z, z + x
    if x + y + z == 0 {
        fmt.Println(x + y + z)
    }
}
```

**变更代码**:
```go
func example2() {
    x, y, z := 0, 1, 2
    x, y, z = x + y, y + z, z + x
    fmt.Println("分支1")
    if x + y + z != 0 {
        fmt.Println(x + y + z)
        fmt.Println("分支2")
    } else {
        fmt.Println("分支3")
    }
    fmt.Println("分支4")
    fmt.Println(x, y, z)
}
```

**埋点后代码**:
```go
func example2() {
    x, y, z := 0, 1, 2
    x, y, z = x + y, y + z, z + x
    // +goat
    goat.Track(goat.EmbeddingID_1)
    fmt.Println("分支1")
    if x + y + z != 0 {
        // +goat - 条件变更埋点
        goat.Track(goat.EmbeddingID_2)
        // +goat - 代码块变更埋点
        goat.Track(goat.EmbeddingID_3)
        fmt.Println(x + y + z)
        fmt.Println("分支2")
    } else {
        // +goat - 条件变更埋点
        goat.Track(goat.EmbeddingID_4)
        // +goat - 代码块变更埋点
        goat.Track(goat.EmbeddingID_5)
        fmt.Println("分支3")
    }
    // +goat
    goat.Track(goat.EmbeddingID_6)
    fmt.Println("分支4")
    fmt.Println(x, y, z)
}
```

## 4. 实现方案

### 4.1 技术路线

- 使用Go语言AST解析实现代码分析与埋点
- 利用Git获取代码差异（使用go-git实施，避免在代码中调用git diff或者git blame 等外部命令）
- 使用标准库实现HTTP服务
- 零依赖设计，最小化对用户项目的影响
- 使用goat patch命令时，应新建一个分支，避免污染原分支

### 4.2 埋点文件结构

**示例：/goat/embeddings.go**

```go
package goat

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "sync/atomic"
)

const AppVersion = "v1.0.0"  // 应用版本
const AppName = "example-app"  // 应用名称

type EmbeddingID = int

const (
    EmbeddingID_Start = iota
    EmbeddingID_1
    EmbeddingID_2
    EmbeddingID_3
    EmbeddingID_4
    // ...其他埋点ID
    EmbeddingID_End
)

// 被手动删除的埋点ID列表
var InvalidEmbeddingID = []EmbeddingID{
    // 用户手动删除的埋点ID会记录在这里
}

// 埋点ID名称
var EmbeddingIDNames []string

// 埋点状态记录 - 使用切片替代map以提高性能
var embeddings []int32

// 初始化埋点
func init() {
    EmbeddingIDNames = make([]string, EmbeddingID_End+1)
    for i := 1; i <= EmbeddingID_End; i++ {
        EmbeddingIDNames[i] = fmt.Sprintf("EmbeddingID_%d", i)
    }

    // 初始化埋点状态切片
    embeddings = make([]int32, EmbeddingID_End+1)
}

// 埋点函数
func Track(id EmbeddingID) {
    if id > 0 && id < EmbeddingID_End {
        atomic.StoreInt32(&embeddings[id], 1)
    }
}

// 应用组件类型
type Composer = int

const (
    _ = iota
    ComposerBin_1  // 组件1
    ComposerBin_2  // 组件2
)

// 组件1的埋点ID列表
var ComposerBin_1_EmbeddingID = []EmbeddingID{
    EmbeddingID_1,
    EmbeddingID_2,
}

// 组件2的埋点ID列表
var ComposerBin_2_EmbeddingID = []EmbeddingID{
    EmbeddingID_3,
    EmbeddingID_4,
}

// 组件与埋点ID的映射关系
var ComposersEmbeddingID = [][]EmbeddingID{
    ComposerBin_1: ComposerBin_1_EmbeddingID,
    ComposerBin_2: ComposerBin_2_EmbeddingID,
}

// 启动HTTP服务
func ServeHTTP(composer Composer) {
    go func() {
        system := http.NewServeMux()
        system.HandleFunc("/", homeHandler)
        system.HandleFunc("/health", healthHandler)
        system.HandleFunc("/embeddings", embeddingsHandler)

        port := ":8080"
        fmt.Printf("Goat埋点服务已启动: http://localhost%s\n", port)
        log.Fatal(http.ListenAndServe(port, system))
    }()
}

// 主页处理函数
func homeHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "<h1>GOAT埋点状态</h1><p>请访问 <a href='/embeddings'>/embeddings</a> 查看详细埋点状态</p>")
}

// 健康检查处理函数
func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, `{"status":"healthy","version":"%s","app":"%s"}`, AppVersion, AppName)
}

// 埋点状态处理函数
func embeddingsHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)

    // 生成埋点状态JSON
    details := make(map[string]bool)
    covered := 0

    for id := EmbeddingID_Start + 1; id < EmbeddingID_End; id++ {
        isTracked := atomic.LoadInt32(&embeddings[id]) == 1
        details[EmbeddingIDNames[id]] = isTracked
        if isTracked {
            covered++
        }
    }

    result := map[string]interface{}{
        "total": EmbeddingID_End - EmbeddingID_Start - 1,
        "covered": covered,
        "details": details,
    }

    // 输出JSON
    jsonData, _ := json.Marshal(result)
    w.Write(jsonData)
}
```

### 4.3 集成示例

**在main函数中集成**:

```go
// bin/ComposerBin_1/main.go
func main() {
    // 启动埋点HTTP服务
    goat.ServeHTTP(goat.ComposerBin_1)

    // 业务代码...
}
```

## 5. 未来拓展

- 支持更多语言（Java、Python等）
- 提供可视化埋点覆盖率面板
- 支持与CI/CD系统集成
- 支持与APM（应用性能监控，Application Performance Monitoring）系统集成，如Prometheus、Grafana、Datadog等，实现埋点数据与应用性能数据的关联分析
- 提供更多埋点策略选项

## 6. 注意事项

- 埋点代码应尽可能轻量，避免影响应用性能
- 埋点不应修改原有业务逻辑
- 用户可通过注释手动排除不需要埋点的代码块
- 生产环境应谨慎使用HTTP服务，建议设置访问控制
