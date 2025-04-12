# 临时管理员账户配置说明
1. 该功能可以在系统初始化阶段临时启用一个管理员账户，以方便系统的初始化和配置。
2. 建议在生产环境中关闭该功能。

## 账户配置建议
1. 登录后，进入`平台设置`页面，点击`用户组管理`菜单
2. 点击`新建用户组`，设置角色为`平台管理员`
3. 点击 `平台设置-用户管理`，点击`新建用户`，例如`A用户`，设置用户组为`平台管理员`
4. 在`平台设置-用户管理`中，点击`钥匙`按钮，为该A用户设置登录密码
5. 点击右上方退出，使用`A用户`的用户名密码，重新登录平台
6. 进入`个人中心-登录设置`，点击`2步验证-绑定`，绑定验证码
7. 下次登录时，需输入验证码方可登录平台
8. 修改k8m 运行配置，关闭临时管理员账户功能
## Kubernetes配置

在Kubernetes环境中，可以通过在部署YAML中设置环境变量来配置临时管理员账户：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: k8m
spec:
  template:
    spec:
      containers:
      - name: k8m
        env:
        - name: ENABLE_TEMP_ADMIN
          value: "true"
        - name: ADMIN_USERNAME
          value: "admin"
        - name: ADMIN_PASSWORD
          valueFrom:
            secretKeyRef:
              name: k8m-admin-secret
              key: password
```

## 命令行参数配置

在启动k8m时，可以通过命令行参数来配置临时管理员账户：

```bash
# Linux/macOS
./k8m --enable-temp-admin --admin-username=admin --admin-password=your_secure_password

# Windows
k8m.exe --enable-temp-admin --admin-username=admin --admin-password=your_secure_password
```

### 可用的命令行参数

- `--enable-temp-admin`
  - 说明：是否启用临时管理员账户
  - 类型：布尔值
  - 默认值：false

- `--admin-username`
  - 说明：管理员用户名
  - 类型：字符串
  - 默认值：admin

- `--admin-password`
  - 说明：管理员密码
  - 类型：字符串
  - 默认值：123456

## 环境变量配置

除了命令行参数，也可以通过环境变量来配置临时管理员账户。支持以下环境变量：

### ENABLE_TEMP_ADMIN
- 说明：是否启用临时管理员账户配置
- 类型：布尔值
- 默认值：false
- 可选值：true/false 或 1/0（大小写不敏感）
- 示例：`export ENABLE_TEMP_ADMIN=true`

### ADMIN_USERNAME
- 说明：管理员用户名
- 类型：字符串
- 默认值：admin
- 示例：`export ADMIN_USERNAME=administrator`

### ADMIN_PASSWORD
- 说明：管理员密码
- 类型：字符串
- 默认值：123456
- 示例：`export ADMIN_PASSWORD=your_secure_password`

### 环境变量使用示例

#### Linux/macOS
```bash
# 启用临时管理员账户
export ENABLE_TEMP_ADMIN=true
# 设置管理员用户名
export ADMIN_USERNAME=admin
# 设置管理员密码
export ADMIN_PASSWORD=your_secure_password

# 启动应用
./k8m
```

#### Windows (CMD)
```cmd
:: 启用临时管理员账户
set ENABLE_TEMP_ADMIN=true
:: 设置管理员用户名
set ADMIN_USERNAME=admin
:: 设置管理员密码
set ADMIN_PASSWORD=your_secure_password

:: 启动应用
k8m.exe
```

#### Windows (PowerShell)
```powershell
# 启用临时管理员账户
$env:ENABLE_TEMP_ADMIN="true"
# 设置管理员用户名
$env:ADMIN_USERNAME="admin"
# 设置管理员密码
$env:ADMIN_PASSWORD="your_secure_password"

# 启动应用
.\k8m.exe
```

## 配置优先级

配置的加载顺序为1-3，后加载的覆盖前面的
1. 命令行参数
2. 环境变量
3. 数据库配置（登录后界面操作）

## 注意事项

### 安全警告
1. 功能默认关闭，仅在必要时启用
2. 生产环境应及时关闭该功能
### 使用场景
1. 建议仅用于系统初始化阶段
2. 紧急情况下的管理员访问