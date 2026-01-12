import React, { useEffect, useRef, useState } from 'react';
import { Terminal } from '@xterm/xterm';
import "@xterm/xterm/css/xterm.css";
import { FitAddon } from "@xterm/addon-fit";
import { AttachAddon } from "@xterm/addon-attach";
import { formatFinalGetUrl, ProcessK8sUrlWithCluster } from "@/utils/utils.ts";

interface XTermTestProps {
    url: string;
    params: Record<string, string>;
    data: Record<string, any>;
    width?: string;
    height?: string;
}

const XTermTestComponent = React.forwardRef<HTMLDivElement, XTermTestProps>((props, _) => {
    const terminalRef = useRef<HTMLDivElement | null>(null);
    const wsRef = useRef<WebSocket | null>(null);
    const termRef = useRef<Terminal | null>(null);
    const fitAddonRef = useRef<FitAddon | null>(null);
    const [terminalSize, setTerminalSize] = useState({ cols: 0, rows: 0 });
    const [lastSentSize, setLastSentSize] = useState({ cols: 0, rows: 0 });
    const fitRafIdRef = useRef<number | null>(null);

    // 处理 URL
    let url = props.url;
    url = formatFinalGetUrl({ url, data: props.data, params: props.params });
    const token = localStorage.getItem('token');
    url = url + (url.includes('?') ? '&' : '?') + `token=${token}`;
    url = ProcessK8sUrlWithCluster(url);

    /**
     * 向后端发送终端窗口大小变更消息
     */
    const sendResizeMessage = (cols: number, rows: number) => {
        const ws = wsRef.current;
        if (!ws || ws.readyState !== WebSocket.OPEN) {
            console.warn("WebSocket not ready for resize message");
            return;
        }
        console.log(`Sending resize message: ${cols}x${rows}`);
        const size = JSON.stringify({ cols, rows });
        const payload = new TextEncoder().encode("\x01" + size);
        ws.send(payload);
        setLastSentSize({ cols, rows });
    };

    /**
     * 根据容器实际显示尺寸拟合终端列行
     * 移植自 XTerm.tsx，增强了对宽度的计算逻辑
     */
    const fitTerminal = () => {
        const term = termRef.current;
        const fitAddon = fitAddonRef.current;
        const containerEl = terminalRef.current;

        if (!term || !fitAddon || !containerEl) {
            return;
        }

        if (fitRafIdRef.current != null) {
            cancelAnimationFrame(fitRafIdRef.current);
        }

        fitRafIdRef.current = requestAnimationFrame(() => {
            const termEl = term.element;
            const viewportEl = termEl?.querySelector('.xterm-viewport') as HTMLElement | null;

            // 优先使用 viewport 的宽度，如果不行则使用容器宽度
            const availableWidth = viewportEl?.clientWidth ?? containerEl.clientWidth;
            const availableHeight = containerEl.clientHeight;

            if (availableWidth === 0 || availableHeight === 0) {
                console.warn("Container size is 0, skipping fit");
                return;
            }

            const core = (term as any)?._core;
            const cell = core?._renderService?.dimensions?.css?.cell;
            const cellWidth = typeof cell?.width === 'number' && cell.width > 0 ? cell.width : null;
            const cellHeight = typeof cell?.height === 'number' && cell.height > 0 ? cell.height : null;

            console.log(`Fitting terminal. Container: ${availableWidth}x${availableHeight}, Cell: ${cellWidth}x${cellHeight}`);

            // 策略1: 如果能获取到单元格宽高，手动计算
            if (cellWidth != null && cellHeight != null) {
                const cols = Math.max(2, Math.floor(availableWidth / cellWidth));
                const rows = Math.max(1, Math.floor(availableHeight / cellHeight));

                console.log(`Calculated size from cell dimensions: ${cols}x${rows}`);
                if (cols !== term.cols || rows !== term.rows) {
                    term.resize(cols, rows);
                }
                return;
            }

            // 策略2: 使用 fitAddon.proposeDimensions()
            const proposed = fitAddon.proposeDimensions();
            if (!proposed) {
                console.log("fitAddon proposed no dimensions, calling fit()");
                fitAddon.fit();
                return;
            }

            console.log(`fitAddon proposed: ${proposed.cols}x${proposed.rows}`);
            if (proposed.cols !== term.cols || proposed.rows !== term.rows) {
                term.resize(proposed.cols, proposed.rows);
                return;
            }

            fitAddon.fit();
        });
    };

    useEffect(() => {
        if (!terminalRef.current) return;

        // 1. 初始化 Terminal
        const term = new Terminal({
            cursorBlink: true,
            fontFamily: 'Menlo, Monaco, "Courier New", monospace',
            fontSize: 14,
            theme: {
                background: '#1e1e1e',
            },
            screenReaderMode: true,
            allowProposedApi: true,
        });
        termRef.current = term;

        const fitAddon = new FitAddon();
        fitAddonRef.current = fitAddon;
        term.loadAddon(fitAddon);

        term.open(terminalRef.current);

        // 2. 建立 WebSocket 连接
        let finalUrl = url;
        if (!finalUrl.startsWith("ws")) {
            const protocol = window.location.protocol === "https:" ? "wss://" : "ws://";
            finalUrl = protocol + location.host + finalUrl;
        }
        console.log("Connecting to WebSocket:", finalUrl);
        const ws = new WebSocket(finalUrl);
        wsRef.current = ws;

        // 3. 使用 AttachAddon
        const attachAddon = new AttachAddon(ws);
        term.loadAddon(attachAddon);

        // 4. WebSocket 事件处理
        ws.onopen = () => {
            console.log("WebSocket connected");
            term.focus();

            // 延迟 fit 并同步尺寸，多次尝试确保成功
            setTimeout(fitTerminal, 50);
            setTimeout(fitTerminal, 200);
            setTimeout(() => {
                fitTerminal();
                if (term.cols > 0 && term.rows > 0) {
                    sendResizeMessage(term.cols, term.rows);
                }
            }, 500);

            // 1秒后再试一次，确保字体加载等因素稳定
            setTimeout(() => {
                fitTerminal();
                sendResizeMessage(term.cols, term.rows);
            }, 1000);
        };

        ws.onclose = () => {
            console.log("WebSocket disconnected");
            term.write('\r\n\x1b[31mConnection closed.\x1b[0m\r\n');
        };

        ws.onerror = (err) => {
            console.error("WebSocket error:", err);
            term.write('\r\n\x1b[31mConnection error.\x1b[0m\r\n');
        };

        // 5. 监听 Resize
        term.onResize(({ cols, rows }) => {
            console.log("Terminal resized event:", cols, rows);
            setTerminalSize({ cols, rows });
            sendResizeMessage(cols, rows);
        });

        // 监听窗口大小变化
        const resizeObserver = new ResizeObserver(() => {
            fitTerminal();
        });
        resizeObserver.observe(terminalRef.current);

        return () => {
            console.log("Cleaning up XTermTestComponent");
            if (fitRafIdRef.current != null) {
                cancelAnimationFrame(fitRafIdRef.current);
            }
            resizeObserver.disconnect();
            ws.close();
            term.dispose();
        };
    }, [url]);

    return (
        <div style={{ display: 'flex', flexDirection: 'column', height: props.height || '80vh', width: props.width || '100%' }}>
            <div style={{
                padding: '4px 8px',
                background: '#333',
                color: '#fff',
                fontSize: '12px',
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center'
            }}>
                <div style={{ display: 'flex', gap: '10px' }}>
                    <span>Status: {wsRef.current?.readyState === WebSocket.OPEN ? 'Connected' : 'Disconnected'}</span>
                    <span>Term Size: {terminalSize.cols} x {terminalSize.rows}</span>
                    <span>Last Sent: {lastSentSize.cols} x {lastSentSize.rows}</span>
                </div>
                <button
                    onClick={() => {
                        console.log("Manual fit triggered");
                        fitTerminal();
                        if (termRef.current) {
                            sendResizeMessage(termRef.current.cols, termRef.current.rows);
                        }
                    }}
                    style={{
                        padding: '2px 8px',
                        fontSize: '10px',
                        cursor: 'pointer',
                        backgroundColor: '#444',
                        color: '#fff',
                        border: '1px solid #666',
                        borderRadius: '3px'
                    }}
                >
                    Force Fit & Resize
                </button>
            </div>
            <div
                ref={terminalRef}
                style={{
                    flex: 1,
                    overflow: 'hidden',
                    backgroundColor: '#1e1e1e'
                }}
            />
        </div>
    );
});

export default XTermTestComponent;
