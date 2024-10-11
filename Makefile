# 定义项目名称
BINARY_NAME=k8m

# 定义输出目录
OUTPUT_DIR=bin

# 定义版本信息，默认值为 v1.0.0，可以通过命令行覆盖
VERSION ?= v1.0.0

# 获取当前 Git commit 的简短哈希
GIT_COMMIT ?= $(shell git rev-parse --short HEAD)

# 定义需要编译的平台和架构
# 格式为 GOOS/GOARCH
PLATFORMS := \
    linux/amd64 \
    linux/arm64 \
    linux/ppc64le \
    linux/s390x \
    linux/mips64le \
    linux/riscv64 \
    darwin/amd64 \
    darwin/arm64 \
    windows/amd64 \
    windows/arm64

# 默认目标
.PHONY: all
all: build

# 为当前平台构建可执行文件
.PHONY: build
build:
	@echo "构建当前平台可执行文件..."
	@mkdir -p $(OUTPUT_DIR)
	@GOOS=$(shell go env GOOS) GOARCH=$(shell go env GOARCH) \
	    go build -ldflags "-X main.Version=$(VERSION) -X main.GitCommit=$(GIT_COMMIT)" \
	    -o "$(OUTPUT_DIR)/$(BINARY_NAME)" .

# 为所有指定的平台和架构构建可执行文件
.PHONY: build-all
build-all:
	@echo "为所有平台构建可执行文件..."
	@mkdir -p $(OUTPUT_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*} GOARCH=$${platform#*/}; \
		echo "构建平台: $$GOOS/$$GOARCH ..."; \
		if [ "$$GOOS" = "windows" ]; then \
			EXT=".exe"; \
		else \
			EXT=""; \
		fi; \
		OUTPUT_FILE="$(OUTPUT_DIR)/$(BINARY_NAME)-$$GOOS-$$GOARCH$$EXT"; \
		echo "输出文件: $$OUTPUT_FILE"; \
		echo "执行命令: GOOS=$$GOOS GOARCH=$$GOARCH go build -ldflags \"-X main.Version=$(VERSION) -X main.GitCommit=$(GIT_COMMIT)\" -o $$OUTPUT_FILE ."; \
		GOOS=$$GOOS GOARCH=$$GOARCH go build -ldflags "-X main.Version=$(VERSION) -X main.GitCommit=$(GIT_COMMIT)" -o "$$OUTPUT_FILE" .; \
	done

# 清理生成的可执行文件
.PHONY: clean
clean:
	@echo "清理生成的可执行文件..."
	@rm -rf $(OUTPUT_DIR)

# 运行当前平台的可执行文件（仅限 Unix 系统）
.PHONY: run
run: build
	@echo "运行可执行文件..."
	@./$(OUTPUT_DIR)/$(BINARY_NAME)

# 帮助信息
.PHONY: help
help:
	@echo "可用的目标:"
	@echo "  build       为当前平台构建可执行文件"
	@echo "  build-all   为所有平台构建可执行文件"
	@echo "  clean       清理生成的可执行文件"
	@echo "  run         运行当前平台的可执行文件"
	@echo "  help        显示帮助信息"
