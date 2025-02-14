// 最简化展示镜像名称及tag
const FormatBytes = (input: number) => {
    if (typeof input !== "number" || isNaN(input)) return String(input); // 统一转换为字符串

    if (input === 0) return "0 B";

    const sizes = ["B", "KB", "MB", "GB", "TB", "PB"];
    const i = Math.floor(Math.log(input) / Math.log(1024));
    const value = (input / Math.pow(1024, i)).toFixed(2); // 保留两位小数

    return `${value} ${sizes[i]}`;
}
export default FormatBytes;
