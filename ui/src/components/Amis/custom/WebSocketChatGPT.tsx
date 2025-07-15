import React, { useEffect, useRef, useState } from "react";
import { render as amisRender } from "amis";
import { formatFinalGetUrl } from "@/utils/utils";
import { Button, Flex, Space, Typography } from "antd";
import {
    BulbOutlined,
    InfoCircleOutlined,
    PlusOutlined,
    RocketOutlined,
    SmileOutlined,
    UserOutlined
} from "@ant-design/icons";
import { Bubble, BubbleProps, Prompts, PromptsProps, Sender, Welcome } from "@ant-design/x";
import { Modal } from "antd";

interface WebSocketChatGPTProps {
    url: string;
    params: Record<string, string>;
    data: Record<string, any>;
}

const WebSocketChatGPT = React.forwardRef<HTMLDivElement, WebSocketChatGPTProps>(
    ({ url, data, params }, _) => {
        url = formatFinalGetUrl({ url, data, params });
        const token = localStorage.getItem('token');
        url = url + (url.includes('?') ? '&' : '?') + `token=${token}`;

        let historyUrl = '/ai/chat/ws_chatgpt/history'
        historyUrl = historyUrl + (historyUrl.includes('?') ? '&' : '?') + `token=${token}`;

        let historyResetUrl = '/ai/chat/ws_chatgpt/history/reset'
        historyResetUrl = historyResetUrl + (historyResetUrl.includes('?') ? '&' : '?') + `token=${token}`;

        const [messages, setMessages] = useState<{ role: "user" | "ai"; content: string }[]>([]);
        const [status, setStatus] = useState<string>("Disconnected");
        const [inputMessage, setInputMessage] = useState<string>(""); // 用户输入的消息
        const wsRef = useRef<WebSocket | null>(null);
        const messageContainerRef = useRef<HTMLDivElement | null>(null); // 滚动到底部
        const [loading, setLoading] = useState<boolean>(false);

        console.log(status)
        useEffect(() => {
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

            ws.onopen = () => setStatus("Connected");

            const formatToolCallResult = (message: any) => {
                try {
                    const data = JSON.parse(message);
                    if (data.tool_name && data.parameters && data.result) {
                        return `🛠️ **工具调用**: ${data.tool_name}\n\n📝 **参数**:\n\`\`\`json\n${JSON.stringify(data.parameters, null, 2)}\n\`\`\`\n\n🎯 **结果**:\n${data.result}\n`;
                    }
                    return message;
                } catch {
                    return message;
                }
            };

            ws.onmessage = (event) => {
                try {
                    const rawMessage = event.data || "";
                    if (rawMessage) {
                        setMessages((prev) => {
                            // 找到最后一个 AI 占位符并替换为实际消息
                            const aiPlaceholderIndex = prev.findIndex(
                                (msg) => msg.role === "ai" && msg.content === "thinking"
                            );
                            const formattedMessage = formatToolCallResult(rawMessage);
                            if (aiPlaceholderIndex !== -1) {
                                return prev.map((msg, index) =>
                                    index === aiPlaceholderIndex ? { ...msg, content: formattedMessage } : msg
                                );
                            }
                            // 如果没有找到占位符，默认行为
                            if (prev.length === 0 || prev[prev.length - 1].role !== "ai") {
                                return [...prev, { role: "ai", content: formattedMessage }];
                            } else {
                                return prev.map((msg, index) =>
                                    index === prev.length - 1 ? { ...msg, content: msg.content + formattedMessage } : msg
                                );
                            }
                        });
                    }
                } catch (error) {
                    console.error("Failed to parse WebSocket message:", error);
                    setMessages((prev) => [...prev, event.data]);
                }
            };

            ws.onclose = () => setStatus("Disconnected");
            ws.onerror = () => setStatus("Error");

            return () => {
                wsRef.current?.close();
                wsRef.current = null;
            };
        }, [url]);

        // 发送消息
        const handleSendMessage = () => {
            setLoading(true);
            if (!inputMessage.trim()) return;

            if (wsRef.current) {
                wsRef.current.send(inputMessage);
            }

            // 立即显示用户消息，并准备新的 AI 回复条目
            setMessages((prev) => [
                ...prev,
                { role: "user", content: `${inputMessage}` },
                { role: "ai", content: "thinking" } // 插入AI思考中的占位符
            ]);

            setInputMessage(""); // 清空输入框
            setLoading(false);
        };

        // 滚动到底部
        const scrollToBottom = () => {
            if (messageContainerRef.current) {
                messageContainerRef.current.scrollTop = messageContainerRef.current.scrollHeight;
            }
        };

        useEffect(() => {
            scrollToBottom();
        }, [messages]);
        const renderMarkdown: BubbleProps['messageRender'] = (content: string) => {
            return amisRender({
                type: "markdown",
                value: content
            })
        };
        const items: PromptsProps['items'] = [
            {
                key: '1',
                icon: <BulbOutlined style={{ color: '#FFD700' }} />,
                label: 'yaml编写',
                description: '请给我一个基本的nginx 部署yaml',
            },
            {
                key: '2',
                icon: <InfoCircleOutlined style={{ color: '#1890FF' }} />,
                label: '网络',
                description: '请解释下Deploy中的HostNetwork如何配置？',
            },
            {
                key: '3',
                icon: <RocketOutlined style={{ color: '#722ED1' }} />,
                label: '自动应用',
                description: '请给我一个基本的nginx 部署yaml，并部署到集群中',
            },
            {
                key: '4',
                icon: <SmileOutlined style={{ color: '#52C41A' }} />,
                label: 'Yaml模板',
                description: '请给我一个基本的nginx 部署yaml，并保存为模板',
            },

        ];
        const fooAvatar: React.CSSProperties = {
            color: '#f56a00',
            backgroundColor: '#fde3cf',
        };

        const barAvatar: React.CSSProperties = {
            color: '#fff',
            backgroundColor: '#87d068',
        };
        return (
            <>
                <div style={{ width: "100%", height: "100%", minHeight: "600px" }}>

                    {
                        messages.length == 0 && <>
                            <Welcome
                                title="ChatGPT"
                                description="我是k8m的AI小助手，你可以问我任何关于kubernetes的问题，我尽量给你提供最准确的答案。"
                                style={{
                                    backgroundImage: 'linear-gradient(97deg, #f2f9fe 0%, #f7f3ff 100%)',
                                    borderStartStartRadius: 4,
                                }}
                            />
                            <Prompts
                                title="✨ 奇思妙想和创新的火花"
                                items={items}
                                wrap
                                styles={{
                                    item: {
                                        flex: 'none',
                                        width: 'calc(50% - 6px)',
                                    },
                                }}
                                onItemClick={(info) => {
                                    setInputMessage(`${info.data.description}`);

                                }}
                            />
                        </>
                    }

                    <Flex gap="middle" vertical>
                        {messages.map((msg) => (
                            <>
                                <Bubble
                                    placement={msg.role === "user" ? "end" : "start"}
                                    content={msg.content}
                                    avatar={{
                                        icon: msg.role === "user"
                                            ? <UserOutlined />
                                            : <RocketOutlined />,
                                        style: msg.role === "user" ? barAvatar : fooAvatar,
                                    }}
                                    messageRender={renderMarkdown}
                                    loading={msg.role === 'ai' && msg.content === 'thinking'}
                                />
                            </>
                        ))}
                    </Flex>

                    <Flex vertical gap="middle" className="mt-20 mb-20">
                        {
                            messages.length > 0 && <>

                                <Space size="small">
                                    <Button
                                        onClick={() => {
                                            setMessages([]);
                                        }}
                                        icon={<PlusOutlined />}
                                        style={{
                                            width: '100px',
                                            backgroundImage: 'linear-gradient(97deg, #f2f9fe 0%, #f7f3ff 100%)',
                                            borderStartStartRadius: 4,
                                            borderStartEndRadius: 4,
                                        }}
                                    >
                                        新会话
                                    </Button>
                                    <Button
                                        onClick={() => {
                                            fetch('/k8m/api' + historyUrl)
                                                .then(response => response.json())
                                                .then(data => {
                                                    const itemCount = data.data ? data.data.length : 0;
                                                    Modal.success({
                                                        content: `对话历史包含 ${itemCount} 条记录。`,
                                                    });
                                                });
                                        }}
                                        icon={<InfoCircleOutlined />}
                                        style={{
                                            width: '100px',
                                            backgroundImage: 'linear-gradient(97deg, #f2f9fe 0%, #f7f3ff 100%)',
                                            borderStartStartRadius: 4,
                                            borderStartEndRadius: 4,
                                        }}
                                    >
                                        对话历史
                                    </Button>
                                    <Button
                                        onClick={() => {
                                            fetch('/k8m/api' + historyResetUrl)
                                                .then(response => response.json())
                                                .then(_ => {
                                                    Modal.success({
                                                        content: '对话历史已清空。',
                                                    });
                                                });
                                        }}
                                        icon={<InfoCircleOutlined />}
                                        style={{
                                            width: '100px',
                                            backgroundImage: 'linear-gradient(97deg, #f2f9fe 0%, #f7f3ff 100%)',
                                            borderStartStartRadius: 4,
                                            borderStartEndRadius: 4,
                                        }}
                                    >
                                        清空历史
                                    </Button>
                                </Space>
                            </>
                        }


                        <Sender
                            loading={loading}
                            value={inputMessage}
                            onChange={(v) => {
                                setInputMessage(v);
                            }}
                            onSubmit={() => {
                                setInputMessage('');
                                handleSendMessage();
                            }}
                            onCancel={() => {
                                setLoading(false);
                            }}
                            actions={(_, info) => {
                                const { SendButton, ClearButton } = info.components;

                                return (
                                    <Space size="small">
                                        <Typography.Text type="secondary">
                                            <small>`Shift + Enter` 换行</small>
                                        </Typography.Text>
                                        <ClearButton />
                                        <SendButton type="primary" disabled={false} />
                                    </Space>
                                );
                            }}
                        />
                    </Flex>
                </div>
            </>
        );
    }
);

export default WebSocketChatGPT;
