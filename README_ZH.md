# GOAT - Golang应用灰度追踪工具

[![Go Report Card](https://goreportcard.com/badge/github.com/monshunter/goat)](https://goreportcard.com/report/github.com/monshunter/goat)
[![GoDoc](https://godoc.org/github.com/monshunter/goat?status.svg)](https://godoc.org/github.com/monshunter/goat)
[![License](https://img.shields.io/github/license/monshunter/goat)](https://github.com/monshunter/goat/blob/main/LICENSE)

[English](README.md)

## 📖 简介

`GOAT`（Golang Application Tracing）是一个高性能的灰度发布代码追踪工具，专为Go语言应用设计。它能够自动识别和追踪增量代码的执行情况，帮助开发人员在灰度发布过程中做出更可靠的决策。通过自动化埋点和实时追踪，GOAT提供了内部依据，确保灰度过程中的增量功能得到充分覆盖测试。

## 🚀 功能

* 自动识别有效增量代码，精确定位修改点
* 智能埋点系统，支持显式和隐式分支的追踪
* 提供多种粒度的代码跟踪能力（行级、补丁级、作用域级、函数级）
* 支持多种差异精度模式，适应不同复杂度的代码变更
* 内嵌HTTP服务，实时展示埋点覆盖状态
* 资源高效利用，对应用性能影响最小化
* 简单易用的命令行工具和API接口
* 多线程支持，提升处理速度
* 支持自定义埋点策略

## 💡 GOAT 工作原理

### 工作流程

1. **初始化**：配置项目参数并生成配置文件
2. **差异分析**：分析稳定分支和发布分支之间的代码差异
3. **智能埋点**：在增量代码中自动插入追踪代码
4. **运行监控**：在应用运行过程中收集埋点执行数据
5. **状态展示**：通过HTTP接口展示埋点覆盖状况

## 🧰 安装

### 方法一：使用Go Install安装（推荐）：

```bash
go install github.com/monshunter/goat/cmd/goat@latest
```

确保您的`$GOPATH/bin`目录已添加到系统PATH中。

### 方法二：从源码构建并安装：

```bash
git clone https://github.com/monshunter/goat.git
cd goat
make install
```

这将编译二进制文件并将其安装到`$GOPATH/bin`目录中。

### 方法三：仅构建而不安装：

```bash
git clone https://github.com/monshunter/goat.git
cd goat
make build
```

构建的二进制文件将位于`bin`目录下。

## 🛠 使用

### 初始化项目

在Go项目根目录下执行：

```bash
goat init
```

这将生成默认配置文件`goat.yaml`。您可以自定义配置选项：

```bash
goat init --old main --new HEAD --app-name "my-app" --granularity func
```

### 配置选项

通过在调用`goat init`时使用各种选项，可以自定义配置：

```bash
goat init --help
```

常用选项包括：
- `--old <oldBranch>`: 稳定分支（默认："main"）
- `--new <newBranch>`: 发布分支（默认："HEAD"）
- `--app-name <appName>`: 应用名称（默认："example-app"）
- `--granularity <granularity>`: 粒度（line, patch, scope, func）（默认："patch"）
- `--diff-precision <diffPrecision>`: 差异精度（1~3）（默认：1）
- `--threads <threads>`: 线程数（默认：1）
- `--ignores <ignores>`: 忽略的文件/目录列表，逗号分隔

### 插入追踪代码

```bash
goat track
```

这将分析项目中的增量代码并自动插入追踪埋点。运行此命令后，可以：
- 使用git diff或其他工具查看变更
- 构建和测试您的应用程序以验证埋点
- 如果项目已经有追踪代码，请先运行`goat clean`清理

### 处理手动埋点标记

```bash
goat patch
```

此命令用于处理项目中的手动埋点标记，主要处理：
- `// +goat:delete`标记 - 删除标记为删除的代码段
- `// +goat:insert`标记 - 在标记位置插入代码
如果您手动添加或移除了埋点，可以运行此命令更新埋点实现。

### 清理追踪代码

```bash
goat clean
```

移除所有已插入的追踪代码。

### 查看版本信息

```bash
goat --version
```

## 📚 示例

GOAT 提供了多个详细的使用示例，帮助您更好地理解其功能：

1. [Track 命令示例](examples/zh-cn/track_example.md) - 如何跟踪代码变更并插入跟踪代码
2. [Patch 命令示例](examples/zh-cn/patch_example.md) - 如何处理手动跟踪标记
3. [Clean 命令示例](examples/zh-cn/clean_example.md) - 如何清理跟踪代码
4. [粒度示例](examples/zh-cn/granularity_example.md) - 演示不同粒度下的代码跟踪

更多示例请查看[示例目录](examples/zh-cn/)。

## 🖥 应用场景

- 灰度发布（蓝绿部署、金丝雀发布）中的代码覆盖追踪
- 新功能的执行路径监控
- 重构代码的验证测试
- 性能变更的影响分析
- 微服务架构中的服务升级追踪

## 🔋 开发环境要求

- Go 1.21+
- Git

## 📊 埋点数据监控

### HTTP服务

通过GOAT插入埋点代码后，在您的应用程序运行时，会自动启动一个HTTP服务，提供实时的埋点覆盖状态。默认情况下，该服务在端口`57005`上运行。

您可以通过设置环境变量`GOAT_PORT`来自定义端口：

```bash
export GOAT_PORT=8080
```

### API端点

GOAT提供了以下API端点用于查询埋点覆盖状态：

#### 1. 获取所有组件的埋点状态

```
GET http://localhost:57005/metrics
```

#### 2. 获取特定组件的埋点状态

```
GET http://localhost:57005/metrics?component=COMPONENT_ID
```

其中`COMPONENT_ID`是组件的ID（通常从0开始的整数）或组件名称。

#### 3. 获取多个组件的埋点状态

```
GET http://localhost:57005/metrics?component=COMPONENT_ID1,COMPONENT_ID2
```

#### 4. 按照不同顺序排序结果

```
# 按照执行次数升序排序
GET http://localhost:57005/metrics?component=COMPONENT_ID&order=0

# 按照执行次数降序排序
GET http://localhost:57005/metrics?component=COMPONENT_ID&order=1

# 按照ID升序排序
GET http://localhost:57005/metrics?component=COMPONENT_ID&order=2

# 按照ID降序排序
GET http://localhost:57005/metrics?component=COMPONENT_ID&order=3
```

### 响应格式

API返回JSON格式的响应，包含以下信息：

```json
{
  "name": "example-app", 
  "version": "1.0.0",
  "results": [
    {
      "id": 0,   // 组件id
      "name": "组件名称", // 组件名字
      "metrics": {
        "total": 10,         // 总埋点数
        "covered": 5,         // 已覆盖的埋点数
        "coveredRate": 50,    // 覆盖率（百分比）
        "items": [
          {
            "id": 1,          // 埋点ID
            "name": "TRACK_ID_1", // 埋点名称
            "count": 3        // 执行次数
          }
          // 更多埋点...
        ]
      }
    }
    // 更多组件...
  ]
}
```

### 使用示例

1. 使用curl查看所有埋点状态：

```bash
curl http://localhost:57005/metrics | jq
```

2. 使用curl查看特定组件的埋点状态：

```bash
curl http://localhost:57005/metrics?component=0 | jq
```

### 观测和分析

1. **实时监控**：在应用运行时，可以随时查看埋点覆盖情况
2. **灰度决策**：根据埋点覆盖率，评估是否进行下一步灰度发布
3. **问题分析**：找出未被执行的代码路径，定位潜在问题
4. **覆盖报告**：生成覆盖率报告，用于团队评审和质量保证

## 🌐 环境变量

GOAT项目支持通过环境变量进行配置，下表列出了所有可用的环境变量及其作用：

| 环境变量 | 描述 | 默认值 | 使用场景 |
| --- | --- | --- | --- |
| `GOAT_PORT` | 设置埋点HTTP服务的端口 | `57005` | 当默认端口被占用或需要自定义端口时 |
| `GOAT_METRICS_IP` | 设置埋点HTTP服务绑定的IP地址 | `127.0.0.1` | 当需要从非本机访问埋点服务时，可设置为`0.0.0.0` |
| `GOAT_CONFIG` | 指定配置文件的路径 | `goat.yaml` | 当需要使用非默认位置的配置文件时 |
| `GOAT_STACK_TRACE` | 是否在发生致命错误时显示堆栈跟踪 | `false` | 调试问题时设置为`1`或`true`或`yes` |

### 环境变量使用示例

1. 修改HTTP服务端口：

```bash
export GOAT_PORT=8080
```

2. 允许从其他机器访问埋点服务：

```bash
export GOAT_METRICS_IP=0.0.0.0
```

3. 使用自定义配置文件路径：

```bash
export GOAT_CONFIG=/path/to/custom-goat.yaml
```

4. 启用错误堆栈跟踪：

```bash
export GOAT_STACK_TRACE=1
```

## 🏷️ 标记说明

GOAT使用特殊的代码注释标记（标记以`// +goat:`开头）来控制代码的插入和删除。标记分为两类：用户可用标记和内部标记。

### 用户可用标记

以下标记可供开发者使用：

| 标记 | 描述 | 使用场景 | 状态 |
| --- | --- | --- | --- |
| `// +goat:delete` | 标记需要删除的代码块开始 | 当需要删除一段代码时使用 | 启用 |
| `// +goat:insert` | 标记需要插入代码的位置 | 手动指定埋点插入位置 | 启用 |

### 内部标记

以下标记由GOAT内部使用，用户不应手动添加：

| 标记 | 描述 | 使用场景 | 状态 |
| --- | --- | --- | --- |
| `// +goat:generate` | 标记埋点代码生成的开始 | 自动生成的埋点代码块的开始标记 | 启用 |
| `// +goat:tips: ...` | 提示信息 | 向开发者提供关于代码块的提示 | 启用 |
| `// +goat:main` | 标记主函数入口埋点 | 在main函数中添加HTTP服务启动代码 | 启用 |
| `// +goat:end` | 标记代码块的结束 | 所有`+goat:`标记块的结束标记 | 启用 |
| `// +goat:import` | 标记导入部分 | 用于标记埋点相关的导入语句 | 未启用 |
| `// +goat:user` | 标记用户自定义埋点 | 用户自定义的埋点代码 | 尚未支持 |

### 注意事项

1. **删除代码块**：如果将`// +goat:generate`改为`// +goat:delete`，然后执行`goat patch`命令，从`// +goat:delete`到`// +goat:end`之间的代码将被删除。

   ```go
   // +goat:delete
   // +goat:tips: do not edit the block between the +goat comments
   goat.Track(goat.TRACK_ID_1)
   // +goat:end
   ```

2. **插入埋点**：在需要手动插入埋点的位置添加：

   ```go
   // +goat:insert
   ```

   执行`goat patch`后，该位置将插入埋点代码。

3. **标记生效时机**：这些标记在执行`goat patch`命令时被处理，而非`goat track`命令。

4. **标记的嵌套**：标记不支持嵌套使用，每个标记块必须完整以`// +goat:end`结束。

## 📄 许可证

`GOAT`的源码基于[MIT许可证](LICENSE)开源。


## 💎 贡献

欢迎贡献代码或提出建议！请查看[贡献指南](CONTRIBUTING_ZH.md)了解更多信息。

## ☕️ 支持

如果您发现GOAT对您有所帮助，可以通过以下方式支持项目：

- 在GitHub上给项目点星
- 提交Pull Request添加新功能或修复bug
- 向他人推荐这个项目