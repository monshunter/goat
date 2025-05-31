# 用Echo问答服务掌握代码埋点与追踪技术

## 开场白

在灰度发布的实践中，如何确保新增代码的质量与稳定性？如何精准监控特定功能的执行情况？如何以最小的成本实现代码追踪？

[GOAT（Golang Application Tracing）](https://github.com/monshunter/goat) 作为一款专为Go语言设计的灰度追踪工具，为这些问题提供了优雅的解决方案。本教程将通过一个简单却实用的Echo问答服务，带您全面掌握 [ GOAT ](https://github.com/monshunter/goat)的核心功能。无论您是Go语言开发者、DevOps工程师，还是对灰度发布和代码质量有深度关注的技术管理者，这篇文章都将帮助您将代码追踪技术落地到实际项目中。

## 场景描述

Echo是一个基于命令行的简易问答服务，它将作为我们体验GOAT工具完整流程的理想载体。通过这个实例，我们将学习如何追踪代码变更、监控代码执行路径，以及如何利用 [ GOAT ](https://github.com/monshunter/goat)提供的实时数据做出更明智的发布决策。

## 构建项目

### 初始化项目

1. 首先，我们创建一个基础项目结构，初始化git仓库和Go模块：

```bash
mkdir echo
cd echo && git init && go mod init echo && touch main.go
```

2. 接下来，编写一个简单的命令行问答服务，代码如下：

```go
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Println("欢迎使用查户口Echo服务!")
	fmt.Println("请回答以下问题，输入'退出'随时结束对话")

	scanner := bufio.NewScanner(os.Stdin)
	stage := 1
	userInfo := make(map[string]string)
	totalSteps := 3
	for {
		var prompt string
		var key string
		switch stage {
		case 1:
			key = "姓名"
			prompt = "请问您的姓名是?"
		case 2:
			key = "年龄"
			prompt = "请问您的年龄是?"
		default:
			// 总结阶段
			fmt.Println("\n===== 您的个人信息汇总 =====")
			for k, v := range userInfo {
				fmt.Printf("%s: %s\n", k, v)
			}
			fmt.Println("==========================")
			fmt.Println("还有什么想告诉我的吗? (输入'重新开始'可以重新填写信息)")
		}

		fmt.Print(prompt + " ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		// 检查是否要退出
		if input == "退出" {
			fmt.Println("谢谢使用，再见!")
			break
		}

		// 回显用户输入
		if stage < totalSteps {
			fmt.Printf("您输入的是: %s\n", input)
		}

		// 处理用户输入
		if stage < totalSteps {
			userInfo[key] = input
			stage++
		} else {
			if input == "重新开始" {
				// 清空信息，重新开始
				userInfo = make(map[string]string)
				stage = 1
				fmt.Println("已重置您的信息，请重新开始。")
			} else {
				fmt.Printf("记录您的附加信息: %s\n", input)
				fmt.Println("还有其他想说的吗?")
			}
		}
	}

	// 处理可能的扫描错误
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "读取输入时发生错误:", err)
	}
}
```

3. 验证程序是否正常运行：

```bash
go build -o echo . && ./echo
```

4. 提交初始代码：

```bash
git add .
git commit -m "feat: 添加 echo 命令"
```

### 迭代项目：模拟功能增强

现在我们将通过新增几个问题来模拟实际开发中的功能迭代。新增的问题包括：

- "请问您是哪里人"
- "请问您是做什么工作的"
- "请问您有什么兴趣爱好"

这些变更将作为我们测试 [ GOAT ](https://github.com/monshunter/goat)代码追踪能力的基础。

1. 首先，创建一个新的功能分支：

```bash
git checkout -b feature/additional-questions
```

2. 修改代码，添加新的问答选项：

```go
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Println("欢迎使用查户口Echo服务!")
	fmt.Println("请回答以下问题，输入'退出'随时结束对话")

	scanner := bufio.NewScanner(os.Stdin)
	stage := 1
	userInfo := make(map[string]string)
	totalSteps := 6 // 修改了总步骤数
	for {
		var prompt string
		var key string
		switch stage {
		case 1:
			key = "姓名"
			prompt = "请问您的姓名是?"
		case 2:
			key = "年龄"
			prompt = "请问您的年龄是?"
		// 以下是新增的问答，关于籍贯、职业、兴趣爱好的问题
		case 3:
			key = "籍贯"
			prompt = "请问您是哪里人?"
		case 4:
			key = "职业"
			prompt = "请问您是做什么工作的?"
		case 5:
			key = "兴趣爱好"
			prompt = "请问您的兴趣爱好是?"

		default:
			// 总结阶段
			fmt.Println("\n===== 您的个人信息汇总 =====")
			for k, v := range userInfo {
				fmt.Printf("%s: %s\n", k, v)
			}
			fmt.Println("==========================")
			fmt.Println("还有什么想告诉我的吗? (输入'重新开始'可以重新填写信息)")
		}

		fmt.Print(prompt + " ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		// 检查是否要退出
		if input == "退出" {
			fmt.Println("谢谢使用，再见!")
			break
		}

		// 回显用户输入
		if stage < totalSteps {
			fmt.Printf("您输入的是: %s\n", input)
		}

		// 处理用户输入
		if stage < totalSteps {
			userInfo[key] = input
			stage++
		} else {
			if input == "重新开始" {
				// 清空信息，重新开始
				userInfo = make(map[string]string)
				stage = 1
				fmt.Println("已重置您的信息，请重新开始。")
			} else {
				fmt.Printf("记录您的附加信息: %s\n", input)
				fmt.Println("还有其他想说的吗?")
			}
		}
	}

	// 处理可能的扫描错误
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "读取输入时发生错误:", err)
	}
}
```

3. 验证程序更新是否正常工作：

```bash
go build -o echo && ./echo
```

4. 提交功能迭代代码：

```bash
git add .
git commit -m "feat: 添加额外的用户信息收集功能"
```

## 使用GOAT实现代码追踪

### 初识GOAT工具

首先，让我们了解一下 [ GOAT ](https://github.com/monshunter/goat)的主要命令。执行以下命令查看 [ GOAT ](https://github.com/monshunter/goat)的帮助信息：

```bash
 go install github.com/monshunter/goat/cmd/goat@latest
 goat help
```

输出内容：

```
Goat is a tool for analyzing and instrumenting Go programs

Usage:
  goat [flags]
  goat [command]

Available Commands:
  clean       Clean up instrumentation code in project
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  init        Initialize a new project
  patch       Process manual instrumentation markers in the project
  track       Insert instrumentation code for the project
  version     Print the version number of Goat

Flags:
  -h, --help      help for goat
```

 [ GOAT ](https://github.com/monshunter/goat)的主要工作流程包括：
- `goat init`：初始化项目配置
- `goat track`：自动识别代码变更并插入追踪代码
- `goat patch`：处理手动添加的追踪标记
- `goat clean`：清理所有追踪代码

接下来，我们将一步步体验这个完整流程。

### 步骤1：配置初始化

首先执行初始化命令：

```bash
goat init
```

初始化完成后， [ GOAT ](https://github.com/monshunter/goat)会输出类似如下信息：

```
[2025/04/25 15:21:57] INFO: Start to run init command
[2025/04/25 15:21:57] INFO: initializing project
[2025/04/25 15:21:57] INFO: initializing config: /Users/xxxx/Documents/echo/goat.yaml
[2025/04/25 15:21:57] INFO: project initialized successfully
[2025/04/25 15:21:57] INFO: you can edit '/Users/xxxx/Documents/echo/goat.yaml' to customize configurations according to your needs
```

 [ GOAT ](https://github.com/monshunter/goat)在项目根目录生成了配置文件`goat.yaml`，该文件包含了项目追踪所需的各项配置。其中几个重要配置项包括：

```yaml
# Goat configuration goat.yaml
#
# This file is generated by goat.
# Please do modify when you know what you are doing.
# Every time the field is changed, please rerun "goat patch" to update the tracking results.
#
## App name (default: current directory name)
appName: echo

## App version (default: current .newBranch's commit short hash)
## Before running "goat track" or "goat patch",please confirm the appVersion is the correct version.
## Change when it's necessary.
appVersion: 7f6e79f

## Old branch name (for comparison base)
## Valid values are:
## 1. commit hash
## 2. branch name
## 3. tag name
## 4. ""  or "HEAD" means the current branch
## 5. "INIT" means the current repository is a new one, no old branch
## default: "main"
oldBranch: main

## New branch name (for comparison target)
## Valid values are:
## 1. commit hash
## 2. branch name
## 3. tag name
## 4. "" or "HEAD" means the current branch
## default: "HEAD"
## newBranch must be the same as the current HEAD
newBranch: HEAD

## Ignore files and directories
ignores:
  - .git
  - .gitignore
  - .DS_Store
  - .idea
  - .vscode
  - .venv
  - vendor
  - testdata
  - node_modules
  - goat/goat_generated.go

## Goat package name
goatPackageName: goat

## Goat package alias
goatPackageAlias: goat

## Goat package path, where goat package is installed in your project
goatPackagePath: goat

## Granularity ([line, patch, scope, func], default: patch)
## granularity = line: Track changes at the line level when the line is modified
## granularity = patch: Track changes at the patch(diff patch in the same scope) level when the patch is modified
## granularity = scope: Track changes at the scope level when the scope is modified
## granularity = func: Track changes at the function level when the function is modified
granularity: patch

## Diff precision (1~3, default: 1)
## diffPrecision = 1: Uses git blame to get file change history and generates tracking information based on it,
## highest precision, worst performance
## diffPrecision = 2: Uses git diff to get file change history and generates tracking information (tracks file
## renames or moves), medium precision, normal performance (100 times faster than diffPrecision = 1), in this mode,
## some renamed files may be granted as new files
## diffPrecision = 3: Uses git diff to get file change history and generates tracking information (cannot track
## file renames or moves), lowest precision, best performance (1 ~ 10 times faster than diffPrecision = 2),
## in this mode, all renamed files are granted as new files
diffPrecision: 1

## Threads (default: 1)
threads: 1

## Race (default: false)
## race: true, enable race detection, performance worse
## race: false, disable race detection, performance better
race: false

## Main entries to track (default: all)
## Specify relative paths to main packages from project root
## Examples:
##  - "*" (track all main packages)
##  - "cmd/server" (track only the main package in cmd/server)
##  - "cmd/client,cmd/server" (track multiple specific main packages)
mainEntries:
  - "*"

## Gofmt printer config, same as printer.Config
## Printer config mode, list of (none, useSpaces, tabIndent, sourcePos, rawFormat) (default: "useSpaces,tabIndent")
printerConfigMode:
  - "useSpaces"
  - "tabIndent"

## Printer config tabwidth (default: 8)
printerConfigTabwidth: 8

## Printer config indent (default: 0)
printerConfigIndent: 0

## Data type ([bool, count], default: bool)
dataType: bool

## Verbose output (default: false)
verbose: false

```

这些配置允许您精确控制代码追踪的范围、粒度和性能。

### 步骤2：自动识别代码变更并添加埋点

现在执行追踪命令，让 [ GOAT ](https://github.com/monshunter/goat)自动分析代码变更并添加埋点：

```bash
goat track
```

 [ GOAT ](https://github.com/monshunter/goat)会输出详细的处理日志：

```
[2025/04/25 15:37:27] INFO: Start to run track command
[2025/04/25 15:37:27] INFO: Tracking project
[2025/04/25 15:37:27] INFO: Getting code differences
[2025/04/25 15:37:27] INFO: Getting main package infos
[2025/04/25 15:37:27] INFO: Initializing trackers
[2025/04/25 15:37:27] INFO: Replacing tracks
[2025/04/25 15:37:27] INFO: Replaced 4 tracking points
[2025/04/25 15:37:27] INFO: Saving generated file goat/goat_generated.go
[2025/04/25 15:37:27] INFO: Saving tracking points to 1 files
[2025/04/25 15:37:27] INFO: Applying main entries
[2025/04/25 15:37:27] INFO: Track applied successfully with 4 tracking points
[2025/04/25 15:37:27] INFO: ----------------------------------------------------------
[2025/04/25 15:37:27] INFO: ✅ Track completed successfully!
[2025/04/25 15:37:27] INFO: You can:
[2025/04/25 15:37:27] INFO: - Review the changes using git diff or your preferred diff tool
[2025/04/25 15:37:27] INFO: - Build and test your application to verify instrumentation
[2025/04/25 15:37:27] INFO: - If you manualy add or remove instrumentation, run 'goat patch' to update the instrumentation
[2025/04/25 15:37:27] INFO: - To remove all instrumentation, run 'goat clean'
[2025/04/25 15:37:27] INFO: ----------------------------------------------------------
```

执行`git status`查看文件变更情况：

```bash
git status
```

输出信息：

```
On branch feature/additional-questions
Changes not staged for commit:
  (use "git add <file>..." to update what will be committed)
  (use "git restore <file>..." to discard changes in working directory)
	modified:   main.go

Untracked files:
  (use "git add <file>..." to include in what will be committed)
	goat.yaml
	goat/

no changes added to commit (use "git add" and/or "git commit -a")
```

 [ GOAT ](https://github.com/monshunter/goat)执行后对项目做了两项主要修改：

1. 生成了`goat/goat_generated.go`文件，包含埋点相关的代码和HTTP服务
2. 在`main.go`中插入了埋点追踪代码

通过`git diff main.go`，我们可以看到 [ GOAT ](https://github.com/monshunter/goat)自动在几个关键位置插入了埋点代码：

```diff
diff --git a/main.go b/main.go
index 7dd80ee..3984e88 100644
--- a/main.go
+++ b/main.go
@@ -5,15 +5,24 @@ import (
 	"fmt"
 	"os"
 	"strings"
+	goat "echo/goat"
 )

 func main() {
+	// +goat:main
+	// +goat:tips: do not edit the block between the +goat comments
+	goat.ServeHTTP(goat.COMPONENT_0)
+	// +goat:end
 	fmt.Println("欢迎使用查户口Echo服务!")
 	fmt.Println("请回答以下问题，输入'退出'随时结束对话")

 	scanner := bufio.NewScanner(os.Stdin)
 	stage := 1
 	userInfo := make(map[string]string)
+	// +goat:generate
+	// +goat:tips: do not edit the block between the +goat comments
+	goat.Track(goat.TRACK_ID_1)
+	// +goat:end
 	totalSteps := 6 // 修改了总步骤数
 	for {
 		var prompt string
@@ -27,12 +36,24 @@ func main() {
 			prompt = "请问您的年龄是?"
 		// 以下是新增的问答，关于籍贯、职业、兴趣爱好的问题
 		case 3:
+			// +goat:generate
+			// +goat:tips: do not edit the block between the +goat comments
+			goat.Track(goat.TRACK_ID_2)
+			// +goat:end
 			key = "籍贯"
 			prompt = "请问您是哪里人?"
 		case 4:
+			// +goat:generate
+			// +goat:tips: do not edit the block between the +goat comments
+			goat.Track(goat.TRACK_ID_3)
+			// +goat:end
 			key = "职业"
 			prompt = "请问您是做什么工作的?"
 		case 5:
+			// +goat:generate
+			// +goat:tips: do not edit the block between the +goat comments
+			goat.Track(goat.TRACK_ID_4)
+			// +goat:end
 			key = "兴趣爱好"
 			prompt = "请问您的兴趣爱好是?"
```

 [ GOAT ](https://github.com/monshunter/goat)自动完成了两项关键任务：
1. 在`main`函数开头添加了`goat.ServeHTTP(goat.COMPONENT_0)`，启动用于收集埋点数据的HTTP服务
2. 在新增的代码位置添加了`goat.Track(goat.TRACK_ID_*)`埋点，用于记录代码执行情况

### 步骤3：验证埋点效果

编译并运行添加埋点后的程序：

```bash
go build -o echo . && ./echo
```

程序运行后，会输出埋点服务的启动信息：

```
欢迎使用查户口Echo服务!
请回答以下问题，输入'退出'随时结束对话
请问您的姓名是? 2025/04/25 16:01:49 Goat track service started: http://127.0.0.1:57005
2025/04/25 16:01:49 Goat track all components metrics in prometheus format: http://127.0.0.1:57005/metrics
2025/04/25 16:01:49 Goat track all components details: http://127.0.0.1:57005/track
2025/04/25 16:01:49 Goat track component details: http://127.0.0.1:57005/track?component=COMPONENT_ID
2025/04/25 16:01:49 Goat track components details: http://127.0.0.1:57005/track?component=COMPONENT_ID,COMPONENT_ID2
2025/04/25 16:01:49 Goat track details order by count asc: http://127.0.0.1:57005/track?component=COMPONENT_ID&order=0
2025/04/25 16:01:49 Goat track details order by count desc: http://127.0.0.1:57005/track?component=COMPONENT_ID&order=1
2025/04/25 16:01:49 Goat track details order by id asc: http://127.0.0.1:57005/track?component=COMPONENT_ID&order=2
2025/04/25 16:01:49 Goat track details order by id desc: http://127.0.0.1:57005/track?component=COMPONENT_ID&order=3
```

让我们通过HTTP接口查看埋点状态：

```bash
curl http://127.0.0.1:57005/track
```

初始状态下的埋点数据：

```json
{
  "name": "echo",
  "version": "7f6e79f",
  "results": [
    {
      "id": 0,
      "name": ".",
      "metrics": {
        "version": "78091c90f1a220db24d82de7dcd49e37",
        "total": 4,
        "covered": 1,
        "coveredRate": 25,
        "items": [
          { "id": 2, "name": "TRACK_ID_2", "count": 0 },
          { "id": 3, "name": "TRACK_ID_3", "count": 0 },
          { "id": 4, "name": "TRACK_ID_4", "count": 0 },
          { "id": 1, "name": "TRACK_ID_1", "count": 1 }
        ]
      }
    }
  ]
}
```

从上面的数据可以看出，程序刚启动时，只有`TRACK_ID_1`执行了一次，其余埋点尚未触发。总计有4个埋点，覆盖率为25%。

现在，进行一次完整的问答流程：

```
请问您的姓名是? 张三
您输入的是: 张三
请问您的年龄是? 30
您输入的是: 30
请问您是哪里人? 北京
您输入的是: 北京
请问您是做什么工作的? 软件工程师
```

再次查询埋点状态：

```bash
curl http://127.0.0.1:57005/track
```

更新后的埋点数据：

```json
{
  "name": "echo",
  "version": "7f6e79f",
  "results": [
    {
      "id": 0,
      "name": ".",
      "metrics": {
        "version": "dfe41b436b9a52b2955e42ccde8dddce",
        "total": 4,
        "covered": 3,
        "coveredRate": 75,
        "items": [
          { "id": 4, "name": "TRACK_ID_4", "count": 0 },
          { "id": 1, "name": "TRACK_ID_1", "count": 1 },
          { "id": 2, "name": "TRACK_ID_2", "count": 1 },
          { "id": 3, "name": "TRACK_ID_3", "count": 1 }
        ]
      }
    }
  ]
}
```

现在，前三个埋点都被触发了，覆盖率提升到了75%。这直观地展示了哪些新增代码被执行到，哪些尚未执行。

### 步骤4：手动调整埋点位置

 [ GOAT ](https://github.com/monshunter/goat)不仅支持自动埋点，还允许开发者手动调整埋点位置。现在，我们尝试进行两项调整：

1. 在`case 1`(姓名问题)中添加一个埋点
2. 删除`case 3`(籍贯问题)中的埋点

修改`main.go`文件，在相应位置添加标记：

```go
package main

import (
	"bufio"
	goat "echo/goat"
	"fmt"
	"os"
	"strings"
)

func main() {
	// +goat:main
	// +goat:tips: do not edit the block between the +goat comments
	goat.ServeHTTP(goat.COMPONENT_0)
	// +goat:end
	fmt.Println("欢迎使用查户口Echo服务!")
	fmt.Println("请回答以下问题，输入'退出'随时结束对话")

	scanner := bufio.NewScanner(os.Stdin)
	stage := 1
	userInfo := make(map[string]string)
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_1)
	// +goat:end
	totalSteps := 6 // 修改了总步骤数
	for {
		var prompt string
		var key string
		switch stage {
		case 1:
			// +goat:insert // 插入跟踪点
			key = "姓名"
			prompt = "请问您的姓名是?"
		case 2:
			key = "年龄"
			prompt = "请问您的年龄是?"
		// 以下是新增的问答，关于籍贯、职业、兴趣爱好的问题
		case 3:
			// +goat:delete // 删除跟踪点
			// +goat:tips: do not edit the block between the +goat comments
			goat.Track(goat.TRACK_ID_2)
			// +goat:end
			key = "籍贯"
			prompt = "请问您是哪里人?"
		case 4:
			// +goat:generate
			// +goat:tips: do not edit the block between the +goat comments
			goat.Track(goat.TRACK_ID_3)
			// +goat:end
			key = "职业"
			prompt = "请问您是做什么工作的?"
		case 5:
			// +goat:generate
			// +goat:tips: do not edit the block between the +goat comments
			goat.Track(goat.TRACK_ID_4)
			// +goat:end
			key = "兴趣爱好"
			prompt = "请问您的兴趣爱好是?"

		default:
			// 总结阶段
			fmt.Println("\n===== 您的个人信息汇总 =====")
			for k, v := range userInfo {
				fmt.Printf("%s: %s\n", k, v)
			}
			fmt.Println("==========================")
			fmt.Println("还有什么想告诉我的吗? (输入'重新开始'可以重新填写信息)")
		}

		fmt.Print(prompt + " ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		// 检查是否要退出
		if input == "退出" {
			fmt.Println("谢谢使用，再见!")
			break
		}

		// 回显用户输入
		if stage < totalSteps {
			fmt.Printf("您输入的是: %s\n", input)
		}

		// 处理用户输入
		if stage < totalSteps {
			userInfo[key] = input
			stage++
		} else {
			if input == "重新开始" {
				// 清空信息，重新开始
				userInfo = make(map[string]string)
				stage = 1
				fmt.Println("已重置您的信息，请重新开始。")
			} else {
				fmt.Printf("记录您的附加信息: %s\n", input)
				fmt.Println("还有其他想说的吗?")
			}
		}
	}

	// 处理可能的扫描错误
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "读取输入时发生错误:", err)
	}
}
```

现在执行`goat patch`命令，处理这些手动标记：

```bash
goat patch
```

输出结果：

```
[2025/04/25 16:30:14] INFO: Start to run patch command
[2025/04/25 16:30:14] INFO: Patching project
[2025/04/25 16:30:14] INFO: Getting main package infos
[2025/04/25 16:30:14] INFO: Preparing files
[2025/04/25 16:30:14] INFO: Applying patch
[2025/04/25 16:30:14] INFO: Total replaced tracks: 4
[2025/04/25 16:30:14] INFO: Patch applied
[2025/04/25 16:30:14] INFO: ----------------------------------------------------------
[2025/04/25 16:30:14] INFO: ✅ Patch applied successfully!
[2025/04/25 16:30:14] INFO: Manual markers have been processed (// +goat:delete, // +goat:insert)
[2025/04/25 16:30:14] INFO: You can:
[2025/04/25 16:30:14] INFO: - Review the changes using git diff or your preferred diff tool
[2025/04/25 16:30:14] INFO: - Build and test your application to verify the changes
[2025/04/25 16:30:14] INFO: - Add more manual markers and run 'goat fix' again if needed
[2025/04/25 16:30:14] INFO: - If you want to remove all instrumentation, run 'goat clean'
[2025/04/25 16:30:14] INFO: ----------------------------------------------------------
```

查看更新后的`main.go`文件，可以发现：

```go
// Omit the existing code ...
for {
		var prompt string
		var key string
		switch stage {
		case 1:
			// +goat:generate
			// +goat:tips: do not edit the block between the +goat comments
			goat.Track(goat.TRACK_ID_2)
			// +goat:end
			key = "姓名"
			prompt = "请问您的姓名是?"
		case 2:
			key = "年龄"
			prompt = "请问您的年龄是?"
		// 以下是新增的问答，关于籍贯、职业、兴趣爱好的问题
		case 3:
			key = "籍贯"
			prompt = "请问您是哪里人?"
		case 4:
			// +goat:generate
			// +goat:tips: do not edit the block between the +goat comments
			goat.Track(goat.TRACK_ID_3)
			// +goat:end
			key = "职业"
			prompt = "请问您是做什么工作的?"
		case 5:
			// +goat:generate
			// +goat:tips: do not edit the block between the +goat comments
			goat.Track(goat.TRACK_ID_4)
			// +goat:end
			key = "兴趣爱好"
			prompt = "请问您的兴趣爱好是?"

		default:
			// 总结阶段
			fmt.Println("\n===== 您的个人信息汇总 =====")
			for k, v := range userInfo {
				fmt.Printf("%s: %s\n", k, v)
			}
			fmt.Println("==========================")
			fmt.Println("还有什么想告诉我的吗? (输入'重新开始'可以重新填写信息)")
		}

		fmt.Print(prompt + " ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		// 检查是否要退出
		if input == "退出" {
			fmt.Println("谢谢使用，再见!")
			break
		}

		// 回显用户输入
		if stage < totalSteps {
			fmt.Printf("您输入的是: %s\n", input)
		}

		// 处理用户输入
		if stage < totalSteps {
			userInfo[key] = input
			stage++
		} else {
			if input == "重新开始" {
				// 清空信息，重新开始
				userInfo = make(map[string]string)
				stage = 1
				fmt.Println("已重置您的信息，请重新开始。")
			} else {
				fmt.Printf("记录您的附加信息: %s\n", input)
				fmt.Println("还有其他想说的吗?")
			}
		}
	}
	// Omit the existing code ...
```

- 在`case 1`中添加了新的埋点代码
- 删除了`case 3`中的埋点代码

这展示了 [ GOAT ](https://github.com/monshunter/goat)对手动精细调整埋点位置的灵活支持。

### 步骤5：清理埋点代码

在完成测试或发布后，您可能希望清理所有埋点代码。 [ GOAT ](https://github.com/monshunter/goat)提供了简单的清理命令：

```bash
goat clean
```

执行结果：

```
[2025/04/25 16:38:07] INFO: Start to run clean command
[2025/04/25 16:38:07] INFO: Cleaning project
[2025/04/25 16:38:07] INFO: Preparing files
[2025/04/25 16:38:07] INFO: Prepared 1 files
[2025/04/25 16:38:07] INFO: Cleaning contents
[2025/04/25 16:38:07] INFO: Total cleaned files: 1
[2025/04/25 16:38:07] INFO: Cleaned project
[2025/04/25 16:38:07] INFO: ----------------------------------------------------------
[2025/04/25 16:38:07] INFO: ✅ Clean completed successfully!
[2025/04/25 16:38:07] INFO: All instrumentation code has been removed from your project.
[2025/04/25 16:38:07] INFO: You can:
[2025/04/25 16:38:07] INFO: - Review the changes using git diff or your preferred diff tool
[2025/04/25 16:38:07] INFO: - Build and test your application to verify clean up
[2025/04/25 16:38:07] INFO: - If you want to reapply instrumentation, run 'goat patch'
[2025/04/25 16:38:07] INFO: ----------------------------------------------------------
```
使用 `git status` 查看工作区域：

```bash
git status
```

状态如下，可以看到工作区已经恢复
```
On branch feature/additional-questions
Untracked files:
  (use "git add <file>..." to include in what will be committed)
	goat.yaml

nothing added to commit but untracked files present (use "git add" to track)
```


此命令会移除项目中所有 [ GOAT ](https://github.com/monshunter/goat)相关的埋点代码，确保代码干净、轻量，适合生产环境部署。

## 总结与应用场景

通过这个Echo问答服务示例，我们完整体验了 [ GOAT ](https://github.com/monshunter/goat)工具的核心功能与工作流程：

1. **初始化配置**：使用`goat init`建立项目配置，精确控制追踪粒度和范围
2. **自动代码埋点**：通过`goat track`智能识别代码变更并自动插入追踪代码
3. **数据实时监控**：通过内置HTTP服务实时查看代码执行覆盖情况
4. **手动埋点调整**：利用`goat patch`灵活添加或删除特定位置的埋点
5. **代码清理恢复**：使用`goat clean`快速移除所有埋点代码

 [ GOAT ](https://github.com/monshunter/goat)在以下场景中尤其有价值：

- **灰度发布**：监控新版本中各项功能的实际执行情况，为扩大灰度比例提供数据支持
- **功能验证**：确保增量代码在真实环境中得到充分测试和验证
- **性能分析**：精准定位代码执行路径，辅助识别性能瓶颈
- **bug排查**：追踪代码执行流程，快速定位异常点
- **代码重构**：确保重构后的代码路径全面覆盖，提升代码质量

作为一款专为Go语言设计的代码追踪工具， [ GOAT ](https://github.com/monshunter/goat)以其轻量、高效、精准的特性，为Go项目的质量保障和平稳发布提供了强有力的支持。它不仅简化了代码埋点的流程，还通过实时数据帮助开发团队做出更明智的决策。

无论您是单体应用还是微服务架构，是初创项目还是大型系统， [ GOAT ](https://github.com/monshunter/goat)都能以最小的侵入性为您的灰度发布保驾护航。

尝试将 [ GOAT ](https://github.com/monshunter/goat)集成到您的开发流程中，体验数据驱动的灰度发布新模式！
