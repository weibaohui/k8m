import {appendQueryParam, replacePlaceholders} from "@/utils/utils.ts";
import React from "react";

// 定义组件的 props 类型
interface SSEDownloadProps {
    url: string; // 下载的 URL
    data: {
        tailLines?: number;
        sinceTime?: string;
        previous?: boolean;
        timestamps?: boolean;
        sinceSeconds?: number;
    }; // 附加参数
}

// 使用 forwardRef 让外部可以调用下载方法
const SSELogDownloadComponent = React.forwardRef((props: SSEDownloadProps, _) => {
    const [downloading, setDownloading] = React.useState(false);

    let finalUrl = replacePlaceholders(props.url, props.data);
    const params = {
        tailLines: props.data.tailLines,
        sinceTime: props.data.sinceTime,
        previous: props.data.previous,
        timestamps: props.data.timestamps,
        sinceSeconds: props.data.sinceSeconds || ""
    };
    // @ts-ignore
    finalUrl = appendQueryParam(finalUrl, params);

    const handleDownload = () => {
        setDownloading(true); // 设置下载状态为 true，显示提示信息

        const anchor = document.createElement('a');
        anchor.href = finalUrl;
        anchor.download = 'log.txt'; // 设置下载的文件名
        document.body.appendChild(anchor);
        anchor.click();
        document.body.removeChild(anchor);

        // 监听下载结束后取消提示
        setTimeout(() => {
            setDownloading(false);
        }, 1000); // 1秒后关闭提示
    };

    return (
        <div>
            {downloading && (
                <p style={{color: 'red', marginBottom: '10px'}}>正在下载，请稍后...</p>
            )}
            <button
                onClick={handleDownload}
                style={{
                    marginLeft: '10px',
                    padding: '8px 16px',
                    backgroundColor: '#4CAF50',
                    color: 'white',
                    border: 'none',
                    cursor: 'pointer',
                    borderRadius: '4px'
                }}
            >
                下载日志
            </button>
        </div>
    );
});

export default SSELogDownloadComponent;
