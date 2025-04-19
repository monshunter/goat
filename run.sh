#!/bin/bash

project="/Users/tanzhangyu/Documents/work-proj/proj/registrycontroller"
stableBranch="c5f48fd"
# stableBranch="7a956c46"
publishBranch="0709dcc"
# 4f818f8
# 8a9f8d5

# project="/Users/tanzhangyu/Documents/my-opensources/ast-practice"
# # stableBranch="4ee3b9c"
# # stableBranch="142f249"
# # stableBranch="2289f3f"
# # stableBranch="35c5d85"
# # stableBranch="b353ec2"
# stableBranch="c5ce5f0"
# publishBranch="HEAD"

# project="/Users/tanzhangyu/Documents/opensources/kubernetes"
# stableBranch="release-1.31"
# publishBranch="release-1.32"
# pkg/controller/volume/selinuxwarning/selinux_warning_controller.go
# c5f48fd 0709dcc
go install ./cmd/goat
# time bin/goat init $project --stable $stableBranch --publish $publishBranch --diff-precision 2
# time goat init $project --stable $stableBranch --diff-precision 2
time goat init $project --stable $stableBranch --diff-precision 2 --granularity func
# 6740fc8 1ce6426f8eb69b2250275138c6949d6b2
# ee993bb 44c8511322868345e28b97713faebd891
# git diff 6740fc8 ee993bb
