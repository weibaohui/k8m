import React, { useEffect, useRef, useState } from 'react';
import { Terminal } from '@xterm/xterm';
import "@xterm/xterm/css/xterm.css";
import { FitAddon } from "@xterm/addon-fit";
import { AttachAddon } from "@xterm/addon-attach";
import { formatFinalGetUrl, ProcessK8sUrlWithCluster } from "@/utils/utils.ts";
import { WebLinksAddon } from "@xterm/addon-web-links";
import { Unicode11Addon } from "@xterm/addon-unicode11";
import { SerializeAddon } from "@xterm/addon-serialize";
import { WebglAddon } from '@xterm/addon-webgl';
import { SearchAddon } from "@xterm/addon-search";
import { ClipboardAddon } from '@xterm/addon-clipboard';

interface XTermProps {
    url: string;
    params: Record<string, string>;
    data: Record<string, any>;
    width?: string;
    height?: string;
}

const XTermComponent = React.forwardRef<HTMLDivElement, XTermProps>((props, _) => {
    const terminalRef = useRef<HTMLDivElement | null>(null);
    const wsRef = useRef<WebSocket | null>(null);
    const termRef = useRef<Terminal | null>(null);
    const fitAddonRef = useRef<FitAddon | null>(null);
    const [terminalSize, setTerminalSize] = useState({ cols: 0, rows: 0 });
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
            // console.warn("WebSocket not ready for resize message");
            return;
        }

        // 避免发送无效尺寸
        if (cols <= 0 || rows <= 0) {
            console.warn(`Ignoring invalid resize message: ${cols}x${rows}`);
            return;
        }

        console.log(`Sending resize message: ${cols}x${rows}`);
        const size = JSON.stringify({ cols, rows });
        const payload = new TextEncoder().encode("\x01" + size);
        ws.send(payload);
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

            // 如果高度太小（比如小于 50px），可能布局还没准备好，跳过
            if (availableHeight < 50) {
                console.warn(`Container height too small (${availableHeight}px), skipping fit to avoid invalid rows`);
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

            // 如果计算出的 rows 非常小（例如小于 5），可能是误判，进行保护
            if (proposed.rows < 5) {
                console.warn(`Proposed rows too small (${proposed.rows}), ignoring to prevent layout collapse`);
                return;
            }

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



        const webLinksAddon = new WebLinksAddon();
        const unicode11Addon = new Unicode11Addon();
        const serializeAddon = new SerializeAddon();
        const webglAddon = new WebglAddon();
        const searchAddon = new SearchAddon();
        const clipboardAddon = new ClipboardAddon();

        term.loadAddon(fitAddon);
        term.loadAddon(webLinksAddon);
        term.loadAddon(unicode11Addon);
        term.loadAddon(serializeAddon);
        term.loadAddon(webglAddon);
        term.loadAddon(searchAddon);
        term.loadAddon(clipboardAddon);

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

        // 辅助函数：轮询直到容器尺寸稳定
        const waitForLayout = (attempts = 0) => {
            if (attempts > 20) return; // 最多尝试 20 次 (约 2 秒)

            const container = terminalRef.current;
            if (container && container.clientHeight > 50) {
                fitTerminal();
                if (term.cols > 0 && term.rows > 0) {
                    sendResizeMessage(term.cols, term.rows);
                }
            } else {
                setTimeout(() => waitForLayout(attempts + 1), 200);
            }
        };

        // 4. WebSocket 事件处理
        ws.onopen = () => {
            console.log("WebSocket connected");
            term.focus();

            // 立即尝试 fit
            fitTerminal();

            // 启动轮询，等待布局就绪
            waitForLayout();

            // 额外延时保障
            setTimeout(fitTerminal, 500);

            // 等待字体加载
            document.fonts?.ready.then(() => {
                console.log("Fonts loaded, refitting...");
                fitTerminal();
                sendResizeMessage(term.cols, term.rows);
            });
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
                alignItems: 'center',
                flexWrap: 'wrap'
            }}>
                <div style={{ display: 'flex', gap: '10px', alignItems: 'center' }}>
                    <span>Status: {wsRef.current?.readyState === WebSocket.OPEN ? 'Connected' : 'Disconnected'}</span>
                    <span style={{ color: terminalSize.rows < 10 ? 'red' : 'inherit' }}>
                        size: {terminalSize.cols} x {terminalSize.rows}
                    </span>
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
                        borderRadius: '3px',
                        marginLeft: '10px'
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
                    backgroundColor: '#1e1e1e',
                    minHeight: '100px' // 增加最小高度防止塌缩
                }}
            />
        </div>
    );
});

export default XTermComponent;
