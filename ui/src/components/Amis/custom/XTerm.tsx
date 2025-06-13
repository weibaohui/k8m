import React, {useEffect, useRef} from 'react';
import {Terminal} from '@xterm/xterm'
import "@xterm/xterm/css/xterm.css";
import {formatFinalGetUrl, ProcessK8sUrlWithCluster} from "@/utils/utils.ts";
import {AttachAddon} from "@xterm/addon-attach";
import {FitAddon} from "@xterm/addon-fit";
import {WebLinksAddon} from "@xterm/addon-web-links";
import {Unicode11Addon} from "@xterm/addon-unicode11";
import {SerializeAddon} from "@xterm/addon-serialize";
import {WebglAddon} from '@xterm/addon-webgl';
import {SearchAddon} from "@xterm/addon-search";
import {ClipboardAddon} from '@xterm/addon-clipboard';

interface XTermProps {
    url: string;
    params: Record<string, string>;
    data: Record<string, any>
    width?: string,
    height?: string
}

const XTermComponent = React.forwardRef<HTMLDivElement, XTermProps>(
    ({url, data, params, width, height}, _) => {
        url = formatFinalGetUrl({url, data, params});
        const token = localStorage.getItem('token');
        url = url + (url.includes('?') ? '&' : '?') + `token=${token}`;
        url = ProcessK8sUrlWithCluster(url)
        const wsRef = useRef<WebSocket | null>(null);
        const terminalRef = useRef<HTMLDivElement | null>(null);
        const fitAddonRef = useRef<FitAddon | null>(null);


        useEffect(() => {

            const term = new Terminal({
                screenReaderMode: true,
                cursorBlink: true,
                cols: 128,
                rows: 30,
                allowProposedApi: true,  // 启用提议的 API

            });

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
            // 添加插件
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


            // 连接事件
            ws.onopen = () => {
                term.focus();
                // 延迟100ms执行fit以确保终端完全初始化
                setTimeout(() => fitAddon.fit(), 100);
                setTimeout(() => fitAddon.fit(), 1000);
                // 连接成功时不显示Connected消息
                // term.write("\x1b[32mConnected\x1b[0m\r\n");
            };
            ws.onmessage = (event) => term.write(event.data);
            ws.onclose = () => term.write("\x1b[31mDisconnected\x1b[0m\r\n");
            ws.onerror = () => term.write("\x1b[31mError\x1b[0m\r\n");
            // 监听终端大小调整
            term.onResize(({cols, rows}) => {
                const size = JSON.stringify({cols, rows: rows + 1});
                const send = new TextEncoder().encode("\x01" + size);
                ws.send(send);
            });

            // 窗口大小变更时适配终端
            const handleResize = () => {
                fitAddon.fit();
            };
            window.addEventListener("resize", handleResize);
            return () => {
                if (wsRef.current) {
                    wsRef.current.close();
                    wsRef.current = null;
                }
                term.dispose();
            };
        }, [url]);


        return (
            <div ref={terminalRef} style={{width: width ? width : "100%", height: height ? height : "80vh"}}></div>
        );
    }
);

export default XTermComponent;


