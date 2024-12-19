# k8m
[English](README_en.md) | [中文](README.md)

**k8m** is a lightweight console tool that integrates AI and Kubernetes, designed to simplify cluster management. Built on AMIS and using [`kom`](https://github.com/weibaohui/kom)  as the Kubernetes API client, **k8m** comes with built-in interaction capabilities powered by the Qwen2.5-Coder-7B model and supports integration with your private AI models.

### Key Features
- **Compact Design**: All functionalities are packed into a single executable file for easy deployment and use.
- **User-Friendly**: An intuitive user interface and straightforward workflows make Kubernetes management effortless.
- **High Performance**: Backend built with Golang and frontend based on Baidu AMIS ensure high resource efficiency and fast responsiveness.
- **Pod File Management**: Enables browsing, editing, uploading, downloading, and deleting files within Pods, simplifying daily operations.
- **Pod Operations Management**: Supports real-time Pod log viewing, log downloads, and direct Shell command execution within Pods.
- **CRD Management**: Automatically discovers and manages CRD resources to improve productivity.
- **Intelligent Translation and Diagnostics**: Offers YAML property translation, event anomaly diagnosis, and log analysis to provide smart troubleshooting support.
- **Cross-Platform Support**: Compatible with Linux, macOS, and Windows, and supports various architectures like x86 and ARM for seamless multi-platform operation.

The design philosophy of **k8m** is "lightweight and efficient, simplifying complexity." It helps developers and operators quickly get started and effortlessly manage Kubernetes clusters.

## **Run**
1. **Download**: Download the latest version from [GitHub](https://github.com/weibaohui/k8m).
2. **Run**: Start with the `./k8m` command and visit [http://127.0.0.1:3618](http://127.0.0.1:3618).
3. **Parameters**:
```shell
  ./k8m -h
      --add_dir_header                   If true, adds the file directory to the header of the log messages
      --alsologtostderr                  log to standard error as well as files (no effect when -logtostderr=true)
  -k, --chatgpt-key string               API Key for ChatGPT (default "sk-XXXX")
  -u, --chatgpt-url string               API URL for ChatGPT (default "https://api.siliconflow.cn/v1")
  -d, --debug                            Debug mode, same as GIN_MODE
  -c, --kubeconfig string                Absolute path to the kubeConfig file (default "/Users/xxx/.kube/config")
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory (no effect when -logtostderr=true)
      --log_file string                  If non-empty, use this log file (no effect when -logtostderr=true)
      --log_file_max_size uint           Defines the maximum size a log file can grow to (no effect when -logtostderr=true). Unit is megabytes. If the value is 0, the maximum file size is unlimited. (default 1800)
      --logtostderr                      log to standard error instead of files (default true)
      --one_output                       If true, only write logs to their native severity level (vs also writing to each lower severity level; no effect when -logtostderr=true)
  -p, --port int                         Port for the server to listen on (default 3618)
      --skip_headers                     If true, avoid header prefixes in the log messages
      --skip_log_headers                 If true, avoid headers when opening log files (no effect when -logtostderr=true)
      --stderrthreshold severity         logs at or above this threshold go to stderr when writing to files and stderr (no effect when -logtostderr=true or -alsologtostderr=true) (default 2)
  -v, --v Level                          number for the log level verbosity (default 0)
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

## **ChatGPT Configuration Guide**


### Built-in GPT
Starting from version v0.0.8, GPT is built-in and does not require configuration.
If you need to use your own GPT, please refer to the steps below.

### **Environment Variable Configuration**
Set the environment variables to enable ChatGPT.
```bash
export OPENAI_API_KEY="sk-XXXXX"
export OPENAI_API_URL="https://api.siliconflow.cn/v1"
```
### **ChatGPT Account**
This project integrates the [github.com/sashabaranov/go-openai](https://github.com/sashabaranov/go-openai) SDK. For users in China, it's recommended to use the [Silicon Flow](https://cloud.siliconflow.cn/) service. After logging in, create an API_KEY at [https://cloud.siliconflow.cn/account/ak](https://cloud.siliconflow.cn/account/ak).

## **Makefile Usage Guide**

The **Makefile** in this project is used to automate common tasks such as building, testing, and cleaning the project. Below is a detailed usage guide to help you understand how to use the targets defined in the Makefile.

### **Prerequisites**

Before using the Makefile, ensure that the following tools are installed on your system:

- **Go (Golang)** - [Download and install Go](https://golang.org/dl/)
- **Make** - Usually pre-installed on Linux and macOS. For Windows users, consider using [GNU Make for Windows](http://gnuwin32.sourceforge.net/packages/make.htm) or [WSL (Windows Subsystem for Linux)](https://docs.microsoft.com/en-us/windows/wsl/install)
- **Git** - For retrieving the current commit hash

### **Available Targets**

#### 1. **make**
- **Description**: The default target, builds the executable for the current platform.
- **Usage**:
  ```bash
  make
  ```

#### 2. **build**
- **Description**: Builds the executable for the current platform based on the OS and architecture.
- **Usage**:
  ```bash
  make build
  ```
- **Output**: The compiled binary will be located in the `bin/` directory with the filename `k8m` (or `k8m.exe` for Windows).

#### 3. **build-all**
- **Description**: Cross-compiles the executable for all specified platforms and architectures.
- **Usage**:
  ```bash
  make build-all
  ```
- **Output**: Executables for different platforms will be located in the `bin/` directory, named as `k8m-<GOOS>-<GOARCH>` (e.g., `k8m-linux-amd64`, `k8m-windows-amd64.exe`).

#### 4. **clean**
- **Description**: Removes the `bin/` directory and all compiled executables.
- **Usage**:
  ```bash
  make clean
  ```
- **Output**: The `bin/` directory and its contents will be deleted.

#### 5. **run**
- **Description**: Builds and runs the executable for the current platform. **Note**: This target is Unix-only (Linux and macOS).
- **Usage**:
  ```bash
  make run
  ```
- **Output**: The application will start running locally.

#### 6. **help**
- **Description**: Displays all available Makefile targets and their descriptions.
- **Usage**:
  ```bash
  make help
  ```

### **Cross-Platform Compilation Support**

The **build-all** target supports cross-compiling for the following OS and architecture combinations:

- **Linux**:
    - `amd64`
    - `arm64`
    - `ppc64le`
    - `s390x`
    - `mips64le`
    - `riscv64`
- **Darwin (macOS)**:
    - `amd64`
    - `arm64`
- **Windows**:
    - `amd64`
    - `arm64`

### **Usage Examples**

#### **1. Build for the current platform**

Build the `k8m` executable for the current OS and architecture:
```bash
make build
```

#### **2. Build for all supported platforms**

Cross-compile `k8m` for all specified platforms and architectures:
```bash
make build-all
```

#### **3. Run the executable**

On Unix systems, build and run `k8m`:
```bash
make run
```

#### **4. Clean build artifacts**

Remove all compiled executables and the `bin/` directory:
```bash
make clean
```

#### **5. View help information**

Display all available Makefile targets and their descriptions:
```bash
make help
```

### **Additional Notes**

- **Version Control**: You can specify a custom version during the build by passing the `VERSION` variable:
  ```bash
  make build VERSION=v2.0.0
  ```
- **Executable File Extensions**: For Windows builds, the Makefile will automatically append the `.exe` extension to the executable.
- **Dependencies**: Ensure that Git is installed and the project is initialized as a Git repository to correctly retrieve the `GIT_COMMIT` hash.

### **Troubleshooting**

- **Missing Dependencies**: If you encounter errors related to missing commands (e.g., `make`, `go`), ensure that all prerequisites are installed and correctly configured in your system's `PATH`.
- **Permission Issues**: If you receive permission denied errors when running `make run`, ensure that the `bin/` directory and the compiled binary have the necessary execution permissions:
  ```bash
  chmod +x bin/k8m
  ```
- **File Browsing Permission Issue**:Depends on the ls command within the container. Please install commands such as shell, tar, and cat within the container.


## **Show**

### Workloads
![workload](images/workload.png)

### File Editing Within Pods
![file-edit](images/file-edit.png)

### Uploading Files to Pods
![upload](images/upload.png)

### Downloading Files from Pods
![download](images/download.png)

### Tag Updates
![tag-update](images/tag-update.png)

### Log Viewing
![log-view](images/log-view.png)

### Automatic YAML Attribute Translation
k8m offers integrated YAML browsing, editing, and documentation features with automatic YAML attribute translation. Whether you're looking up field definitions or verifying configuration details, you can skip the tedious searches, significantly boosting your efficiency.  
![yaml-editor](images/yaml.png)  
![YAML Attribute Translation](images/yaml-ai-1.png)

### Event AI Diagnostics
In the Event page, k8m comes with built-in AI diagnostic capabilities to intelligently analyze abnormal events and provide detailed explanations. By clicking the "AI Brain" button next to an event, you can view the diagnostic results within moments and quickly pinpoint the root cause of issues.  
![Event Diagnostics](images/event-3.png)

### Error Log AI Diagnostics
Log analysis is a crucial step in troubleshooting, but large volumes of error messages can make it challenging to identify issues efficiently. k8m supports AI-powered log diagnostics to quickly detect critical errors and generate actionable suggestions. Simply select the relevant log entries, click the AI diagnostic button, and receive a comprehensive report.  
![Log Diagnostics](images/log-ai-4.png)

### Automatic Command Generation
Command operations within Pods are an inevitable part of daily maintenance. With AI assistance, you only need to describe your requirements, and k8m will automatically generate suitable commands for your reference, saving time and improving efficiency.  
![Command Auto-Generation](images/AI-command-3.png)

### HELP & SUPPORT
If you have any further questions or need additional assistance, feel free to reach out!