import {defineConfig} from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'
import monacoEditorPlugin from 'vite-plugin-monaco-editor';
import { copy } from 'fs-extra'
 
export default defineConfig(({mode}) => {
    console.log('current mode', mode)

    return {
        build: {
            rollupOptions: {
                output: {
                    manualChunks: undefined
                }
            }
        },

        base: '/',
        server: {
            port: 3000,
            open: true,
            host: '0.0.0.0',
            // 添加代理配置
            proxy: {
                '/auth': {
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
                '/mgm': {
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

                '/k8s/pod/xterm': {
                    target: 'ws://127.0.0.1:3618', // 替换为实际的目标地址
                    ws: true, // 开启 WebSocket 代理
                    changeOrigin: true,
                },
                '/k8s/node/xterm': {
                    target: 'ws://127.0.0.1:3618', // 替换为实际的目标地址
                    ws: true, // 开启 WebSocket 代理
                    changeOrigin: true,
                }
            },
        },
        resolve: {
            alias: Object.assign(
                {
                    '@': path.resolve(__dirname, 'src'),
                }
            ),
        },
        plugins: [react(),
            monacoEditorPlugin({
                publicPath: '/monacoeditorwork', // 静态资源输出路径
            }),
            {
                name: 'copy-monaco-loader',
                closeBundle() {
                        copy('node_modules/monaco-editor/min/vs/loader.js', 'dist/monacoeditorwork/loader.js', { overwrite: true })
                        copy('node_modules/monaco-editor/min/vs/editor', 'dist/monacoeditorwork/editor', { overwrite: true })
                        copy('node_modules/monaco-editor/min/vs/language', 'dist/monacoeditorwork/language', { overwrite: true })
                        copy('node_modules/monaco-editor/min/vs/base', 'dist/monacoeditorwork/base', { overwrite: true })
                        copy('node_modules/monaco-editor/min/vs/basic-languages', 'dist/monacoeditorwork/basic-languages', { overwrite: true })
                    }
            }
        ],
        

    }
})
