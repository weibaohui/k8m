import React, { useEffect, useRef } from 'react';
import { Terminal } from '@xterm/xterm';
import "@xterm/xterm/css/xterm.css";
import { FitAddon } from "@xterm/addon-fit";

interface XTermTestProps {
    width?: string;
    height?: string;
}

const XTermTestComponent = React.forwardRef<HTMLDivElement, XTermTestProps>((props, ref) => {
    const terminalRef = useRef<HTMLDivElement | null>(null);

    useEffect(() => {
        if (!terminalRef.current) return;

        // 初始化 Terminal
        const term = new Terminal({
            cursorBlink: true,
            fontFamily: 'Menlo, Monaco, "Courier New", monospace',
            fontSize: 14,
            theme: {
                background: '#1e1e1e',
            }
        });

        const fitAddon = new FitAddon();
        term.loadAddon(fitAddon);

        term.open(terminalRef.current);

        // 稍微延迟一下 fit，确保容器已渲染
        setTimeout(() => {
            fitAddon.fit();
            term.focus();
        }, 100);

        term.write('Welcome to XTerm Test Demo (Local Echo)\r\n');
        term.write('Try typing a long sentence to test line wrapping...\r\n');
        term.write('$ ');

        // 本地回显逻辑 (Local Echo)
        term.onData(e => {
            switch (e) {
                case '\r': // Enter
                    term.write('\r\n$ ');
                    break;
                case '\u007F': // Backspace (DEL)
                    // Note: This is a very basic backspace implementation
                    // Real implementation needs to track cursor position
                    term.write('\b \b');
                    break;
                default:
                    // 简单的回显
                    if (e >= String.fromCharCode(0x20) && e <= String.fromCharCode(0x7E) || e >= '\u00a0') {
                        term.write(e);
                    }
            }
        });

        // 监听窗口大小变化
        const resizeObserver = new ResizeObserver(() => {
            fitAddon.fit();
        });
        resizeObserver.observe(terminalRef.current);

        return () => {
            resizeObserver.disconnect();
            term.dispose();
        };
    }, []);

    return (
        <div
            ref={terminalRef}
            style={{
                width: props.width || '100%',
                height: props.height || '100%',
                overflow: 'hidden',
                backgroundColor: '#1e1e1e'
            }}
        />
    );
});

export default XTermTestComponent;
