# 需求或者项目： goat
## 背景
- 在对软件应用或者服务进行灰度发布（如： 红蓝部署、金丝雀发布）的过程中，运维或者开发人员推进灰度流程的决策中往往是依赖外部指标（如：报错率、业务指标、内存cpu等资源指标）， 但是从第一性原理考虑，这些外部指标，并不能作为推进灰度的全部因素。
- 根据第一性原理，稳健的决策依据应该是由内而外的：在灰度过程中，软件的所有增量功能都被覆盖，而外部指标都在预期范围内。
- 外部指标因业务而异，很难有一个通用的评估标准，不是本项目的讨论范围
- 本项目注意考虑如何从内部为推进灰度进行的决策提供依据
- 以下是方案的草稿
## 目标
- 为通过golang构建的项目提供自动化埋点工具（也就是它至少是一个命令行工具）
- 埋点的范围： 有效的增量代码
- 埋点的颗粒度： 逻辑分支（如if、 select、switch  ）
- 分支埋点规则：
   - 每个逻辑分支只进行一次埋点，且是前置埋点， 埋点位置在分支body内部
   - 但是condition的变更，需要对其影响的所有一级分支额外插入一个埋点
   - 空select{} 或者空switch{} 等同于普通语句，不能作为条件分支处理，所以新增的空select{} 或者空switch{}都是前置埋点，以免语法错误
- 埋点代码的作用： 仅用于判断增量代码是否被执行，不关心增量代码的执行效率
- 埋点后的应用应提供http服务，可以被开发人员查看或者外部服务查看或者采集埋点数据的状态
- 插入的埋点代码应该尽量保持简洁，无锁，可以轻易被开发人员识别
- 提供途径让开发人员决定是否对某部分代码进行埋点处理
## 方案

### 概念定义
- Q： 什么是有效的增量代码？
- A： 本次发布相对于稳定版本的可以被执行的function内部或者method内部的增量代码。所以以下代码都不是有效的增量代码：
    - 所有被删除的代码
    - 所有非 *.go文件的增量代码
    - 所有*_test.go 的增量代码
    - 所有新增的注释、空行
    - 所有符合本行代码没有改变，但是这一行的尾部增加了注释
    - 所有不是在function或者method body内部的产生的增量代码，如全局常量、变量、类型、interface、function等声明
    - 所有在function或者method内部的type声明


- Q：什么是逻辑分支？
- A：分支分类如下：
  - 显式分支： if - else、 switch-case、 select-case 分支
  - 隐式分支： 两个同一水平的显式分支之间的代码，下面将通过一个例子说明：
  ```golang
  func example() {
    fmt.Println("这是分支0")
    if x == 1 {
        fmt.Println("这是分支1")
    }else {
        fmt.Println("这是分支2")
    }
    fmt.Println("这是分支3")
    fmt.Println("这是分支3")
    fmt.Println("这是分支3")
    var x int
    switch x {
        case 0:
            fmt.Println("这是分支4")
        case 1:
            fmt.Println("这是分支5")
        default: 
            fmt.Println("这是分支6")
            if x == 2 {
                fmt.Println("这是分支7")
                if x * x == 4 {
                    fmt.Println("这是分支8")
                }
                fmt.Println("这是分支9")
            }else{
                fmt.Println("这是分支10")
            }
            fmt.Println("这是分支11")
            fmt.Println("这是分支11")
    }

    fmt.Println("这是分支12")
    fmt.Println("这是分支12")
    fmt.Println("这是分支12")
    
  }
  ```

### 命令行工具定义及其用法

- 这是命令行工具的名字
  - goat
  
- 插入埋点子命令 
  - goat patch $project $stableBranch $publishBranch --newBranch $branchName --bin $bin

- 修复埋点子命令
  - goat fixed $project
  
- 清理埋点 
  - goat clean $project 

- args或者flag说明
  - $project： 目标项目的路径，即/path/to/project  
  - $stableBranch： 稳定运行的分支名字或者commit
  - $publishBranch：用于进行灰度发布的分支或者commit
  - --newBranch $branchName： 通过 $publishBranch 生成的隔离分支，用于进行埋点代码插入，如果未指定则生成 $publishBranch-goat
  - --bin $bin： 指定需要为 $project 中的哪个二进制执行埋点，如未指定，默认为所有的 func main初始化
### 埋点举例
#### example-1
- 原始代码
  ```golang
    func example1() {
        x, y, z := 0,1,2
        x, y, z = x + y, y + z, z + x
    }
  ```
- 变更代码
   ```golang
    func example1() {
        x, y, z := 0,1,2
        x, y, z = x + y, y + z, z + x
        fmt.Println("分支1")
        fmt.Println(x,y, z)
    }
  ```
- 正确埋点的变更代码： 前置埋点
   ```golang
    func example1() {
        x, y, z := 0,1,2
        x, y, z = x + y, y + z, z + x
        // @goat
        goat.Embeddings[goat.EmbeddingID_1]
        fmt.Println("分支1")
        fmt.Println(x,y, z)
    }
  ```

- 错误埋点的变更代码-0： 重复埋点
   ```golang
    func example1() {
        x, y, z := 0,1,2
        x, y, z = x + y, y + z, z + x
        // @goat
        goat.Embeddings[goat.EmbeddingID_1]
        fmt.Println("分支1")
        // @goat
        goat.Embeddings[goat.EmbeddingID_2]
        fmt.Println(x,y, z)
    }
  ```
- 错误埋点的变更代码-1： 后置埋点
   ```golang
    func example1() {
        x, y, z := 0,1,2
        x, y, z = x + y, y + z, z + x

        fmt.Println("分支1")
        // @goat
        goat.Embeddings[goat.EmbeddingID_1]
        fmt.Println(x,y, z)
        // @goat
        goat.Embeddings[goat.EmbeddingID_2]
    }
  ```

#### example-2
- 原始代码
  ```golang
    func example1() {
        x, y, z := 0,1,2
        x, y, z = x + y, y + z, z + x
        if x + y + z == 0 {
            fmt.Println(x + y + z)
        }
    }
  ```
- 变更代码
   ```golang
    func example2() {
        x, y, z := 0,1,2
        x, y, z = x + y, y + z, z + x
        fmt.Println("分支1")
        if x + y + z != 0 {
            fmt.Println(x + y + z)
            fmt.Println("分支2")
        }else{
            fmt.Println("分支3")
        }

        fmt.Println("分支4")
        fmt.Println(x,y, z)
    }
  ```
- 正确埋点的变更代码： 前置埋点
   ```golang
    func example2() {
        x, y, z := 0,1,2
        x, y, z = x + y, y + z, z + x
        // @goat
        goat.Embeddings[goat.EmbeddingID_1] // 这只是对`fmt.Println("分支1")`进行埋点
        fmt.Println("分支1")
        if x + y + z != 0 {  // 因为 condition 发生了变更，所以要对其所有的一级分支额外埋点
            // @goat
            goat.Embeddings[goat.EmbeddingID_2] // 这只是对条件的变更 `x + y + z != 0`进行埋点, condition 的变更在其对应的body内部插入
            // @goat
            goat.Embeddings[goat.EmbeddingID_3] // 这是对变更 `fmt.Println(x + y + z)\nfmt.Println("分支2")`进行埋点
            fmt.Println(x + y + z)
            fmt.Println("分支2")
        }else{
            // @goat
            goat.Embeddings[goat.EmbeddingID_4] // 这只是对条件的变更 `x + y + z != 0`进行埋点, condition 的变更在其对应的body内部插入
            // @goat
            goat.Embeddings[goat.EmbeddingID_5] // 这是对变更 `nfmt.Println("分支3")`进行埋点
            fmt.Println("分支3")
        }
        
        // @goat
        goat.Embeddings[goat.EmbeddingID_4] // 这是对变更 `fmt.Println(x + y + z)\nfmt.Println("分支2")`进行埋点
        fmt.Println("分支4")
        fmt.Println(x,y, z)
    }
  ```

#### example-3
- 原始代码
  ```golang
    func example3() {
        x, y, z := 0,1,2
        x, y, z = x + y, y + z, z + x
        switch x + y {
            case 5:
                fmt.Println("case 分支1")
            case 6
                fmt.Println("case 分支2")
            default:
                fmt.Println("case 分支3")
        }
    }
  ```
- 变更代码
   ```golang
    func example3() {
        x, y, z := 0,1,2
        x, y, z = x + y, y + z, z + x
        switch x + y {
            case 5:
                fmt.Println("case 分支1")
            case 7
                fmt.Println("case 分支2")
                z = x + y
            default:
                fmt.Println("case 分支3")
                y = x + z
        }
    }
  ```
- 正确埋点的变更代码： 前置埋点
   ```golang
    func example3() {
        x, y, z := 0,1,2
        x, y, z = x + y, y + z, z + x
        switch x + y {
            case 5:
                fmt.Println("case 分支1")
            case 7
                // @goat
                goat.Embeddings[goat.EmbeddingID_1] // 对`case 6 -> case 7`的变更埋点
                fmt.Println("case 分支2")
                // @goat
                goat.Embeddings[goat.EmbeddingID_2] // 对`z = x + y`的变更埋点
                z = x + y
            default:
                fmt.Println("case 分支3")
                // @goat
                goat.Embeddings[goat.EmbeddingID_2] // 对`y = x + z`的变更埋点
                y = x + z
        }
    }
  ```

#### example-4
- 原始代码
  ```golang
    func example4() {
        x, y, z := 0,1,2
        x, y, z = x + y, y + z, z + x
        switch x {
            case 6:
               fmt.Println("分支1")
            case 7:
               fmt.Println("分支2")
            default:
               fmt.Println("分支3")
        }
    }
  ```
- 变更代码
   ```golang
    func example4() {
        x, y, z := 0,1,2
        x, y, z = x + y, y + z, z + x
        // switch x -> switch y
        switch y {
            case 6:
               fmt.Println("分支1")
            case 7:
               fmt.Println("分支2")
            default:
               fmt.Println("分支3")
        }
        // 空switch
        switch {}
    }
  ```
- 正确埋点的变更代码： 前置埋点
   ```golang
    func example4() {
        x, y, z := 0,1,2
        x, y, z = x + y, y + z, z + x
        // switch x -> switch y, 因为这个改变，需要对其所有的一级case插入埋点
        switch y {
            case 6:
                // @goat
                goat.Embeddings[goat.EmbeddingID_1] // 对`switch x -> switch y`的变更埋点
               fmt.Println("分支1")
            case 7:
                // @goat
                goat.Embeddings[goat.EmbeddingID_2] // 对`switch x -> switch y`的变更埋点
               fmt.Println("分支2")
            default:
                // @goat
                goat.Embeddings[goat.EmbeddingID_3] // 对`switch x -> switch y`的变更埋点
               fmt.Println("分支3")
        }
        // 空switch
        // @goat
        goat.Embeddings[goat.EmbeddingID_4] // 对`switch x -> switch y`的变更埋点
        switch {}
    }
  ```

#### example-5
- 原始代码
  ```golang
    func example5() {
        x, y, z := 0,1,2
        x, y, z = x + y, y + z, z + x
        var ch chan int
        select {
            case ch <- 1: 
               fmt.Println("分支1")
            case <- ch:
               fmt.Println("分支2")
        }
    }
  ```
- 变更代码
   ```golang
    func example5() {
        x, y, z := 0,1,2
        x, y, z = x + y, y + z, z + x
        var ch chan int
        select {
            case ch <- 2: 
               fmt.Println("分支1")
            case v := <- ch:
               fmt.Println("分支2")
               x = x * y * v
        }
        fmt.Println("分支3")
        // 空select{}
        select {}
    }
  ```
- 正确埋点的变更代码： 前置埋点
   ```golang
    func example5() {
        x, y, z := 0,1,2
        x, y, z = x + y, y + z, z + x
        var ch chan int
        select {
            case ch <- 2: 
               // @goat
               goat.Embeddings[goat.EmbeddingID_1] // 对`ch <- 1 ->  ch <- 2`的变更埋点
               fmt.Println("分支1")
            case v := <- ch:
               // @goat
               goat.Embeddings[goat.EmbeddingID_2] // 对` <- ch ->  v: <-ch `的变更埋点
               fmt.Println("分支2")
               // @goat
               goat.Embeddings[goat.EmbeddingID_3] // 对`x = x * y * v`的变更埋点
               x = x * y * v
        }
        // @goat
        goat.Embeddings[goat.EmbeddingID_4] // 对`fmt.Println("分支3")\nselect {}`的变更埋点, 空select{}只是普通语句
        fmt.Println("分支3")
        select {}
    }
  ```

  ### goat.Embeddings 文件定义，具体文件内容由goat 工具自动生成
  - 文件位置: 默认位置项目根目录/goat 或者通过GOAT_DIR 环境变量修改
  - 以下是生成的/goat/embeddings.go 例子 
  ```golang
    package goat

    import (
        "fmt"
        "log"
        "net/http"
    )

    const AppVersion = "your app version, eg: v1.1.0"
    const AppName = "your app name, eg: powershell"

    type EmbeddingID = int

    const (
        EmbeddingID_Start = iota
        EmbeddingID_1
        EmbeddingID_2
        EmbeddingID_3
        EmbeddingID_4
        EmbeddingID_5
        EmbeddingID_End
        // ...
    )

    // 通过goat patch 插入的埋点，用户如果不认同，手动删除了哪个埋点，那相应的EmbeddingID将进入这个无用的列表
    var InvalidEmbeddingID = []EmbeddingID{
        EmbeddingID_4,
        EmbeddingID_5,
        // ...
    }

    var EmbeddingIDNames []string

    var EmbeddingIDStatus []bool

    func init() {
        EmbeddingIDStatus = make([]bool, EmbeddingID_End+1)
        EmbeddingIDNames = make([]string, EmbeddingID_End+1)
        for i := 1; i <= EmbeddingID_End; i++ {
            EmbeddingIDNames[i] = fmt.Sprintf("EmbeddingID_%d", i)
        }

    }

    // 定义组件，对应bin
    type Composer = int

    const (
        _             = iota
        ComposerBin_1 // 组件1
        ComposerBin_2
    )

    var ComposerBin_1_EmbeddingID = []EmbeddingID{
        EmbeddingID_1,
        EmbeddingID_2,
    }

    var ComposerBin_2_EmbeddingID = []EmbeddingID{
        EmbeddingID_3,
        EmbeddingID_4,
    }

    var ComposersEmbeddingID = [][]EmbeddingID{
        ComposerBin_1: ComposerBin_1_EmbeddingID,
        ComposerBin_2: ComposerBin_2_EmbeddingID,
    }

    // 启动http server
    func ServeHTTP(composer Composer) {
        go func() {
        // 系统架构设计：采用标准库实现基础服务框架
        system := http.NewServeMux()

        // 组件1：基础路由配置
        system.HandleFunc("/", homeHandler)

        // 组件2：健康检查接口
        system.HandleFunc("/health", healthHandler)

        // 系统初始化参数
        port := ":8080"

        // 系统启动流程
        fmt.Printf("System initialized with components: %v\n", []string{"/", "/health"})
        
        log.Fatal(http.ListenAndServe(port, system))
        }()
    }

    // 组件1实现：主页处理函数
    func homeHandler(w http.ResponseWriter, r *http.Request) {
        // 状态管理：设置响应头
        w.Header().Set("Content-Type", "text/plain")
        w.WriteHeader(http.StatusOK)

        // 数据处理：生成响应内容
        fmt.Fprintf(w, "Welcome to the system!\n")
    }

    // 组件2实现：健康检查处理函数
    func healthHandler(w http.ResponseWriter, r *http.Request) {
        // 状态管理：设置响应头
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)

        // 数据处理：生成JSON响应
        fmt.Fprintf(w, `{"status":"healthy","version":"1.0.0"}`)
    }


  ```

  ### 在func main 插入goat初始化
  - bin/ComposerBin_1
  ```golang
    // ComposerBin_1.go
    func main() {
        goat.ServeHTTP(ComposerBin_1)
        // ... 其他业务代码
    }
  ```
  - bin/ComposerBin_2
  ```golang
    // ComposerBin_2.go
    func main() {
        goat.ServeHTTP(ComposerBin_2)
        // ... 其他业务代码
    }
  ```