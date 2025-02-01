const unitMultipliers: Record<string, number> = {
    "Ki": 1024,
    "Mi": 1024 ** 2,
    "Gi": 1024 ** 3,
    "Ti": 1024 ** 4,
    "Pi": 1024 ** 5,
    "Ei": 1024 ** 6,
};

const convertMemory = (input: string) => {
    const match = input?.toString().match(/(\d+)([KMGTPE]i)/i);
    if (!match) return input; // 无法匹配时，返回原始输入

    const value = parseInt(match[1], 10);
    const unit = match[2];
    const bytes = value * unitMultipliers[unit];

    if (bytes < 1024 ** 2) {
        return `${(bytes / 1024).toFixed(1)}Ki`;
    } else if (bytes < 500 * 1024 ** 2) {
        return `${(bytes / 1024 ** 2).toFixed(1)}Mi`;
    } else {
        return `${(bytes / 1024 ** 3).toFixed(1)}Gi`;
    }
};


export default convertMemory;
