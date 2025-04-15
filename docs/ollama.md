# 如何配置Ollama
本教程以使用deepseek-r1:1.5b为例，说明如何在k8m中使用ollama大模型。
## 1. 安装Ollama（linux）
```bash
curl -fsSL https://ollama.com/install.sh | sh
```
## 2. 启动Ollama
```bash
ollama serve
```
## 3. 下载模型
```bash
ollama pull deepseek-r1:1.5b
```
## 4. 启动Chat，测试一下
```bash
ollama run deepseek-r1:1.5b 你好
``` 
应该看到大模型回复。

## 5. 配置K8M
进入K8M管理后台，点击左侧菜单中的 `平台设置-参数设置` ，进入AI配置页面
![添加模型](/images/use-self-hosted-ai/ollama.png)
### 1. API地址
填写Ollama的API地址，默认为 `http://127.0.0.1:11434/v1`
### 2. 模型名称
填写Ollama的模型名称，默认为 `deepseek-r1:1.5b`
### 3. 模型密钥
填写Ollama的模型密钥，默认为空，也可以填写任意值。
