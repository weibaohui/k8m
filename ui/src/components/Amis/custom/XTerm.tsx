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

        const adjustCols = (delta: number) => {
            if (termRef.current) {
                const newCols = Math.max(10, terminalSize.cols + delta);
                termRef.current.resize(newCols, terminalSize.rows);
            }
        };


        useEffect(() => {

            const term = new Terminal({
                screenReaderMode: true,
                cursorBlink: true,
                allowProposedApi: true,

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


            const fitTerminal = () => {
                requestAnimationFrame(() => {
                    fitAddon.fit();
                    const dimensions = term;
                    setTerminalSize({ cols: dimensions.cols, rows: dimensions.rows });
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
            };
            ws.onmessage = (event) => term.write(event.data);
            ws.onclose = () => term.write("\x1b[31mDisconnected\x1b[0m\r\n");
            ws.onerror = () => term.write("\x1b[31mError\x1b[0m\r\n");
            term.onResize(({ cols, rows }) => {
                const size = JSON.stringify({ cols, rows: rows + 1 });
                const send = new TextEncoder().encode("\x01" + size);
                ws.send(send);
            });

            return () => {
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


