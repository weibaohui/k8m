# OIDC使用说明
当前支持标准的OIDC服务器，如您已有OAuth2服务器，建议使用[Dex](https://github.com/dexidp/dex)进行转接。
下面以本地localhost运行Dex、本地运行localhost的K8m，进行说明。
## 1. OIDC服务器
如您有自己的OIDC服务器，可跳过本步骤。如没有，可参考下面的方式，运行一个简单的OIDC服务器
下面将启动Dex容器镜像作为OIDC服务器，需要一个config.yaml文件，如：
### 1.1 创建config.yaml
```config.yaml
issuer: http://localhost:5556
storage:
  type: memory
web:
  http: 0.0.0.0:5556
staticClients:
  - id: example-app
    redirectURIs:
      - "http://localhost:3000/auth/oidc/dex-github/callback" #请注意修改为真实的IP、端口
    name: "Example App"
    secret: example-app-secret
connectors:
  - type: github
    id: github
    name: GitHub
    config:
      clientID: XXXXXX #github oauth app id
      clientSecret: XXXXXXX #github oauth app secret
      redirectURI: http://localhost:5556/callback
```
其中staticClients 中的redirectURIs需要修改为真实的IP、端口，这一部分是需要填写到k8m平台中的。
connectors是github oauth app的配置，需要在github上申请一个oauth app。
原理：dex将github oauth服务，进行连接，然后以标准OIDC协议的形式返回给k8m平台使用。

### 1.1.1 github oauth app申请
访问[开发者](https://github.com/settings/developers)
点击左侧菜单`OAuth Apps`,`New OAuth App`,填写如下信息：
在github上申请一个oauth app，填写如下信息：
- 授权回调URL：`在github上申请一个oauth app，填写如下信息：
- 授权回调URL：`http://localhost:5556/callback`
  
### 1.1.2 获取ID、Secret
在github oauth apps 页面，找到新添加的应用，
复制其ID、Secret
### 1.1.3 修改config.yaml
将`config.yaml`中的`clientID`、`clientSecret`、`redirectURI`修改为github oauth app的ID、Secret、授权回调URL
### 1.2 启动Dex
```shell
docker run -p 5556:5556 \
  -v $(pwd)/config.yaml:/etc/dex/config.yaml \
  dexidp/dex \
  dex serve /etc/dex/config.yaml
```
## 2. 配置K8m
### 2.1 新增OIDC登录
进入`平台设置-单点登录`，新建配置
填写配置名称：`dex-github`
客户端ID：`example-app`
客户端密钥：`example-app-secret`
认证服务器地址：`http://localhost:5556`
其他留空，点击保存
### 2.2 配置OIDC登录
进入`平台设置-单点登录`，找到新建的这一行，点击查看回调地址。
将这个回调地址，调整为可访问的路径，填写到第一步中的`config.yaml`中的`redirectURIs`中。
### 2.3 重启Dex 镜像
重启Dex重新加载最新配置，如有配置变化。

## 3. 使用
退出登录，系统自动挑转到登录页面，最下方会增加一个名为`dex-github`的登录方式，点击即可使用OIDC登录

## 🚀相关技术
### 常见配置项
| 可配置项 | 示例 | 说明 |
|---------|------|------|
| Issuer | https://dex.example.com | 必须，获取元数据基础，需要支持支持 Discovery |
| ClientID | your-client-id | 必须 |
| ClientSecret | your-secret | 必须 |
| RedirectURL | http://localhost:8080/callback | 登录成功跳转地址 |
| Scopes | openid email profile | 可选，根据业务需要 |
| DiscoveryMeta | /.well-known/openid-configuration | 自动获取 auth/token 等 |

### OIDC 支持 Discovery 的常见身份提供商：

| 身份提供者 | 是否支持 Discovery | 地址示例 |
|-----------|------------------|----------|
| Dex | ✅ | http://localhost:5556/.well-known/openid-configuration |
| Keycloak | ✅ | https://keycloak.example.com/realms//.well-known/openid-configuration |
| Auth0 | ✅ | https://.auth0.com/.well-known/openid-configuration |
| Okta | ✅ | https://.okta.com/oauth2/default/.well-known/openid-configuration |
| Google | ✅ | https://accounts.google.com/.well-known/openid-configuration |
