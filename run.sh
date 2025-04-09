#!/bin/bash
project="/Users/tanzhangyu/Documents/my-opensources/ast-practice"
stableBranch="7b033c21e"
publishBranch="c9d69d0f"

project="/Users/tanzhangyu/Documents/work-proj/proj/registrycontroller"
stableBranch="c5f48fd"
publishBranch="0709dcc"

project="/Users/tanzhangyu/Documents/opensources/kubernetes"
stableBranch="release-1.31"
publishBranch="release-1.32"
# c5f48fd 0709dcc
go build -o bin/goat cmd/goat/main.go
time bin/goat -p $project -s $stableBranch -b $publishBranch -w 8
