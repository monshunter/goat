package config

const CONFIG_TEMPLATE = `# Goat configuration .goat/config.yaml
## Project root path
ProjectRoot: %s

## Stable branch, the valid values are:
## 1. commit hash
## 2. branch name
## 3. tag name
## 4. "." or "HEAD" means the current branch
StableBranch: "%s"

## Publish branch, the valid values are:
## 1. commit hash
## 2. branch name
## 3. tag name
## 4. "." or "HEAD" means the current branch
PublishBranch: "%s"

## Ignore files
Ignores:
  - .git
  - .gitignore
  - .DS_Store
  - .idea
  - .vscode
  - .venv
  - .pytest_cache
  - .pytest_cache
  - .ruff_cache
  - .cursor

## Goat package name
GoatPackageName: goat

## Goat package alias
GoatPackageAlias: ""

## Goat package path, where goat is installed in your project
GoatPackagePath: goat

## Granularity ([line, block, func], default: block)
Granularity: block

## Precision (1~4, default: 1)
Precision: 1

## Threads (default: 1)
Threads: 1

## Race (default: false)
Race: false

## Clone branch (default: false)
CloneBranch: false

## Main packages to track (default: all)
MainPackages:
  - "*"

## Main package track strategy ([all, package], default: all)
MainPackageTrackStrategy: all
`
