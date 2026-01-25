import React, { useEffect, useRef, useState } from 'react';
import { appendQueryParam, ProcessK8sUrlWithCluster, replacePlaceholders } from "@/utils/utils.ts";
import AnsiToHtml from 'ansi-to-html';
import { Modal, Input, Alert, Button, Switch, Card, Tag, Collapse, Space, message, Select } from 'antd';
import { RobotOutlined, ClockCircleOutlined } from '@ant-design/icons';

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
        labelSelector?: string;  // 对应 -l app=nginx
        allPods?: boolean;       // 对应 --all-pods
        allContainers?: boolean; // 对应 --all-containers
        namespace?: string;      // 命名空间，用于 AI 上下文
        podName?: string;        // Pod 名称，用于 AI 上下文
    };
    // 扩展属性，用于接收外部传入的控制元素
    extraControls?: React.ReactNode;
}

interface AISummaryData {
    status: 'normal' | 'warning' | 'error';
    summary: string;
    issues?: string[];
    reasons?: string[];
    suggestions?: string[];
}

interface LogItem {
    type: 'log' | 'summary';
    content: string | AISummaryData;
    timestamp?: number;
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
        sinceSeconds: props.data.sinceSeconds || "",
        labelSelector: props.data.labelSelector,
        allPods: props.data.allPods,
        allContainers: props.data.allContainers
    };
    // @ts-ignore
    let finalUrl = appendQueryParam(url, params);
    const token = localStorage.getItem('token');
    //拼接url token
    finalUrl = finalUrl + (finalUrl.includes('?') ? '&' : '?') + `token=${token}`;
    finalUrl = ProcessK8sUrlWithCluster(finalUrl);


    const dom = useRef<HTMLDivElement | null>(null);
    const eventSourceRef = useRef<EventSource | null>(null);
    const [errorMessage, setErrorMessage] = useState('');
    const [lines, setLines] = useState<LogItem[]>([]);


    // 连接 SSE 服务器
    const connectSSE = () => {
        if (eventSourceRef.current) {
            eventSourceRef.current.close();
        }

        eventSourceRef.current = new EventSource(finalUrl);

        eventSourceRef.current.addEventListener('message', (event) => {
            const newLine = event.data;
            setLines((prevLines) => [...prevLines, { type: 'log', content: newLine, timestamp: Date.now() }]);
        });
        eventSourceRef.current.addEventListener('open', (_) => {
            // setErrorMessage('Connected');
        });
        eventSourceRef.current.addEventListener('error', (_) => {
            if (eventSourceRef.current?.readyState === EventSource.CLOSED) {
                // setErrorMessage('连接已关闭');
            } else if (eventSourceRef.current?.readyState === EventSource.CONNECTING) {
                // setErrorMessage('正在尝试重新连接...');
            } else {
                // setErrorMessage('发生未知错误...');
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
        setLines([]); // 清空日志
        setErrorMessage('');
        connectSSE();
        return () => {
            disconnectSSE();
        };
    }, [finalUrl]);


    // 创建一个转换器实例
    const converter = new AnsiToHtml();
    const [filterModalVisible, setFilterModalVisible] = useState(false);
    const [filterCommand, setFilterCommand] = useState('');
    const [filteredLines, setFilteredLines] = useState<LogItem[] | null>(null);
    const inputRef = useRef<any>(null);

    // 监听ctrl+f快捷键，弹出命令行输入框
    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if ((e.ctrlKey || e.metaKey) && e.key === 'f') {
                e.preventDefault();
                setFilterModalVisible(true);
                setTimeout(() => inputRef.current?.focus(), 100);
            }
        };
        window.addEventListener('keydown', handleKeyDown);
        return () => window.removeEventListener('keydown', handleKeyDown);
    }, []);

    // 监听键盘事件，Ctrl+F 弹窗
    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            // Ctrl+F 打开过滤弹窗
            if (e.ctrlKey && e.key === 'f') {
                e.preventDefault();
                setFilterModalVisible(true);
            }

        };
        window.addEventListener('keydown', handleKeyDown);
        return () => {
            window.removeEventListener('keydown', handleKeyDown);
        };
    }, [filterModalVisible]);

    /**
     * 解析grep命令并过滤日志
     * 支持grep xxx -A n -B m -i
     * 返回过滤后的行和关键字
     */
    function filterLinesByCommand(command: string, lines: LogItem[]): { result: LogItem[], keyword: string, ignoreCase: boolean } {
        // 简单解析命令
        const grepMatch = command.match(/grep\s+([^-\s]+)(.*)/);
        if (!grepMatch) return { result: [], keyword: '', ignoreCase: false };
        const keyword = grepMatch[1];
        const options = grepMatch[2] || '';
        let after = 0, before = 0;
        let ignoreCase = false;
        if (/\s-i(\s|$)/.test(options)) ignoreCase = true;
        const afterMatch = options.match(/-A\s*(\d+)/);
        const beforeMatch = options.match(/-B\s*(\d+)/);
        if (afterMatch) after = parseInt(afterMatch[1]);
        if (beforeMatch) before = parseInt(beforeMatch[1]);
        // 过滤逻辑
        const result: LogItem[] = [];
        lines.forEach((item, idx) => {
            if (item.type !== 'log' || typeof item.content !== 'string') return;
            const line = item.content;
            let match = false;
            if (ignoreCase) {
                match = line.toLowerCase().includes(keyword.toLowerCase());
            } else {
                match = line.includes(keyword);
            }
            if (match) {
                const start = Math.max(0, idx - before);
                const end = Math.min(lines.length, idx + after + 1);
                for (let i = start; i < end; i++) {
                    if (!result.includes(lines[i])) {
                        result.push(lines[i]);
                    }
                }
            }
        });
        return { result, keyword, ignoreCase };
    }

    // 打开过滤弹窗时，输入框默认填充为 grep 
    useEffect(() => {
        if (filterModalVisible) {
            // 如果当前输入为空，自动填充为 'grep '
            setFilterCommand(cmd => (cmd && cmd.trim() !== '' ? cmd : 'grep '));
            setTimeout(() => inputRef.current?.focus(), 100);
        }
    }, [filterModalVisible]);
    // 新增：过滤命令输入错误提示
    const [filterError, setFilterError] = useState<string>('');

    // 确认过滤命令，执行过滤
    const handleFilterOk = () => {
        // 检查命令是否合法（不能只有grep或无关键字）
        const grepMatch = filterCommand.match(/grep\s+([^-\s]+)(.*)/);
        if (!grepMatch || !grepMatch[1] || grepMatch[1].trim() === '' || filterCommand.trim() === 'grep') {
            setFilterError('请输入有效的grep命令，例如：grep 关键字');
            return;
        }
        setFilterError('');
        const { result, keyword, ignoreCase } = filterLinesByCommand(filterCommand, lines);
        setFilteredLines(result);
        setFilterKeyword(keyword);
        setIgnoreCaseFilter(ignoreCase);
        setFilterModalVisible(false);
    };

    // 新增：保存当前过滤关键字
    const [filterKeyword, setFilterKeyword] = useState<string>('');
    // 新增：保存是否忽略大小写
    const [ignoreCaseFilter, setIgnoreCaseFilter] = useState<boolean>(false);

    // 取消过滤
    const handleFilterCancel = () => {
        setFilterModalVisible(false);
    };
    // 关闭过滤，恢复原始日志
    const handleCloseFilter = () => {
        setFilteredLines(null);
        setFilterCommand('');
    };

    // AI Logic
    const [aiEnabled, setAiEnabled] = useState(false);
    const [summaryInterval, setSummaryInterval] = useState(2 * 60 * 1000); // Default 2 minutes
    const [askModalVisible, setAskModalVisible] = useState(false);
    const [askQuestion, setAskQuestion] = useState('');
    const [askAnswer, setAskAnswer] = useState('');
    const [asking, setAsking] = useState(false);
    const lastSummaryTimeRef = useRef(Date.now());
    const linesRef = useRef<LogItem[]>([]);
    const filteredLinesRef = useRef<LogItem[] | null>(null);

    // Sync lines to ref for interval access
    useEffect(() => {
        linesRef.current = lines;
    }, [lines]);

    // Sync filteredLines to ref
    useEffect(() => {
        filteredLinesRef.current = filteredLines;
    }, [filteredLines]);

    // Auto Summary
    useEffect(() => {
        let interval: NodeJS.Timeout;
        if (aiEnabled) {
            // Immediate trigger when enabled
            triggerSummary();

            interval = setInterval(() => {
                const now = Date.now();
                const lastTime = lastSummaryTimeRef.current;
                const currentLines = linesRef.current;

                // 1. Time based: custom interval
                if (now - lastTime > summaryInterval) {
                    triggerSummary();
                    return;
                }

                // 2. Error rate based: > 20 errors since last summary (and at least 30s interval)
                if (now - lastTime > 30 * 1000) {
                    const recentLogs = currentLines.filter(l => l.type === 'log' && l.timestamp && l.timestamp > lastTime);
                    const errorCount = recentLogs.filter(l => /error|exception|fail|panic/i.test(l.content as string)).length;
                    if (errorCount > 20) {
                        triggerSummary();
                    }
                }
            }, 5000);
        }
        return () => clearInterval(interval);
    }, [aiEnabled, summaryInterval]);

    const triggerSummary = async () => {
        const currentLines = linesRef.current;
        if (currentLines.length === 0) return;

        // Get logs since last summary
        const logItems = currentLines.filter(l => l.type === 'log' && l.timestamp && l.timestamp > lastSummaryTimeRef.current);
        if (logItems.length === 0) return;

        lastSummaryTimeRef.current = Date.now();
        const logContent = logItems.map(l => l.content).join('\n').slice(-5000); // Limit size

        try {
            const token = localStorage.getItem('token');
            const res = await fetch(`/mgm/plugins/ai/chat/log/summary`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ data: logContent })
            });

            const reader = res.body?.getReader();
            const decoder = new TextDecoder();
            let fullText = '';
            while (true) {
                const { done, value } = await reader!.read();
                if (done) break;
                const chunk = decoder.decode(value);
                const lines = chunk.split('\n');
                for (const line of lines) {
                    if (line.startsWith('data: ')) {
                        const data = line.slice(6);
                        if (data === '[DONE]') continue;
                        try {
                            const json = JSON.parse(data);
                            if (json.choices && json.choices[0].delta.content) {
                                fullText += json.choices[0].delta.content;
                            }
                        } catch (e) { }
                    }
                }
            }

            try {
                const jsonMatch = fullText.match(/```json\n([\s\S]*)\n```/) || fullText.match(/\{[\s\S]*\}/);
                const jsonStr = jsonMatch ? jsonMatch[0].replace(/```json|```/g, '') : fullText;
                const summaryData = JSON.parse(jsonStr);

                setLines(prev => [...prev, {
                    type: 'summary',
                    content: summaryData,
                    timestamp: Date.now()
                }]);
            } catch (e) {
                console.error("Failed to parse AI summary", fullText);
            }

        } catch (e) {
            console.error("Failed to fetch summary", e);
        }
    };

    const handleAskAI = async () => {
        if (!askQuestion) return;
        setAsking(true);
        setAskAnswer('');

        // 优先使用过滤后的日志，如果没有过滤则使用所有日志
        const currentLines = filteredLinesRef.current || linesRef.current;
        const recentLogs = currentLines
            .filter(l => l.type === 'log')
            .slice(-100)
            .map(l => l.content)
            .join('\n');

        try {
            const token = localStorage.getItem('token');
            const response = await fetch(`/mgm/plugins/ai/chat/log/ask`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ data: recentLogs, question: askQuestion })
            });

            const reader = response.body?.getReader();
            const decoder = new TextDecoder();

            while (true) {
                const { done, value } = await reader!.read();
                if (done) break;
                const chunk = decoder.decode(value);
                const lines = chunk.split('\n');
                for (const line of lines) {
                    if (line.startsWith('data: ')) {
                        const data = line.slice(6);
                        if (data === '[DONE]') continue;
                        try {
                            const json = JSON.parse(data);
                            if (json.choices && json.choices[0].delta.content) {
                                setAskAnswer(prev => prev + json.choices[0].delta.content);
                            }
                        } catch (e) { }
                    }
                }
            }
        } catch (e) {
            message.error("请求失败");
        } finally {
            setAsking(false);
        }
    };

    // Render AI Summary Card
    const renderSummaryCard = (item: LogItem, index: number) => {
        const summary = item.content as AISummaryData;
        return (
            <Card key={index} size="small" style={{ marginBottom: 8, border: '1px solid #1890ff', background: '#001529' }}>
                <Space direction="vertical" style={{ width: '100%' }}>
                    <Space>
                        <Tag color={summary.status === 'error' ? 'red' : summary.status === 'warning' ? 'orange' : 'green'}>
                            {summary.status.toUpperCase()}
                        </Tag>
                        <span style={{ color: '#fff', fontWeight: 'bold' }}>AI 智能总结</span>
                        <span style={{ color: '#aaa' }}>{new Date(item.timestamp || 0).toLocaleTimeString()}</span>
                    </Space>
                    <div style={{ color: '#fff' }}>{summary.summary}</div>
                    {summary.issues && summary.issues.length > 0 && (
                        <Collapse ghost size="small">
                            <Collapse.Panel header={<span style={{ color: '#ff4d4f' }}>发现 {summary.issues.length} 个异常</span>} key="1">
                                <ul style={{ color: '#ddd' }}>
                                    {summary.issues.map((issue, i) => <li key={i}>{issue}</li>)}
                                </ul>
                                {summary.suggestions && (
                                    <div style={{ marginTop: 8 }}>
                                        <div style={{ color: '#40a9ff' }}>建议：</div>
                                        <ul style={{ color: '#ddd' }}>
                                            {summary.suggestions.map((s, i) => <li key={i}>{s}</li>)}
                                        </ul>
                                    </div>
                                )}
                            </Collapse.Panel>
                        </Collapse>
                    )}
                </Space>
            </Card>
        );
    };

    return (
        <div style={{ display: 'flex', flexDirection: 'column', height: '100%', backgroundColor: '#1f1f1f' }}>
            {/* Controls Toolbar */}
            <div style={{ padding: '8px 16px', borderBottom: '1px solid #333', background: '#2c2c2c', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                {/* Left side: External Controls (e.g., Container Select, Options, Download) */}
                <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                    {props.extraControls}
                </div>

                {/* Right side: AI Controls */}
                <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                    <Space>
                        <RobotOutlined style={{ color: '#1890ff', fontSize: '16px' }} />
                        <span style={{ color: '#fff', fontWeight: 'bold' }}>AI 智能解析</span>
                        <Switch
                            checked={aiEnabled}
                            onChange={setAiEnabled}
                            checkedChildren="开启"
                            unCheckedChildren="关闭"
                            size="small"
                        />
                        {aiEnabled && (
                            <Select
                                size="small"
                                value={summaryInterval}
                                onChange={setSummaryInterval}
                                style={{ width: 100 }}
                                options={[
                                    { label: '30 秒', value: 30 * 1000 },
                                    { label: '1 分钟', value: 60 * 1000 },
                                    { label: '2 分钟', value: 2 * 60 * 1000 },
                                    { label: '5 分钟', value: 5 * 60 * 1000 },
                                    { label: '10 分钟', value: 10 * 60 * 1000 },
                                ]}
                                prefix={<ClockCircleOutlined />}
                            />
                        )}
                        <Button type="link" onClick={() => setAskModalVisible(true)} style={{ color: '#40a9ff', paddingLeft: 8 }}>
                            询问 AI
                        </Button>
                    </Space>
                    {aiEnabled && <Tag color="blue" style={{ marginRight: 0 }}>自动总结中...</Tag>}
                </div>
            </div>

            {/* AI Ask Modal */}
            <Modal
                title="AI 智能问答"
                open={askModalVisible}
                onCancel={() => setAskModalVisible(false)}
                footer={null}
                width={600}
            >
                <div style={{ marginBottom: 16 }}>
                    <Input.TextArea
                        rows={3}
                        value={askQuestion}
                        onChange={e => setAskQuestion(e.target.value)}
                        placeholder="请输入关于当前日志的问题..."
                    />
                    <div style={{ marginTop: 8, textAlign: 'right' }}>
                        <Button type="primary" onClick={handleAskAI} loading={asking} disabled={!askQuestion}>
                            提问
                        </Button>
                    </div>
                </div>
                {askAnswer && (
                    <Card size="small" style={{ background: '#f5f5f5' }}>
                        <div style={{ whiteSpace: 'pre-wrap' }}>{askAnswer}</div>
                    </Card>
                )}
            </Modal>

            {/* 过滤命令弹窗 */}
            <Modal
                title="日志过滤 (如: grep 关键字 -A 2 -B 2 -i )"
                open={filterModalVisible}
                onOk={handleFilterOk}
                onCancel={handleFilterCancel}
                okText="确定"
                cancelText="取消"
            >

                <Input
                    ref={inputRef}
                    value={filterCommand}
                    onChange={e => { setFilterCommand(e.target.value); setFilterError(''); }}
                    onPressEnter={handleFilterOk}
                    placeholder="请输入grep命令"
                />
                {filterError && <div style={{ color: 'red', marginTop: '8px' }}>{filterError}</div>}
                <Alert
                    message="参数说明：-A n 表示匹配后显示后面 n 行，-B m 表示匹配前显示前面 m 行，-i 表示忽略大小写。"
                    type="success"
                    style={{ marginBottom: 12 }}
                />
            </Modal>

            {/* Main Content Area */}
            <div ref={dom} style={{ flex: 1, display: 'flex', overflow: 'hidden' }}>
                {/* Left Column: Log Content */}
                <div style={{ flex: 1, backgroundColor: 'black', color: 'white', padding: '10px', overflow: 'auto' }}>
                    {/* 过滤结果提示及关闭按钮 */}
                    {filteredLines && (
                        <div style={{ background: '#222', color: '#0f0', padding: '4px', marginBottom: '8px', position: 'sticky', top: 0 }}>
                            <span>已过滤 {filteredLines.length} 条日志</span>
                            <a style={{ marginLeft: '16px', color: '#f66', cursor: 'pointer' }} onClick={handleCloseFilter}>关闭过滤</a>
                        </div>
                    )}
                    {errorMessage && <div
                        style={{ color: errorMessage == "Connected" ? '#00FF00' : 'red' }}>{errorMessage} 共计：{lines.length}行</div>}

                    <pre style={{ whiteSpace: 'pre-wrap', margin: 0 }}>
                        {(filteredLines || lines).map((item, index) => {
                            if (item.type === 'summary') {
                                return null; // Summary rendered in right column
                            }

                            const lineContent = item.content as string;
                            let html = converter.toHtml(lineContent);
                            // 关键字高亮（仅过滤时生效）
                            if (filteredLines && filterKeyword) {
                                // 使用正则替换所有关键字为黄色背景
                                const reg = new RegExp(filterKeyword.replace(/[.*+?^${}()|[\]\\]/g, "\\$&"), ignoreCaseFilter ? 'gi' : 'g');
                                html = html.replace(reg, '<span style="background:yellow;color:black;">' + filterKeyword + '</span>');
                            }
                            return (
                                <div
                                    key={index}
                                    dangerouslySetInnerHTML={{
                                        __html: html
                                    }}
                                />
                            );
                        })}
                    </pre>
                </div>

                {/* Right Column: AI Summary Cards */}
                {aiEnabled && (
                    <div style={{ width: '320px', backgroundColor: '#141414', borderLeft: '1px solid #333', overflowY: 'auto', padding: '10px' }}>
                        <div style={{ color: '#1890ff', marginBottom: 12, fontWeight: 'bold', borderBottom: '1px solid #333', paddingBottom: 8 }}>AI 智能总结列表</div>
                        {lines
                            .filter(item => item.type === 'summary')
                            .slice().reverse() // Show newest first
                            .map((item, index) => renderSummaryCard(item, index))
                        }
                    </div>
                )}
            </div>
        </div>
    );
});

export default SSELogDisplayComponent;
