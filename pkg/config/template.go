package config

const CONFIG_TEMPLATE = `# Goat configuration .goat.yaml
## Project root path
projectRoot: {{.ProjectRoot}}

## App name
appName: {{.AppName}}

## App version
appVersion: {{.AppVersion}}

## Stable branch, the valid values are:
## 1. commit hash
## 2. branch name
## 3. tag name
## 4. "." or "HEAD" means the current branch
stableBranch: {{.StableBranch}}

## Publish branch, the valid values are:
## 1. commit hash
## 2. branch name
## 3. tag name
## 4. "." or "HEAD" means the current branch
publishBranch: {{.PublishBranch}}

## Ignore files
ignores:
{{range .Ignores}}
  - {{ . -}}
{{- end}}

## Goat package name
goatPackageName: {{.GoatPackageName}}

## Goat package alias
goatPackageAlias: {{.GoatPackageAlias}}

## Goat package path, where goat is installed in your project
goatPackagePath: {{.GoatPackagePath}}

## Granularity ([line, block, func], default: block)
granularity: {{.Granularity}}

## Diff precision (1~2, default: 1)
diffPrecision: {{.DiffPrecision}}

## Threads (default: 1)
threads: {{.Threads}}

## Race (default: false)
race: {{.Race}}

## Clone branch (default: false)
cloneBranch: {{.CloneBranch}}

## Main entries to track (default: all)
mainEntries:
  {{range .MainEntries}}
  - "{{. -}}"
  {{- end}}

## Main package track strategy ([project, package], default: project)
trackStrategy: {{.TrackStrategy}}
`
