import {defineConfig} from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'
import monacoEditorPlugin from 'vite-plugin-monaco-editor';

export default defineConfig(({mode}) => {
    console.log('current mode', mode)
    return {
        base: '/',
        server: {
            port: 3000,
            open: true,
            host: '0.0.0.0',
            // 添加代理配置
            proxy: {
                '/k8s': {
                    target: 'http://127.0.0.1:3618',
                    changeOrigin: true,
                    configure: (proxy) => {
                        proxy.on('proxyReq', (proxyReq, req) => {
                            const originalPath = req.url;
                            console.log(`Before restoring: ${originalPath}`);
                            // @ts-expect-error
                            proxyReq.path = originalPath.replace('%2F%2F', '//');
                            console.log(`Restored path: ${proxyReq.path}`);
                        });
                    },
                },
                '/k8s/chat': {
                    target: 'ws://127.0.0.1:3618', // 替换为实际的目标地址
                    ws: true, // 开启 WebSocket 代理
                    changeOrigin: true,
                },
            },
        },
        resolve: {
            alias: Object.assign(
                {
                    '@': path.resolve(__dirname, 'src'),
                }
            ),
        },
        plugins: [react(), monacoEditorPlugin({})],
    }
})
