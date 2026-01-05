import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'
import monacoEditorPlugin from 'vite-plugin-monaco-editor';
import { copy } from 'fs-extra'
import fs from 'fs'

export default defineConfig(({ mode }) => {
    console.log('current mode', mode)

    // 复制插件前端JSON到public目录
    const copyPluginFrontendsToPublic = async () => {
        const modulesDir = path.resolve(__dirname, '../pkg/plugins/modules');
        const destRoot = path.resolve(__dirname, 'public/pages/plugins');
        try {
            if (!fs.existsSync(modulesDir)) return;
            // 确保目标根目录存在
            fs.mkdirSync(destRoot, { recursive: true });
            const modules = fs.readdirSync(modulesDir, { withFileTypes: true })
                .filter(d => d.isDirectory());
            for (const m of modules) {
                const frontendDir = path.join(modulesDir, m.name, 'frontend');
                if (fs.existsSync(frontendDir)) {
                    const dest = path.join(destRoot, m.name);
                    await copy(frontendDir, dest, { overwrite: true });
                }
            }
            console.log('[插件前端] 已复制到 public/pages/plugins');
        } catch (e) {
            console.warn('[插件前端] 复制到 public 失败：', e);
        }
    };

    // 复制插件前端JSON到dist目录（构建时）
    const copyPluginFrontendsToDist = async () => {
        const modulesDir = path.resolve(__dirname, '../pkg/plugins/modules');
        const destRoot = path.resolve(__dirname, 'dist/pages/plugins');
        try {
            if (!fs.existsSync(modulesDir)) return;
            fs.mkdirSync(destRoot, { recursive: true });
            const modules = fs.readdirSync(modulesDir, { withFileTypes: true })
                .filter(d => d.isDirectory());
            for (const m of modules) {
                const frontendDir = path.join(modulesDir, m.name, 'frontend');
                if (fs.existsSync(frontendDir)) {
                    const dest = path.join(destRoot, m.name);
                    await copy(frontendDir, dest, { overwrite: true });
                }
            }
            console.log('[插件前端] 已复制到 dist/pages/plugins');
        } catch (e) {
            console.warn('[插件前端] 复制到 dist 失败：', e);
        }
    };

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
        /**
         * 自定义插件：处理 POST 请求并打印请求信息
         */
        {
            name: 'post-request-logger',
            configureServer(server: any) {
                server.middlewares.use((req: any, res: any, next: any) => {
                    // 只处理 POST 请求
                    if (req.method === 'POST' && req.url === '/echo') {
                        let body = '';

                        // 收集请求体数据
                        req.on('data', (chunk: any) => {
                            body += chunk.toString();
                        });

                        // 请求体接收完成后处理
                        req.on('end', () => {
                            console.log('=== POST 请求信息 ===');
                            console.log('URL:', req.url);
                            console.log('Content-Type:', req.headers['content-type'] || 'undefined');
                            console.log('Body:', body);
                            console.log('========================');

                            // 返回响应
                            res.writeHead(200, { 'Content-Type': 'application/json' });
                            res.end(JSON.stringify({
                                message: 'POST 请求已接收并打印',
                                url: req.url,
                                contentType: req.headers['content-type'] || 'undefined',
                                bodyLength: body.length
                            }));
                        });
                    } else {
                        // 非 POST 请求继续传递给下一个中间件
                        next();
                    }
                });
            }
        },
        /**
         * 插件：实时复制 Go 插件前端JSON到 public
         */
        {
            name: 'copy-plugin-frontends-dev',
            async configureServer(server: any) {
                // 启动时先复制一次
                await copyPluginFrontendsToPublic();
                // 监听插件frontend目录变化，变化则复制
                const watchRoot = path.resolve(__dirname, '../pkg/plugins/modules');
                server.watcher.add(watchRoot);
                server.watcher.on('add', (file: string) => {
                    if (file.includes('/frontend/')) {
                        copyPluginFrontendsToPublic();
                    }
                });
                server.watcher.on('change', (file: string) => {
                    if (file.includes('/frontend/')) {
                        copyPluginFrontendsToPublic();
                    }
                });
                server.watcher.on('unlink', (file: string) => {
                    if (file.includes('/frontend/')) {
                        copyPluginFrontendsToPublic();
                    }
                });
            }
        },
        {
            name: 'copy-monaco-loader',
            closeBundle() {
                copy('node_modules/monaco-editor/min/vs/loader.js', 'dist/monacoeditorwork/loader.js', { overwrite: true })
                copy('node_modules/monaco-editor/min/vs/editor', 'dist/monacoeditorwork/editor', { overwrite: true })
                copy('node_modules/monaco-editor/min/vs/language', 'dist/monacoeditorwork/language', { overwrite: true })
                copy('node_modules/monaco-editor/min/vs/base', 'dist/monacoeditorwork/base', { overwrite: true })
                copy('node_modules/monaco-editor/min/vs/basic-languages', 'dist/monacoeditorwork/basic-languages', { overwrite: true })
            }
        },
        // 构建结束时复制插件前端到 dist
        {
            name: 'copy-plugin-frontends-build',
            async closeBundle() {
                await copyPluginFrontendsToDist();
            }
        },
        {
            name: 'favicon',
            closeBundle() {
                copy('src/assets/favicon.ico', 'dist/favicon.ico', { overwrite: true })
            }
        }
        ],


    }
})
