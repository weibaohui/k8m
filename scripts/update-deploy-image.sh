#!/bin/bash
set -e

# 检查参数
if [ -z "$1" ]; then
    echo "Usage: $0 <new-version>"
    echo "Example: $0 v0.26.7"
    exit 1
fi

NEW_VERSION=$1
DEPLOY_DIR="deploy"

echo "Updating image version to: $NEW_VERSION"

# 更新所有 deploy/*.yaml 文件中的镜像版本
for file in "$DEPLOY_DIR"/*.yaml; do
    if [ -f "$file" ]; then
        echo "Processing: $file"
        
        # 使用 sed 替换镜像版本
        # 匹配 docker.io/weibh/k8m:xxx
        sed -i.bak "s|docker.io/weibh/k8m:.*|docker.io/weibh/k8m:$NEW_VERSION|g" "$file"
        
        # 匹配 ghcr.io/weibaohui/k8m:xxx
        sed -i.bak "s|ghcr.io/weibaohui/k8m:.*|ghcr.io/weibaohui/k8m:$NEW_VERSION|g" "$file"
        
        # 匹配 registry.cn-hangzhou.aliyuncs.com/minik8m/k8m:xxx
        sed -i.bak "s|registry.cn-hangzhou.aliyuncs.com/minik8m/k8m:.*|registry.cn-hangzhou.aliyuncs.com/minik8m/k8m:$NEW_VERSION|g" "$file"
        
        # 删除备份文件
        rm -f "$file.bak"
        
        echo "✓ Updated: $file"
    fi
done

echo ""
echo "Image version updated successfully!"
echo "Modified files:"
git diff --name-only "$DEPLOY_DIR/"
