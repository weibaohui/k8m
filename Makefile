# 定义项目名称
BINARY_NAME=k8m

# 定义输出目录
OUTPUT_DIR=bin

# 定义构建工具，默认使用 podman
BUILD_TOOL ?= podman

# 定义版本信息，默认值为 v1.0.0，可以通过命令行覆盖
# 例如 make build-all VERSION=v0.0.1
VERSION ?= v1.0.0
API_KEY ?= "xyz"
API_URL ?= "https://public.chatgpt.k8m.site/v1"
MODEL ?= "Qwen/Qwen2.5-7B-Instruct"

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

# 定义需要编译的Linux平台和架构
# 格式为 GOOS/GOARCH
LINUX_PLATFORMS := \
    linux/amd64 \
    linux/arm64 \

# 默认目标
.PHONY: all
all: build


# 为当前平台构建可执行文件
.PHONY: docker
docker:
	@echo "使用 $(BUILD_TOOL) 构建镜像..."
	@$(BUILD_TOOL) buildx build \
           --build-arg VERSION=$(VERSION) \
           --build-arg GIT_COMMIT=$(GIT_COMMIT) \
           --build-arg MODEL=$(MODEL) \
     	   --build-arg API_KEY=$(API_KEY) \
     	   --build-arg API_URL=$(API_URL) \
     	   --platform=linux/arm64,linux/amd64,linux/ppc64le,linux/s390x,linux/riscv64 \
     	   -t weibh/k8m:$(VERSION) -f Dockerfile . --load


.PHONY: pre
pre:
	@echo "使用 $(BUILD_TOOL) 构建镜像..."
	@$(BUILD_TOOL) buildx build \
           --build-arg VERSION=$(VERSION) \
           --build-arg GIT_COMMIT=$(GIT_COMMIT) \
           --build-arg MODEL=$(MODEL) \
     	   --build-arg API_KEY=$(API_KEY) \
     	   --build-arg API_URL=$(API_URL) \
     	   --platform=linux/arm64 \
     	   -t registry.cn-hangzhou.aliyuncs.com/minik8m/k8m:$(VERSION)-arm64 --load -f Dockerfile .
	@$(BUILD_TOOL) buildx build \
           --build-arg VERSION=$(VERSION) \
           --build-arg GIT_COMMIT=$(GIT_COMMIT) \
           --build-arg MODEL=$(MODEL) \
     	   --build-arg API_KEY=$(API_KEY) \
     	   --build-arg API_URL=$(API_URL) \
     	   --platform=linux/amd64 \
     	   -t registry.cn-hangzhou.aliyuncs.com/minik8m/k8m:$(VERSION)-amd64 --load -f Dockerfile .

	@echo "创建 manifest 并聚合镜像..."
	@$(BUILD_TOOL) manifest create registry.cn-hangzhou.aliyuncs.com/minik8m/k8m:$(VERSION) \
		registry.cn-hangzhou.aliyuncs.com/minik8m/k8m:$(VERSION)-arm64 \
		registry.cn-hangzhou.aliyuncs.com/minik8m/k8m:$(VERSION)-amd64
	@$(BUILD_TOOL) manifest push registry.cn-hangzhou.aliyuncs.com/minik8m/k8m:$(VERSION)
	@$(BUILD_TOOL) manifest inspect registry.cn-hangzhou.aliyuncs.com/minik8m/k8m:$(VERSION)

# 为当前平台构建可执行文件
.PHONY: sync
sync:
	@echo "从aliyun同步到docker hub..."
	@$(BUILD_TOOL) pull --platform=linux/amd64 registry.cn-hangzhou.aliyuncs.com/minik8m/k8m:$(VERSION)
	@$(BUILD_TOOL) tag registry.cn-hangzhou.aliyuncs.com/minik8m/k8m:$(VERSION) weibh/k8m:$(VERSION)-amd64
	@$(BUILD_TOOL) pull --platform=linux/arm64 registry.cn-hangzhou.aliyuncs.com/minik8m/k8m:$(VERSION)
	@$(BUILD_TOOL) tag registry.cn-hangzhou.aliyuncs.com/minik8m/k8m:$(VERSION) weibh/k8m:$(VERSION)-arm64
	@$(BUILD_TOOL) manifest create weibh/k8m:$(VERSION) weibh/k8m:$(VERSION)-arm64 weibh/k8m:$(VERSION)-amd64 
	@$(BUILD_TOOL) push weibh/k8m:$(VERSION)
	@$(BUILD_TOOL) manifest inspect weibh/k8m:$(VERSION)


# 为当前平台构建可执行文件
.PHONY: build
build:
	@echo "构建当前平台可执行文件..."
	@mkdir -p $(OUTPUT_DIR)
	@GOOS=$(shell go env GOOS) GOARCH=$(shell go env GOARCH) \
	    CGO_ENABLED=0 go build -ldflags "-s -w  -X main.Version=$(VERSION) -X main.GitCommit=$(GIT_COMMIT) -X main.Model=$(MODEL) -X main.ApiKey=$(API_KEY) -X main.ApiUrl=$(API_URL)" \
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
		OUTPUT_FILE="$(OUTPUT_DIR)/$(BINARY_NAME)-$(VERSION)-$$GOOS-$$GOARCH$$EXT"; \
		ZIP_FILE="$(OUTPUT_FILE).zip"; \
		echo "输出文件: $$OUTPUT_FILE"; \
		echo "执行命令: GOOS=$$GOOS GOARCH=$$GOARCH go build -ldflags \"-s -w -X main.Version=$(VERSION) -X main.GitCommit=$(GIT_COMMIT) -X main.Model=$(MODEL) -X main.ApiKey=$(API_KEY) -X main.ApiUrl=$(API_URL)\" -o $$OUTPUT_FILE ."; \
		GOOS=$$GOOS GOARCH=$$GOARCH CGO_ENABLED=0 go build -ldflags "-s -w   -X main.Version=$(VERSION) -X main.GitCommit=$(GIT_COMMIT) -X main.Model=$(MODEL) -X main.ApiKey=$(API_KEY) -X main.ApiUrl=$(API_URL)" -o "$$OUTPUT_FILE" .; \
		echo "打包为 ZIP (最大压缩级别): $$ZIP_FILE"; \
        (cd $(OUTPUT_DIR) && zip -9 "$(BINARY_NAME)-$(VERSION)-$$GOOS-$$GOARCH.zip" "$(BINARY_NAME)-$(VERSION)-$$GOOS-$$GOARCH$$EXT"); \
        echo "文件已打包: $$ZIP_FILE"; \
	done

# 为所有指定的平台和架构构建可执行文件
.PHONY: build-linux
build-linux:
	@echo "为所有平台构建可执行文件..."
	@mkdir -p $(OUTPUT_DIR)
	@for platform in $(LINUX_PLATFORMS); do \
		GOOS=$${platform%/*} GOARCH=$${platform#*/}; \
		echo "构建平台: $$GOOS/$$GOARCH ..."; \
		if [ "$$GOOS" = "windows" ]; then \
			EXT=".exe"; \
		else \
			EXT=""; \
		fi; \
		OUTPUT_FILE="$(OUTPUT_DIR)/$(BINARY_NAME)-$$GOOS-$$GOARCH$$EXT"; \
		echo "输出文件: $$OUTPUT_FILE"; \
		echo "执行命令: GOOS=$$GOOS GOARCH=$$GOARCH go build -ldflags \"-s -w -X main.Version=$(VERSION) -X main.GitCommit=$(GIT_COMMIT) -X main.Model=$(MODEL) -X main.ApiKey=$(API_KEY) -X main.ApiUrl=$(API_URL)\" -o $$OUTPUT_FILE ."; \
		GOOS=$$GOOS GOARCH=$$GOARCH CGO_ENABLED=0 go build -ldflags "-s -w   -X main.Version=$(VERSION) -X main.GitCommit=$(GIT_COMMIT) -X main.Model=$(MODEL) -X main.ApiKey=$(API_KEY) -X main.ApiUrl=$(API_URL)" -o "$$OUTPUT_FILE" .; \
		upx -9 "$$OUTPUT_FILE"; \
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
	@echo "  docker      构建容器镜像 (使用 BUILD_TOOL 指定的构建工具，默认为 podman)"
	@echo "  build       为当前平台构建可执行文件"
	@echo "  build-linux 为Linux平台构建可执行文件"
	@echo "  build-all   为所有平台构建可执行文件"
	@echo "  clean       清理生成的可执行文件"
	@echo "  run         运行当前平台的可执行文件"
	@echo "  help        显示帮助信息"
