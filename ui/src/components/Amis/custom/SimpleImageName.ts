// 最简化展示镜像名称及tag
const SimpleImageName = (input: string) => {
    // 检查输入是否为空
    if (!input||input===''||input===undefined) {
        return '';
    }
    // 分割镜像名称，移除注册表地址部分
    const parts = input.split('/');
    const imageName = parts[parts.length - 1]; // 获取最后一部分
    // 去除镜像版本
    const [name, tag] = imageName.split(':');
    // 返回基本的名称和 tag，tag 默认为 "latest"
    return `${name}:${tag || 'latest'}`;
}
export default SimpleImageName;
