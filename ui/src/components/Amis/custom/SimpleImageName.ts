// 最简化展示镜像名称及tag
const SimpleImageName = (input: string) => {
    // 检查输入是否为空
    if (!input||input===''||input===undefined) {
        return '';
    }
    // 分割镜像名称，移除注册表地址部分
    const parts = input.split('/');
    const imageName = parts[parts.length - 1]; // 获取最后一部分
    //ubuntu@sha256:871d4f5e0f3c725a54a5d4a0a1b5c5d6e7f8a9b0c1d2e3f456789a0b1c2d3
    //去除@及后面的哈希
    let nameWithoutHash = imageName;
    if (imageName.includes('@')) {
        nameWithoutHash = imageName.split('@')[0];
    }

    // 去除镜像版本
    const [name, tag] = nameWithoutHash.split(':');
    
    // 返回基本的名称和 tag
    // 如果原始输入包含@符号(使用了哈希值)，则不添加默认tag
    if (imageName.includes('@')) {
        return `${name}${tag ? `:${tag}` : ''}`;
    } else {
        // 否则保持原有逻辑，tag 默认为 'latest'
        return `${name}:${tag || 'latest'}`;
    }
}
export default SimpleImageName;
