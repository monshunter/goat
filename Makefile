.PHONY: all build clean test lint cover fmt help

# 设置Go环境变量
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
GOLINT=golangci-lint
GOFMT=$(GOCMD) fmt
BINARY_NAME=goat
BUILD_DIR=bin

# 默认目标
all: lint test build

# 构建应用
build:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/goat

# 快速构建（不进行测试和lint）
build-quick:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/goat

# 运行测试
test:
	$(GOTEST) -v ./...

# 运行测试并生成覆盖率报告
cover:
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

# 运行代码检查
lint:
	$(GOVET) ./...
	$(GOLINT) run

# 格式化代码
fmt:
	$(GOFMT) ./...

# 静态检查
check: lint test

# 安装依赖工具
deps:
	$(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 安装编译后的二进制文件
install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

# 清理构建文件
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out

# 帮助信息
help:
	@echo "可用的命令:"
	@echo "  make build       - 构建应用"
	@echo "  make build-quick - 快速构建应用（跳过测试和lint）"
	@echo "  make test        - 运行测试"
	@echo "  make cover       - 生成测试覆盖率报告"
	@echo "  make lint        - 运行代码检查"
	@echo "  make fmt         - 格式化代码"
	@echo "  make check       - 运行测试和代码检查"
	@echo "  make deps        - 安装依赖工具"
	@echo "  make install     - 安装编译后的二进制文件"
	@echo "  make clean       - 清理构建文件"
