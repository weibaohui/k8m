import React, { useEffect, useRef, useState } from 'react';
import { appendQueryParam, ProcessK8sUrlWithCluster, replacePlaceholders } from "@/utils/utils.ts";
import AnsiToHtml from 'ansi-to-html';
import { Modal, Input, Alert } from 'antd';

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
    finalUrl = ProcessK8sUrlWithCluster(finalUrl);


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
    const [filteredLines, setFilteredLines] = useState<string[] | null>(null);
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
    function filterLinesByCommand(command: string, lines: string[]): { result: string[], keyword: string, ignoreCase: boolean } {
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
        const result: string[] = [];
        lines.forEach((line, idx) => {
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

    return (
        <div ref={dom} style={{ whiteSpace: 'pre-wrap', backgroundColor: 'black', color: 'white', padding: '10px' }}>
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
            {/* 过滤结果提示及关闭按钮 */}
            {filteredLines && (
                <div style={{ background: '#222', color: '#0f0', padding: '4px', marginBottom: '8px' }}>
                    <span>已过滤 {filteredLines.length} 条日志</span>
                    <a style={{ marginLeft: '16px', color: '#f66', cursor: 'pointer' }} onClick={handleCloseFilter}>关闭过滤</a>
                </div>
            )}
            {errorMessage && <div
                style={{ color: errorMessage == "Connected" ? '#00FF00' : 'red' }}>{errorMessage} 共计：{lines.length}行</div>}
            <pre style={{ whiteSpace: 'pre-wrap' }}>
                {(filteredLines || lines).map((line, index) => {
                    let html = converter.toHtml(line);
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
    );
});

export default SSELogDisplayComponent;
