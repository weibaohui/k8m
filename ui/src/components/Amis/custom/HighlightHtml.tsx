import React from 'react';
import {replacePlaceholders} from "@/utils/utils.ts";

interface HighlightHtmlProps {
    html: string;
    data: Record<string, any>;
    keywords?: string[];
    backgroundColor?: {
        highlight?: string;
        normal?: string;
    };
}

// 使用 forwardRef 以适配 AMIS
const HighlightHtmlComponent = React.forwardRef<HTMLDivElement, HighlightHtmlProps>(
    ({html, data, keywords = [], backgroundColor = {}}, _) => {
        // 获取渲染内容
        const content = replacePlaceholders(html, data);
        // 检查是否包含关键词
        const hasKeyword = keywords.some((keyword) => content.toLowerCase().includes(keyword.toLowerCase()));

        // 设定背景色
        const finalBackgroundColor = hasKeyword ? backgroundColor.highlight || '#ffe6e6' : backgroundColor.normal || '#f0faf0';


        return (
            <pre
                style={{
                    backgroundColor: finalBackgroundColor,
                    padding: '10px',
                    borderRadius: '4px',
                    marginTop: '5px',
                    whiteSpace: 'pre-wrap',
                    wordBreak: 'break-word',
                }}
                dangerouslySetInnerHTML={{__html: content}} // 渲染 HTML 内容
            />
        );
    }
);

export default HighlightHtmlComponent;
