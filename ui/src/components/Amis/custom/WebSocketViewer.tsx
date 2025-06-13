import React, {useEffect, useRef, useState} from 'react';
import {formatFinalGetUrl} from "@/utils/utils";


// 定义组件的 Props 接口
interface WebSocketViewerProps {
    url: string;
    params: Record<string, string>;
    data: Record<string, any>
}

// WebSocket 组件，支持外部控制
const WebSocketViewerComponent = React.forwardRef<HTMLDivElement, WebSocketViewerProps>(
    ({url, data, params}, _) => {
        url = formatFinalGetUrl({url, data, params});
        const token = localStorage.getItem('token');
        url = url + (url.includes('?') ? '&' : '?') + `token=${token}`;

        const [messages, setMessages] = useState<string[]>([]); // 存储接收到的消息
        const [status, setStatus] = useState('Disconnected'); // 连接状态
        const wsRef = useRef<WebSocket | null>(null); // WebSocket 实例

        const connectWebSocket = () => {
            if (wsRef.current) {
                wsRef.current.close();
            }
            let finalUrl = url;
            if (!finalUrl.startsWith("ws")) {
                const protocol = window.location.protocol === "https:" ? "wss://" : "ws://";
                finalUrl = protocol + location.host + finalUrl;
            }
            const ws = new WebSocket(finalUrl);
            wsRef.current = ws;

            ws.onopen = () => setStatus('Connected');

            ws.onmessage = (event) => {
                try {
                    const parsedData = JSON.parse(event.data);
                    const message = parsedData.data || event.data;
                    setMessages((prev) => [...prev, message]);
                } catch {
                    setMessages((prev) => [...prev, event.data]);
                }
            };

            ws.onerror = () => setStatus('Error');

            ws.onclose = () => {
                setStatus('Disconnected');
                wsRef.current = null;
            };
        };

        const disconnectWebSocket = () => {
            if (wsRef.current) {
                wsRef.current.close();
                wsRef.current = null;
            }
        };

        useEffect(() => {
            connectWebSocket();
            return disconnectWebSocket;
        }, [url]);


        return (
            <div>
                <p style={{fontWeight: 'bold', display: 'none'}}>WebSocket Status: {status}</p>
                <div
                    style={{
                        backgroundColor: '#f5f5f5',
                        padding: '2px',
                        borderRadius: '1px',
                        overflowX: 'auto',
                    }}
                >
                    {messages.map((message, index) => (
                        <pre key={index} style={{whiteSpace: 'pre-wrap', marginBottom: '1px'}}>
                            {message}
                        </pre>
                    ))}
                </div>
            </div>
        );
    })

export default WebSocketViewerComponent;
