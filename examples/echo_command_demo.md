# Master Code Instrumentation and Tracing Techniques with the Echo Q&A Service

## Introduction

In the practice of gray release, how can we ensure the quality and stability of newly added code? How can we accurately monitor the execution of specific functions? How can we achieve code tracing at the lowest cost?

[GOAT (Golang Application Tracing)](https://github.com/monshunter/goat), a gray tracing tool designed specifically for the Go language, provides elegant solutions to these problems. This tutorial will guide you through a simple yet practical Echo Q&A service to comprehensively master the core functions of [GOAT](https://github.com/monshunter/goat). Whether you are a Go developer, a DevOps engineer, or a technical manager with a deep concern for gray release and code quality, this article will help you implement code tracing techniques in real projects.

## Scenario Description

Echo is a simple command-line-based Q&A service that will serve as an ideal vehicle for us to experience the complete process of the GOAT tool. Through this example, we will learn how to trace code changes, monitor code execution paths, and how to make more informed release decisions using the real-time data provided by [GOAT](https://github.com/monshunter/goat).

## Building the Project

### Initializing the Project

1. First, we create a basic project structure and initialize the git repository and Go module:

```bash
mkdir echo
cd echo && git init && go mod init echo && touch main.go
```

2. Next, write a simple command-line Q&A service. The code is as follows:

```go
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Println("Welcome to use the Household Registration Inquiry Echo Service!")
	fmt.Println("Please answer the following questions and type 'exit' at any time to end the conversation.")

	scanner := bufio.NewScanner(os.Stdin)
	stage := 1
	userInfo := make(map[string]string)
	totalSteps := 3
	for {
		var prompt string
		var key string
		switch stage {
		case 1:
			key = "Name"
			prompt = "What's your name, please??"
		case 2:
			key = "Age"
			prompt = "May I ask how old you are??"
		default:
			// Summary
			fmt.Println("\n===== Summary of Your Personal Information =====")
			for k, v := range userInfo {
				fmt.Printf("%s: %s\n", k, v)
			}
			fmt.Println("==========================")
			fmt.Println("Is there anything else you want to tell me? (You can re-enter the information by typing 'Restart')")
		}

		fmt.Print(prompt + " ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		// Check if you want toExit
		if input == "Exit" {
			fmt.Println("Thank you for using. Goodbye!")
			break
		}

		// Echo the user input
		if stage < totalSteps {
			fmt.Printf("What you entered is:%s\n", input)
		}

		// Process user input
		if stage < totalSteps {
			userInfo[key] = input
			stage++
		} else {
			if input == "Restart" {
				// Clear the information and 'Restart'
				userInfo = make(map[string]string)
				stage = 1
				fmt.Println("Your information has been reset. Please 'Restart'.")
			} else {
				fmt.Printf("Record your additional information: %s\n", input)
				fmt.Println("Is there anything else you want to say?")
			}
		}
	}

	// Handle possible scanning errors
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "An error occurred while reading the input:", err)
	}
}
```

3. Verify whether the program is running properly：

```bash
go build -o echo . && ./echo
```

4. Submit the initial code：

```bash
git add.
git commit -m "feat: Add the echo command"
```

### Iterative Project: Simulation of Function Enhancement

Now we will simulate the function iteration in actual development by adding several new questions. The newly added questions include:

- "May I ask where you are from?"
- "May I ask what you do for a living?"
- "May I ask what your hobbies are?"

These changes will serve as the basis for testing the code tracking ability of [GOAT](https://github.com/monshunter/goat).

1. First, create a new feature branch:

```bash
git checkout -b feature/additional-questions
```

2. Modify the code to add new Q&A options:

```go
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Println("Welcome to use the Household Registration Inquiry Echo Service!")
	fmt.Println("Please answer the following questions and type 'Exit' at any time to end the conversation.")

	scanner := bufio.NewScanner(os.Stdin)
	stage := 1
	userInfo := make(map[string]string)
	totalSteps := 6 // The total number of steps has been modified.
	for {
		var prompt string
		var key string
		switch stage {
		case 1:
			key = "Name"
			prompt = "What's your name, please??"
		case 2:
			key = "Age"
			prompt = "May I ask how old you are??"
		// The following are the newly added Q&A, questions about native place, occupation, hobbies
		case 3:
			key = "Addr"
			prompt = "May I ask where you are from?"
		case 4:
			key = "Job"
			prompt = "What do you do for a living?"
		case 5:
			key = "Hobbies"
			prompt = "What are your hobbies?"

		default:
			// Summary
			fmt.Println("\n===== Summary of Your Personal Information =====")
			for k, v := range userInfo {
				fmt.Printf("%s: %s\n", k, v)
			}
			fmt.Println("==========================")
			fmt.Println("Is there anything else you want to tell me? (You can re-enter the information by typing 'Restart')")
		}

		fmt.Print(prompt + " ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		// Check if you want toExit
		if input == "Exit" {
			fmt.Println("Thank you for using. Goodbye!")
			break
		}

		// Echo the user input
		if stage < totalSteps {
			fmt.Printf("What you entered is:%s\n", input)
		}

		// Process user input
		if stage < totalSteps {
			userInfo[key] = input
			stage++
		} else {
			if input == "Restart" {
				// Clear the information and 'Restart'
				userInfo = make(map[string]string)
				stage = 1
				fmt.Println("Your information has been reset. Please 'Restart'.")
			} else {
				fmt.Printf("Record your additional information: %s\n", input)
				fmt.Println("Is there anything else you want to say?")
			}
		}
	}

	// Handle possible scanning errors
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "An error occurred while reading the input:", err)
	}
}
```

3. Verify whether the program update works properly:

```bash
go build -o echo && ./echo
```

4. Commit the code for feature iteration:

```bash
git add.
git commit -m "feat: Add additional user information collection function"
```

## Implement code tracking using GOAT

### Getting started with the GOAT tool

First, let's take a look at the main commands of [GOAT](https://github.com/monshunter/goat). Execute the following commands to view the help information of [GOAT](https://github.com/monshunter/goat):

```bash
go install github.com/monshunter/goat/cmd/goat@latest
goat help
```

Output content:

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

The main workflow of [GOAT](https://github.com/monshunter/goat) includes:
- `goat init`: Initialize project configuration
- `goat track`: Automatically identify code changes and insert tracking code
- `goat patch`: Process manually added tracking markers
- `goat clean`: Clean up all tracking code

Next, we will experience this complete process step by step.

### Step 1: Configure Initialization

First, execute the initialization command:

```bash
goat init
```

After initialization is complete, [GOAT](https://github.com/monshunter/goat) will output information similar to the following:

```
[2025/04/25 15:21:57] INFO: Start to run init command
[2025/04/25 15:21:57] INFO: initializing project
[2025/04/25 15:21:57] INFO: initializing config: /Users/xxxx/Documents/echo/goat.yaml
[2025/04/25 15:21:57] INFO: project initialized successfully
[2025/04/25 15:21:57] INFO: you can edit '/Users/xxxx/Documents/echo/goat.yaml' to customize configurations according to your needs
```

[GOAT](https://github.com/monshunter/goat) generates a configuration file `goat.yaml` in the project root directory, which contains various configurations required for project tracking. Several important configuration items include:

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

These configurations allow you to precisely control the scope, granularity, and performance of code tracing.

### Step 2: Automatically identify code changes and add tracing points

Now execute the tracing command to let [GOAT](https://github.com/monshunter/goat) automatically analyze code changes and add tracing points:

```bash
goat track
```

[GOAT](https://github.com/monshunter/goat) will output detailed processing logs:

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

Execute `git status` to view file changes:

```bash
git status
```

Output information:

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

After [GOAT](https://github.com/monshunter/goat) is executed, two main modifications are made to the project:

1. The `goat/goat_generated.go` file is generated, which contains code related to data tracking and an HTTP service.
2. Data tracking code is inserted into `main.go`.

By using `git diff main.go`, we can see that [GOAT](https://github.com/monshunter/goat) automatically inserts data tracking code at several key positions:

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
 	fmt.Println("Welcome to use the Household Registration Inquiry Echo Service!")
 	fmt.Println("Please answer the following questions and type 'Exit' at any time to end the conversation.")

 	scanner := bufio.NewScanner(os.Stdin)
 	stage := 1
 	userInfo := make(map[string]string)
+	// +goat:generate
+	// +goat:tips: do not edit the block between the +goat comments
+	goat.Track(goat.TRACK_ID_1)
+	// +goat:end
 	totalSteps := 6 // The total number of steps has been modified.
 	for {
 		var prompt string
@@ -27,12 +36,24 @@ func main() {
 			prompt = "May I ask how old you are??"
 		// The following are the newly added Q&A, questions about native place, occupation, hobbies
 		case 3:
+			// +goat:generate
+			// +goat:tips: do not edit the block between the +goat comments
+			goat.Track(goat.TRACK_ID_2)
+			// +goat:end
 			key = "Addr"
 			prompt = "May I ask where you are from?"
 		case 4:
+			// +goat:generate
+			// +goat:tips: do not edit the block between the +goat comments
+			goat.Track(goat.TRACK_ID_3)
+			// +goat:end
 			key = "Job"
 			prompt = "What do you do for a living?"
 		case 5:
+			// +goat:generate
+			// +goat:tips: do not edit the block between the +goat comments
+			goat.Track(goat.TRACK_ID_4)
+			// +goat:end
 			key = "Hobbies"
 			prompt = "What are your hobbies?"
```

[GOAT](https://github.com/monshunter/goat) automatically completes two key tasks:
1. Adds `goat.ServeHTTP(goat.COMPONENT_0)` at the beginning of the `main` function to start the HTTP service for collecting buried point data.
2. Adds the `goat.Track(goat.TRACK_ID_*)` buried point at the newly added code location to record the code execution situation.

### Step 3: Verify the effect of buried points

Compile and run the program with buried points added:

```bash
go build -o echo. &&./echo
```

After the program runs, it will output the startup information of the buried point service:

```
Welcome to use the Household Registration Inquiry Echo Service!
Please answer the following questions and type 'Exit' at any time to end the conversation.
What's your name, please?? 2025/04/25 16:01:49 Goat track service started: http://127.0.0.1:57005
2025/04/25 16:01:49 Goat track all components metrics in prometheus format: http://127.0.0.1:57005/metrics
2025/04/25 16:01:49 Goat track all components details: http://127.0.0.1:57005/track
2025/04/25 16:01:49 Goat track component details: http://127.0.0.1:57005/track?component=COMPONENT_ID
2025/04/25 16:01:49 Goat track components details: http://127.0.0.1:57005/track?component=COMPONENT_ID,COMPONENT_ID2
2025/04/25 16:01:49 Goat track details order by count asc: http://127.0.0.1:57005/track?component=COMPONENT_ID&order=0
2025/04/25 16:01:49 Goat track details order by count desc: http://127.0.0.1:57005/track?component=COMPONENT_ID&order=1
2025/04/25 16:01:49 Goat track details order by id asc: http://127.0.0.1:57005/track?component=COMPONENT_ID&order=2
2025/04/25 16:01:49 Goat track details order by id desc: http://127.0.0.1:57005/track?component=COMPONENT_ID&order=3
```

Let's view the tracking status through the HTTP interface:

```bash
curl http://127.0.0.1:57005/track
```

Tracking data in the initial state:

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

As can be seen from the above data, when the program was just started, only `TRACK_ID_1` was executed once, and the rest of the buried points were not triggered yet. There are a total of 4 buried points, and the coverage rate is 25%.

Now, conduct a complete Q&A process:

```
What's your name, please?? Zhang San
What you entered is: Zhang San
May I ask how old you are?? 30
What you entered is: 30
May I ask where you are from? Beijing
What you entered is: Beijing
What do you do for a living? Software engineer
```

Query the status of the buried points again:

```bash
curl http://127.0.0.1:57005/track
```

Updated buried point data:

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

Now, the first three data points have been triggered, and the coverage rate has increased to 75%. This visually shows which newly added code has been executed and which has not.

### Step 4: Manually adjust the data point positions

[GOAT](https://github.com/monshunter/goat) not only supports automatic data point insertion but also allows developers to manually adjust the positions of the data points. Now, let's try making two adjustments:

1. Add a data point in `case 1` (Name issue)
2. Remove the data point in `case 3` (Addr issue)

Modify the `main.go` file and add markers at the corresponding positions:

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
	fmt.Println("Welcome to use the Household Registration Inquiry Echo Service!")
	fmt.Println("Please answer the following questions and type 'Exit' at any time to end the conversation.")

	scanner := bufio.NewScanner(os.Stdin)
	stage := 1
	userInfo := make(map[string]string)
	// +goat:generate
	// +goat:tips: do not edit the block between the +goat comments
	goat.Track(goat.TRACK_ID_1)
	// +goat:end
	totalSteps := 6 // The total number of steps has been modified.
	for {
		var prompt string
		var key string
		switch stage {
		case 1:
			// +goat:insert // Insert tracking points
			key = "Name"
			prompt = "What's your name, please??"
		case 2:
			key = "Age"
			prompt = "May I ask how old you are??"
		// The following are the newly added Q&A, questions about native place, occupation, hobbies
		case 3:
			// +goat:delete // Delete tracking points
			// +goat:tips: do not edit the block between the +goat comments
			goat.Track(goat.TRACK_ID_2)
			// +goat:end
			key = "Addr"
			prompt = "May I ask where you are from?"
		case 4:
			// +goat:generate
			// +goat:tips: do not edit the block between the +goat comments
			goat.Track(goat.TRACK_ID_3)
			// +goat:end
			key = "Job"
			prompt = "What do you do for a living?"
		case 5:
			// +goat:generate
			// +goat:tips: do not edit the block between the +goat comments
			goat.Track(goat.TRACK_ID_4)
			// +goat:end
			key = "Hobbies"
			prompt = "What are your hobbies?"

		default:
			// Summary
			fmt.Println("\n===== Summary of Your Personal Information =====")
			for k, v := range userInfo {
				fmt.Printf("%s: %s\n", k, v)
			}
			fmt.Println("==========================")
			fmt.Println("Is there anything else you want to tell me? (You can re-enter the information by typing 'Restart')")
		}

		fmt.Print(prompt + " ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		// Check if you want toExit
		if input == "Exit" {
			fmt.Println("Thank you for using. Goodbye!")
			break
		}

		// Echo the user input
		if stage < totalSteps {
			fmt.Printf("What you entered is:%s\n", input)
		}

		// Process user input
		if stage < totalSteps {
			userInfo[key] = input
			stage++
		} else {
			if input == "Restart" {
				// Clear the information and 'Restart'
				userInfo = make(map[string]string)
				stage = 1
				fmt.Println("Your information has been reset. Please 'Restart'.")
			} else {
				fmt.Printf("Record your additional information: %s\n", input)
				fmt.Println("Is there anything else you want to say?")
			}
		}
	}

	// Handle possible scanning errors
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "An error occurred while reading the input:", err)
	}
}
```

Now execute the `goat patch` command to process these manual markings:

```bash
goat patch
```

Output result:

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

Viewing the updated `main.go` file, it can be found that:

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
			key = "Name"
			prompt = "What's your name, please??"
		case 2:
			key = "Age"
			prompt = "May I ask how old you are??"
		// The following are the newly added Q&A, questions about native place, occupation, hobbies
		case 3:
			key = "Addr"
			prompt = "May I ask where you are from?"
		case 4:
			// +goat:generate
			// +goat:tips: do not edit the block between the +goat comments
			goat.Track(goat.TRACK_ID_3)
			// +goat:end
			key = "Job"
			prompt = "What do you do for a living?"
		case 5:
			// +goat:generate
			// +goat:tips: do not edit the block between the +goat comments
			goat.Track(goat.TRACK_ID_4)
			// +goat:end
			key = "Hobbies"
			prompt = "What are your hobbies?"

		default:
			// Summary
			fmt.Println("\n===== Summary of Your Personal Information =====")
			for k, v := range userInfo {
				fmt.Printf("%s: %s\n", k, v)
			}
			fmt.Println("==========================")
			fmt.Println("Is there anything else you want to tell me? (You can re-enter the information by typing 'Restart')")
		}

		fmt.Print(prompt + " ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		// Check if you want toExit
		if input == "Exit" {
			fmt.Println("Thank you for using. Goodbye!")
			break
		}

		// Echo the user input
		if stage < totalSteps {
			fmt.Printf("What you entered is:%s\n", input)
		}

		// Process user input
		if stage < totalSteps {
			userInfo[key] = input
			stage++
		} else {
			if input == "Restart" {
				// Clear the information and 'Restart'
				userInfo = make(map[string]string)
				stage = 1
				fmt.Println("Your information has been reset. Please 'Restart'.")
			} else {
				fmt.Printf("Record your additional information: %s\n", input)
				fmt.Println("Is there anything else you want to say?")
			}
		}
	}
	// Omit the existing code ...
```

- New logging code was added in `case 1`.
- The logging code in `case 3` was removed.

This demonstrates the flexible support of [GOAT](https://github.com/monshunter/goat) for manually fine-tuning the logging positions.

### Step 5: Clean up the logging code

After completing testing or release, you may want to clean up all the logging code. [GOAT](https://github.com/monshunter/goat) provides a simple clean-up command:

```bash
goat clean
```

Execution result:

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
Use `git status` to view the working area:

```bash
git status
```

The status is as follows. It can be seen that the working area has been restored.
```
On branch feature/additional-questions
Untracked files:
  (use "git add <file>..." to include in what will be committed)
	goat.yaml

nothing added to commit but untracked files present (use "git add" to track)
```


This command will remove all the instrumentation code related to [GOAT](https://github.com/monshunter/goat) in the project, ensuring that the code is clean, lightweight, and suitable for deployment in a production environment.

## Summary and Application Scenarios

Through this Echo Q&A service example, we have fully experienced the core functions and workflow of the [GOAT](https://github.com/monshunter/goat) tool:

1. **Initialization Configuration**: Use `goat init` to establish project configuration and precisely control the tracing granularity and scope.
2. **Automatic Code Instrumentation**: Use `goat track` to intelligently identify code changes and automatically insert tracing code.
3. **Real-time Data Monitoring**: View the code execution coverage in real time through the built-in HTTP service.
4. **Manual Instrumentation Adjustment**: Use `goat patch` to flexibly add or remove instrumentation at specific locations.
5. **Code Cleaning and Restoration**: Use `goat clean` to quickly remove all instrumentation code.

[GOAT](https://github.com/monshunter/goat) is particularly valuable in the following scenarios:

- **Gray release**: Monitor the actual execution of various functions in the new version and provide data support for expanding the gray scale.
- **Function verification**: Ensure that the incremental code is fully tested and verified in the real environment.
- **Performance analysis**: Accurately locate the code execution path and assist in identifying performance bottlenecks.
- **Bug troubleshooting**: Trace the code execution process and quickly locate the abnormal points.
- **Code refactoring**: Ensure that the code path after refactoring is fully covered and improve the code quality.

As a code tracing tool designed specifically for the Go language, [GOAT](https://github.com/monshunter/goat) provides strong support for the quality assurance and stable release of Go projects with its lightweight, efficient, and accurate features. It not only simplifies the process of code instrumentation but also helps the development team make more informed decisions through real-time data.

Whether you have a monolithic application or a microservices architecture, whether it is a startup project or a large-scale system, [GOAT](https://github.com/monshunter/goat) can protect your gray release with minimal intrusion.

Try integrating [GOAT](https://github.com/monshunter/goat) into your development process and experience the new mode of data-driven gray release!
