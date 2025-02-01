// 格式化 ls -l 短日期格式
const formatLsShortDate = (input: unknown): string => {
    if (typeof input !== "string") return String(input);

    // 获取当前年份
    const year = new Date().getFullYear();

    // 月份缩写映射
    const months: Record<string, string> = {
        Jan: "01", Feb: "02", Mar: "03", Apr: "04", May: "05", Jun: "06",
        Jul: "07", Aug: "08", Sep: "09", Oct: "10", Nov: "11", Dec: "12",
    };

    // 正则匹配："Mon 7 17:12"
    const regex = /^(\w{3})\s+(\d{1,2})\s+(\d{2}):(\d{2})$/;
    const match = input.match(regex);

    if (!match) return input; // 如果格式不对，返回原始值

    // 解析日期
    const month = months[match[1]];
    const day = match[2].padStart(2, "0");
    const hours = match[3].padStart(2, "0");
    const minutes = match[4].padStart(2, "0");

    return `${year}-${month}-${day} ${hours}:${minutes}`;
}
export default formatLsShortDate;
