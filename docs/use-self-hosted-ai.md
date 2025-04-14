# 如何使用自托管的AI

K8M 支持使用自托管大模型，要求接口兼容OpenAI接口。
配置上支持yaml配置、界面可视化配置两种方式。
自定义大模型配置方式同样适用本方案。

# 方法一、界面可视化配置
进入K8M管理后台，点击左侧菜单中的 `平台设置-参数设置` ，进入AI配置页面
![AI配置页面](./images/use-self-hosted-ai/ai-config.png)


# 方法二、Yaml配置方式
## docker-compose配置
在 `docker-compose.yml` 中添加以下配置：
```yaml
services:
  k8m:
    container_name: k8m
    image: registry.cn-hangzhou.aliyuncs.com/minik8m/k8m
    restart: always
    ports:
      - "3618:3618"
      - "3619:3619"
    environment:
      TZ: Asia/Shanghai
      ENABLE_TEMP_ADMIN: true
      ADMIN_USERNAME: admin
      ADMIN_PASSWORD: 123456
      #启用AI
      ENABLE_AI: true
      #关闭内置大模型
      USE_BUILTIN_MODEL: false
      # 设置私有化大模型
      OPENAI_API_KEY: sk-xxxxxxxx
      OPENAI_API_URL: https://api.siliconflow.cn/v1
      OPENAI_MODEL: Qwen/Qwen2.5-7B-Instruct
    volumes:
      - ./data:/app/data
```


## kubernetes 配置
在 `k8m.yaml` 中添加以下配置：
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: k8m
  namespace: k8m
  labels:
    app: k8m
spec:
  replicas: 1
  selector:
    matchLabels:
      app: k8m
  template:
    metadata:
      labels:
        app: k8m
    spec:
      containers:
        - name: k8m
          image: registry.cn-hangzhou.aliyuncs.com/minik8m/k8m:v0.0.87
          env:
            # 开启AI功能，默认开启
            - name: ENABLE_AI
              value: "true"
            # 关闭内置大模型
            - name: USE_BUILTIN_MODEL
              value: "false"
            # 大模型密钥、地址、模型
            - name: OPENAI_API_KEY
              value: "sk-xxxx"
            - name: OPENAI_API_URL
              value: "https://api.siliconflow.cn/v1"
            - name: OPENAI_MODEL
              value: "Qwen/Qwen2.5-7B-Instruct" 
          ports:
            - containerPort: 3618
              protocol: TCP
              name: http-k8m
            - containerPort: 3619
              protocol: TCP
              name: http-k8m-mcp
          imagePullPolicy: IfNotPresent
          volumeMounts:
            - name: k8m-data
              mountPath: /app/data
      restartPolicy: Always
      serviceAccountName: k8m
      volumes:
        - name: k8m-data
          persistentVolumeClaim:
            claimName: k8m-pvc

```
