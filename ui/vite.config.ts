import {defineConfig} from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'
import monacoEditorPlugin from 'vite-plugin-monaco-editor';
import {copy} from 'fs-extra'

export default defineConfig(({mode}) => {
    console.log('current mode', mode)

    return {
        base: '/',
        server: {
            port: 3000,
            open: true,
            host: '0.0.0.0',
            // 添加静态文件服务
            fs: {
                allow: ['..'],
            },
            // 添加代理配置
            // 添加monaco-editor静态文件代理
            proxy: {
                '/monacoeditorwork': {
                    target: 'http://localhost:3000',
                    rewrite: (path) => path.replace(/^\/monacoeditorwork/, '/node_modules/monaco-editor/min/vs'),
                },
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
                }, '/swagger': {
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
                '/params': {
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
                '/admin': {
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
                '/ai/chat': {
                    target: 'ws://127.0.0.1:3618', // 替换为实际的目标地址
                    ws: true, // 开启 WebSocket 代理
                    changeOrigin: true,
                },

                '^/k8s/cluster/[^/]+/pod/xterm': {
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
        plugins: [react(),
            monacoEditorPlugin({
                publicPath: '/monacoeditorwork', // 静态资源输出路径
            }),
            {
                name: 'copy-monaco-loader',
                closeBundle() {
                    copy('node_modules/monaco-editor/min/vs/loader.js', 'dist/monacoeditorwork/loader.js', {overwrite: true})
                    copy('node_modules/monaco-editor/min/vs/editor', 'dist/monacoeditorwork/editor', {overwrite: true})
                    copy('node_modules/monaco-editor/min/vs/language', 'dist/monacoeditorwork/language', {overwrite: true})
                    copy('node_modules/monaco-editor/min/vs/base', 'dist/monacoeditorwork/base', {overwrite: true})
                    copy('node_modules/monaco-editor/min/vs/basic-languages', 'dist/monacoeditorwork/basic-languages', {overwrite: true})
                }
            },
            {
                name: 'favicon',
                closeBundle() {
                    copy('src/assets/favicon.ico', 'dist/favicon.ico', {overwrite: true})
                }
            }
        ],


    }
})
