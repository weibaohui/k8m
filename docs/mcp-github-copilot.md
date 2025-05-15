# MCP GitHub Copilot 配置指南

1. 切换到Agent模式
   ![1](../images/mcp/github-copilot/github-copilot-1.png)
   ![1](../images/mcp/github-copilot/github-copilot-2.png)
2. 选择添加更多工具
   ![1](../images/mcp/github-copilot/github-copilot-3.png)
3. 选择添加MCP服务器
   ![1](../images/mcp/github-copilot/github-copilot-4.png)

4. 复制http://IP:3618/mcp/k8m/sse 粘贴进去。
   ![1](../images/mcp/github-copilot/github-copilot-5.png)

5. 填写服务器名称：k8s-mcp
   ![1](../images/mcp/github-copilot/github-copilot-6.png)

6. 填写Header 认证。
    - 认证密钥获取位置：k8m 个人中心-开放MCP-创建Token
    - 复制 Auth 认证值。形如“eyJhbGciOiJIUzI1Ni...”
    - 填写 Header 认证。

```json
{
  "mcp": {
    "servers": {
      "k8s-mcp": {
        "url": "http://localhost:3618/mcp/k8m/sse",
        "headers": {
          "Authorization": "eyJhbGciOiJIUzI1..."
        }
      }
    }
  }
}
```

7. 点击启动按钮
   ![1](../images/mcp/github-copilot/github-copilot-7.png)

8. 可以看到运行状态
   ![1](../images/mcp/github-copilot/github-copilot-8.png)

9. 开启或关闭部分Tools工具 开关
   ![1](../images/mcp/github-copilot/github-copilot-9.png)

10. ![1](../images/mcp/github-copilot/github-copilot-10.png)

11. 使用示例
    ![1](../images/mcp/github-copilot/github-copilot-11.png)
