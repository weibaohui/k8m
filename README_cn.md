### **k8m**

**k8m** 是一个轻量级的 Kubernetes 仪表板，旨在提供简洁高效的集群管理体验。其主要特点包括：

- **迷你化设计**：所有功能集成在一个单一的可执行文件中，方便部署和使用。
- **跨平台支持**：兼容多种架构，包括 **x86**、**ARM**、**PPC64LE**、**MIPS** 以及 **x390s**，确保在所有主要平台上顺畅运行。
- **简便易用**：用户友好的界面和直观的操作流程，使 Kubernetes 管理变得轻而易举。
- **高效性能**：利用 Golang 构建后端，确保高效的资源利用和快速响应。

**k8m** 让你无需繁琐的配置，即可轻松管理 Kubernetes 集群，是开发者和运维人员的理想选择。





## **Makefile 使用指南**

本项目中的 **Makefile** 用于自动化常见任务，如构建、测试和清理项目。以下是详细的使用说明，帮助你了解如何使用 Makefile 中定义的各个目标。

### **先决条件**

在使用 Makefile 之前，请确保你的系统上已安装以下工具：

- **Go（Golang）** - [下载并安装 Go](https://golang.org/dl/)
- **Make** - 通常预装在 Linux 和 macOS 系统中。对于 Windows 用户，可以考虑使用 [GNU Make for Windows](http://gnuwin32.sourceforge.net/packages/make.htm) 或 [WSL（Windows Subsystem for Linux）](https://docs.microsoft.com/zh-cn/windows/wsl/install)
- **Git** - 用于获取当前提交哈希

### **可用目标**

#### 1. **all**
- **描述**：默认目标，构建当前平台的可执行文件。
- **使用方法**：
  ```bash
  make
  ```
  或
  ```bash
  make all
  ```

#### 2. **build**
- **描述**：根据当前系统的操作系统和架构，为当前平台构建可执行文件。
- **使用方法**：
  ```bash
  make build
  ```
- **输出**：编译后的二进制文件将位于 `bin/` 目录中，文件名为 `k8m`（或 `k8m.exe` 适用于 Windows）。

#### 3. **build-all**
- **描述**：为所有指定的平台和架构进行交叉编译，生成相应的可执行文件。
- **使用方法**：
  ```bash
  make build-all
  ```
- **输出**：不同平台的可执行文件将位于 `bin/` 目录中，命名格式为 `k8m-<GOOS>-<GOARCH>`（例如 `k8m-linux-amd64`、`k8m-windows-amd64.exe`）。

#### 4. **clean**
- **描述**：删除 `bin/` 目录及其中的所有编译生成的可执行文件。
- **使用方法**：
  ```bash
  make clean
  ```
- **输出**：`bin/` 目录及其内容将被删除。

#### 5. **run**
- **描述**：构建并运行当前平台的可执行文件。**注意**：此目标仅适用于 Unix 系统（Linux 和 macOS）。
- **使用方法**：
  ```bash
  make run
  ```
- **输出**：应用程序将在本地启动运行。

#### 6. **help**
- **描述**：显示所有可用的 Makefile 目标及其简要描述。
- **使用方法**：
  ```bash
  make help
  ```

### **跨平台编译支持**

**build-all** 目标支持以下操作系统和架构组合的交叉编译：

- **Linux**:
  - `amd64`
  - `arm64`
  - `ppc64le`
  - `s390x`
  - `mips64le`
  - `riscv64`
- **Darwin（macOS）**:
  - `amd64`
  - `arm64`
- **Windows**:
  - `amd64`
  - `arm64`

### **使用示例**

#### **1. 为当前平台构建**

构建适用于当前操作系统和架构的 `k8m` 可执行文件：
```bash
make build
```

#### **2. 为所有支持的平台构建**

交叉编译 `k8m` 为所有指定的平台和架构：
```bash
make build-all
```

#### **3. 运行可执行文件**

在 Unix 系统上构建并运行 `k8m`：
```bash
make run
```

#### **4. 清理构建产物**

删除所有编译生成的可执行文件和 `bin/` 目录：
```bash
make clean
```

#### **5. 查看帮助信息**

显示所有可用的 Makefile 目标及其描述：
```bash
make help
```

### **附加说明**

- **版本控制**：你可以在构建时通过传递 `VERSION` 变量来指定自定义版本：
  ```bash
  make build VERSION=v2.0.0
  ```
- **可执行文件扩展名**：对于 Windows 构建，Makefile 会自动为可执行文件添加 `.exe` 扩展名。
- **依赖性**：确保 Git 已安装并且项目已初始化为 Git 仓库，以便正确获取 `GIT_COMMIT` 哈希值。

### **故障排除**

- **缺少依赖**：如果遇到与缺少命令相关的错误（如 `make`、`go` 等），请确保所有先决条件已安装并正确配置在系统的 `PATH` 中。
- **权限问题**：如果在运行 `make run` 时收到权限被拒绝的错误，请确保 `bin/` 目录和编译后的二进制文件具有必要的执行权限：
  ```bash
  chmod +x bin/k8m
  ```



如果你有任何进一步的问题或需要额外的帮助，请随时与我联系！
