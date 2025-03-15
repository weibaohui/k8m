import React, { useEffect, useRef, useState } from "react";
import { render as amisRender } from "amis";
import { formatFinalGetUrl } from "@/utils/utils";

interface WebSocketMarkdownViewerProps {
    url: string;
    params: Record<string, string>;
    data: Record<string, any>;
    width?: string;
}

const WebSocketMarkdownViewerComponent = React.forwardRef<HTMLDivElement, WebSocketMarkdownViewerProps>(
    ({ url, data, params, width }, _) => {
        url = formatFinalGetUrl({ url, data, params });
        const token = localStorage.getItem('token');
        url = url + (url.includes('?') ? '&' : '?') + `token=${token}`;

        const [messages, setMessages] = useState<string[]>([]);
        const [status, setStatus] = useState<string>("Disconnected");
        const [loading, setLoading] = useState<boolean>(true);

        const wsRef = useRef<WebSocket | null>(null);

        useEffect(() => {
            if (wsRef.current) {
                wsRef.current.close();
            }
            const ws = new WebSocket(url);
            wsRef.current = ws;

            ws.onopen = () => setStatus("Connected");

            ws.onmessage = (event) => {
                try {
                    const parsedData = JSON.parse(event.data);
                    const rawMessage = parsedData.data || "";
                    if (rawMessage) {
                        setMessages((prev) => [...prev, rawMessage]);
                        setLoading(false); // 有数据后取消 loading
                    }
                } catch (error) {
                    console.error("Failed to parse WebSocket message:", error);
                    setMessages((prev) => [...prev, event.data]);
                    setLoading(false); // 有数据后取消 loading
                }
            };

            ws.onclose = () => setStatus("Disconnected");
            ws.onerror = () => setStatus("Error");

            return () => {
                wsRef.current?.close();
                wsRef.current = null;
            };
        }, [url]);

        const markdownContent = messages.join("");

        return (
            <>
                <div style={{ width: width || "90vh" }}>
                    <p style={{ display: "none" }}>WebSocket Status: {status}</p>
                    <div
                        style={{
                            backgroundColor: "#f5f5f5",
                            padding: "10px",
                            borderRadius: "5px",
                            maxHeight: "calc(100% - 40px)",
                            overflowX: "auto",
                        }}
                    >
                        {loading ? (
                            <span>Loading...</span>
                        ) : (
                            amisRender({
                                type: "markdown",
                                value: markdownContent,
                                options: {
                                    linkify: true,
                                    html: true,
                                    breaks: true
                                },
                            })
                        )}
                    </div>
                </div>
            </>
        );
    }
);

export default WebSocketMarkdownViewerComponent;
