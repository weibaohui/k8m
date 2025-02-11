import React, {useEffect, useRef, useState} from 'react';
import {appendQueryParam, replacePlaceholders} from "@/utils/utils.ts";

// 定义组件的 Props 接口
interface SSEComponentProps {
    url: string;
    data: {
        tailLines?: number;
        sinceTime?: string;
        follow?: boolean;
        previous?: boolean;
        timestamps?: boolean;
        sinceSeconds?: number;
    };
}

// SSE 组件，使用 forwardRef 让父组件可以手动控制
const SSELogDisplayComponent = React.forwardRef((props: SSEComponentProps, _) => {
    const url = replacePlaceholders(props.url, props.data);
    const params = {
        tailLines: props.data.tailLines,
        sinceTime: props.data.sinceTime,
        follow: props.data.follow,
        previous: props.data.previous,
        timestamps: props.data.timestamps,
        sinceSeconds: props.data.sinceSeconds || ""
    };
    // @ts-ignore
    let finalUrl = appendQueryParam(url, params);
    const token = localStorage.getItem('token');
    //拼接url token
    finalUrl = finalUrl + (finalUrl.includes('?') ? '&' : '?') + `token=${token}`;


    const dom = useRef<HTMLDivElement | null>(null);
    const eventSourceRef = useRef<EventSource | null>(null);
    const [errorMessage, setErrorMessage] = useState('');
    const [lines, setLines] = useState<string[]>([]);


    // 连接 SSE 服务器
    const connectSSE = () => {
        if (eventSourceRef.current) {
            eventSourceRef.current.close();
        }

        eventSourceRef.current = new EventSource(finalUrl);

        eventSourceRef.current.addEventListener('message', (event) => {
            const newLine = event.data;
            setLines((prevLines) => [...prevLines, newLine]);
        });

        eventSourceRef.current.addEventListener('error', (_) => {
            if (eventSourceRef.current?.readyState === EventSource.CLOSED) {
                setErrorMessage('无日志 连接已关闭');
            } else if (eventSourceRef.current?.readyState === EventSource.CONNECTING) {
                setErrorMessage('无日志 正在尝试重新连接...');
            } else {
                setErrorMessage('发生未知错误...');
            }
            eventSourceRef.current?.close();
        });
    };

    // 关闭 SSE 连接
    const disconnectSSE = () => {
        if (eventSourceRef.current) {
            eventSourceRef.current.close();
            eventSourceRef.current = null;
        }
    };

    useEffect(() => {
        connectSSE();

        return () => {
            disconnectSSE();
        };
    }, [finalUrl]);


    return (
        <div ref={dom} style={{whiteSpace: 'pre-wrap'}}>
            {errorMessage && <div style={{color: 'red'}}>{errorMessage}</div>}

            {lines.map((line, index) => (
                <div key={index} style={{display: 'flex', alignItems: 'center'}}>

                    <span>{line}</span>
                </div>
            ))}
        </div>
    );
});

export default SSELogDisplayComponent;
