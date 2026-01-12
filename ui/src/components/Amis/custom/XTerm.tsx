import React, { useEffect, useRef, useState } from 'react';
import { Terminal } from '@xterm/xterm'
import "@xterm/xterm/css/xterm.css";
import { formatFinalGetUrl, ProcessK8sUrlWithCluster } from "@/utils/utils.ts";
import { AttachAddon } from "@xterm/addon-attach";
import { FitAddon } from "@xterm/addon-fit";
import { WebLinksAddon } from "@xterm/addon-web-links";
import { Unicode11Addon } from "@xterm/addon-unicode11";
import { SerializeAddon } from "@xterm/addon-serialize";
import { WebglAddon } from '@xterm/addon-webgl';
import { SearchAddon } from "@xterm/addon-search";
import { ClipboardAddon } from '@xterm/addon-clipboard';

interface XTermProps {
    url: string;
    params: Record<string, string>;
    data: Record<string, any>
    width?: string,
    height?: string
}

const XTermComponent = React.forwardRef<HTMLDivElement, XTermProps>(
    ({ url, data, params, width, height }, _) => {
        url = formatFinalGetUrl({ url, data, params });
        const token = localStorage.getItem('token');
        url = url + (url.includes('?') ? '&' : '?') + `token=${token}`;
        url = ProcessK8sUrlWithCluster(url)
        const wsRef = useRef<WebSocket | null>(null);
        const terminalRef = useRef<HTMLDivElement | null>(null);
        const fitAddonRef = useRef<FitAddon | null>(null);
        const termRef = useRef<Terminal | null>(null);
        const [terminalSize, setTerminalSize] = useState({ cols: 0, rows: 0 });
        const fitRafIdRef = useRef<number | null>(null);
        const lastSentSizeRef = useRef<{ cols: number; rows: number } | null>(null);

        /**
         * 向后端发送终端窗口大小变更消息，让 PTY 的列数/行数与前端一致
         */
        const sendResizeMessage = (cols: number, rows: number, force?: boolean) => {
            const ws = wsRef.current;
            if (!ws || ws.readyState !== WebSocket.OPEN) {
                return;
            }
            if (!force) {
                const last = lastSentSizeRef.current;
                if (last && last.cols === cols && last.rows === rows) {
                    return;
                }
            }
            const size = JSON.stringify({ cols, rows });
            const payload = new TextEncoder().encode("\x01" + size);
            ws.send(payload);
            lastSentSizeRef.current = { cols, rows };
        };

        /**
         * 手动调整终端列数，并同步通知后端 PTY 变更窗口大小
         */
        const adjustCols = (delta: number) => {
            const term = termRef.current;
            if (!term) {
                return;
            }
            const rows = Math.max(1, term.rows || terminalSize.rows || 1);
            const newCols = Math.max(10, (term.cols || terminalSize.cols || 10) + delta);
            term.resize(newCols, rows);
            sendResizeMessage(newCols, rows);
        };


        useEffect(() => {

            const term = new Terminal({
                screenReaderMode: true,
                cursorBlink: true,
                allowProposedApi: true,
                fontFamily: 'Menlo, Monaco, "Courier New", monospace',

            });

            termRef.current = term;

            if (terminalRef.current) {
                term.open(terminalRef.current);
            } else {
                console.error("terminalRef.current is undefined");
                return;
            }


            let finalUrl = url;
            if (!finalUrl.startsWith("ws")) {
                const protocol = window.location.protocol === "https:" ? "wss://" : "ws://";
                finalUrl = protocol + location.host + finalUrl;
            }
            const ws = new WebSocket(finalUrl);

            wsRef.current = ws;
            const attachAddon = new AttachAddon(ws);
            const fitAddon = new FitAddon();
            const webLinksAddon = new WebLinksAddon();
            const unicode11Addon = new Unicode11Addon();
            const serializeAddon = new SerializeAddon();
            const webglAddon = new WebglAddon();
            const searchAddon = new SearchAddon();
            const clipboardAddon = new ClipboardAddon();
            term.loadAddon(attachAddon);
            term.loadAddon(fitAddon);
            term.loadAddon(webLinksAddon);
            term.loadAddon(unicode11Addon);
            term.loadAddon(serializeAddon);
            term.loadAddon(webglAddon);
            term.loadAddon(searchAddon);
            term.loadAddon(clipboardAddon);

            fitAddonRef.current = fitAddon;

            /**
             * 根据容器实际显示尺寸拟合终端列行，避免出现可输入宽度偏小的问题
             */
            const fitTerminal = () => {
                if (fitRafIdRef.current != null) {
                    cancelAnimationFrame(fitRafIdRef.current);
                }
                fitRafIdRef.current = requestAnimationFrame(() => {
                    const containerEl = terminalRef.current;
                    if (!containerEl) {
                        return;
                    }

                    const termEl = term.element;
                    const viewportEl = termEl?.querySelector('.xterm-viewport') as HTMLElement | null;
                    const availableWidth = viewportEl?.clientWidth ?? containerEl.clientWidth;
                    const availableHeight = containerEl.clientHeight;

                    const core = (term as any)?._core;
                    const cell = core?._renderService?.dimensions?.css?.cell;
                    const cellWidth = typeof cell?.width === 'number' && cell.width > 0 ? cell.width : null;
                    const cellHeight = typeof cell?.height === 'number' && cell.height > 0 ? cell.height : null;

                    if (cellWidth != null && cellHeight != null) {
                        const cols = Math.max(2, Math.floor(availableWidth / cellWidth));
                        const rows = Math.max(1, Math.floor(availableHeight / cellHeight));

                        if (cols !== term.cols || rows !== term.rows) {
                            term.resize(cols, rows);
                        }
                        return;
                    }

                    const proposed = fitAddon.proposeDimensions();
                    if (!proposed) {
                        fitAddon.fit();
                        return;
                    }

                    if (proposed.cols !== term.cols || proposed.rows !== term.rows) {
                        term.resize(proposed.cols, proposed.rows);
                        return;
                    }

                    fitAddon.fit();
                });
            };

            const resizeObserver = new ResizeObserver(() => {
                fitTerminal();
            });

            if (terminalRef.current) {
                resizeObserver.observe(terminalRef.current);
            }

            ws.onopen = () => {
                term.focus();
                setTimeout(fitTerminal, 50);
                setTimeout(fitTerminal, 200);
                setTimeout(fitTerminal, 500);
                setTimeout(() => sendResizeMessage(term.cols, term.rows, true), 80);
                setTimeout(() => sendResizeMessage(term.cols, term.rows, true), 220);
                setTimeout(() => sendResizeMessage(term.cols, term.rows, true), 520);
                document.fonts?.ready.then(() => {
                    fitTerminal();
                    sendResizeMessage(term.cols, term.rows, true);
                }).catch(() => {
                    fitTerminal();
                    sendResizeMessage(term.cols, term.rows, true);
                });
            };
            ws.onmessage = (event) => term.write(event.data);
            ws.onclose = () => term.write("\x1b[31mDisconnected\x1b[0m\r\n");
            ws.onerror = () => term.write("\x1b[31mError\x1b[0m\r\n");
            term.onResize(({ cols, rows }) => {
                setTerminalSize({ cols, rows });
                sendResizeMessage(cols, rows);
            });

            return () => {
                if (fitRafIdRef.current != null) {
                    cancelAnimationFrame(fitRafIdRef.current);
                    fitRafIdRef.current = null;
                }
                resizeObserver.disconnect();
                if (wsRef.current) {
                    wsRef.current.close();
                    wsRef.current = null;
                }
                term.dispose();
            };
        }, [url]);


        return (
            <div>
                <div style={{
                    padding: '8px 12px',
                    backgroundColor: '#f5f5f5',
                    borderBottom: '1px solid #ddd',
                    fontSize: '12px',
                    fontFamily: 'monospace',
                    color: '#666',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '12px',
                    flexWrap: 'wrap'
                }}>
                    <div>终端大小: {terminalSize.cols} 列 x {terminalSize.rows} 行</div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                        <button
                            onClick={() => adjustCols(-5)}
                            style={{
                                padding: '2px 8px',
                                fontSize: '12px',
                                cursor: 'pointer',
                                backgroundColor: '#fff',
                                border: '1px solid #ccc',
                                borderRadius: '3px'
                            }}
                        >
                            -5
                        </button>
                        <button
                            onClick={() => adjustCols(-1)}
                            style={{
                                padding: '2px 8px',
                                fontSize: '12px',
                                cursor: 'pointer',
                                backgroundColor: '#fff',
                                border: '1px solid #ccc',
                                borderRadius: '3px'
                            }}
                        >
                            -1
                        </button>
                        <button
                            onClick={() => adjustCols(1)}
                            style={{
                                padding: '2px 8px',
                                fontSize: '12px',
                                cursor: 'pointer',
                                backgroundColor: '#fff',
                                border: '1px solid #ccc',
                                borderRadius: '3px'
                            }}
                        >
                            +1
                        </button>
                        <button
                            onClick={() => adjustCols(5)}
                            style={{
                                padding: '2px 8px',
                                fontSize: '12px',
                                cursor: 'pointer',
                                backgroundColor: '#fff',
                                border: '1px solid #ccc',
                                borderRadius: '3px'
                            }}
                        >
                            +5
                        </button>
                    </div>
                </div>
                <div ref={terminalRef} style={{ width: width ? width : "100%", height: height ? height : "80vh" }}></div>
            </div>
        );
    }
);

export default XTermComponent;
