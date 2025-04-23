# Increament 包

## 概述

`increament`包是一个用于生成埋点跟踪代码的工具包，主要用于在Go应用中添加跟踪点。该包提供了一套完整的API，用于构建、渲染和保存埋点跟踪代码。

## 核心功能

### Values 结构体

`Values` 结构体是该包的核心，它包含了生成埋点代码所需的所有信息：

```go
type Values struct {
    PackageName string      // 生成代码的包名
    Version     string      // 应用版本
    Name        string      // 应用名称
    Components  []Component // 组件列表
    TrackIds    []int       // 埋点ID列表
    Race        bool        // 是否启用竞态条件保护
}

type Component struct {
    ID       int    // 组件ID
    Name     string // 组件名称
    TrackIds []int  // 组件关联的埋点ID
}
```

### 主要方法

- `NewValues(packageName, version, name string, race bool)` - 创建一个新的Values实例
- `AddComponent(id int, name string, trackIds []int)` - 添加组件
- `AddTrackId(id int)` - 添加埋点ID
- `Validate()` - 验证Values的参数是否完整有效
- `MergeValues(other *Values)` - 合并另一个Values对象的内容
- `Clone()` - 创建Values的深拷贝
- `Render()` - 使用内置模板渲染代码，返回字节数组
- `RenderToString()` - 使用内置模板渲染代码，返回字符串
- `RenderWithCustomTemplate(customTemplate string)` - 使用自定义模板渲染
- `Save(outputPath string)` - 将渲染结果保存到文件
- `SaveWithCustomTemplate(outputPath, customTemplate string)` - 使用自定义模板渲染并保存

### 工具函数

- `BuildCustomTemplate(header, body, footer string)` - 从模板片段构建一个完整的模板

## 使用示例

### 基本用法

```go
// 创建一个新的Values实例
values := increament.NewValues(
    "myapp",           // 包名
    "1.0.0",           // 版本号
    "ExampleApp",      // 应用名称
    true,              // 是否启用race条件保护
)

// 添加埋点ID
values.AddTrackId(100)
values.AddTrackId(101)

// 添加组件及其埋点ID
values.AddComponent(1, "LoginComponent", []int{100, 101})

// 渲染并保存到文件
if err := values.Save("./myapp/track.go"); err != nil {
    log.Fatalf("保存文件失败: %v", err)
}
```

### 使用自定义模板

```go
customTemplate := `
package {{.PackageName}}

// 埋点ID常量
const (
    {{range .TrackIds}}
    TRACK_ID_{{.}} = {{.}}
    {{end}}
)
`

// 渲染自定义模板并保存
if err := values.SaveWithCustomTemplate("./myapp/custom_track.go", customTemplate); err != nil {
    log.Fatalf("保存自定义模板失败: %v", err)
}
```

### 合并多个Values对象

```go
values1 := increament.NewValues("myapp", "1.0.0", "App1", true)
values1.AddTrackId(100)

values2 := increament.NewValues("myapp", "1.0.0", "App2", true)
values2.AddTrackId(101)

// 合并values2到values1
values1.MergeValues(values2)
// 现在values1包含了两个TrackID: 100和101
```

## 生成的代码结构

使用默认模板生成的代码包含以下部分：

1. 包声明和导入
2. 应用基本信息（版本和名称）
3. 埋点ID常量定义
4. 埋点状态记录
5. 提供Track函数用于埋点
6. 组件定义和埋点对应关系

## 高级用法

### 构建自定义模板

```go
header := `package tracking

import (
    "log"
)`

body := `
// 埋点函数
func Track(id int) {
    log.Printf("埋点触发: %d", id)
}
`

footer := `// 生成时间: {{now}}`

customTemplate := increament.BuildCustomTemplate(header, body, footer)
```

### 创建Values的深拷贝

```go
original := increament.NewValues("myapp", "1.0.0", "App", true)
original.AddTrackId(100)

// 创建深拷贝
clone := original.Clone()

// 修改原始对象不会影响克隆对象
original.AddTrackId(101)
// clone仍然只有一个TrackID: 100
``` 