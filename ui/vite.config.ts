import {defineConfig} from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'
import monacoEditorPlugin from 'vite-plugin-monaco-editor';
import {copy} from 'fs-extra'
import * as http from 'http'
import * as https from 'https'

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
            // 添加代理配置：仅保留 monaco-editor 静态文件代理，其余交由自定义负载均衡中间件处理
            proxy: {
                '/monacoeditorwork': {
                    target: 'http://localhost:3000',
                    rewrite: (path) => path.replace(/^\/monacoeditorwork/, '/node_modules/monaco-editor/min/vs'),
                },
                // 保留 WebSocket 代理，避免影响现有功能（暂不在此处做负载均衡）
                '/ai/chat': {
                    target: 'ws://127.0.0.1:3618',
                    ws: true,
                    changeOrigin: true,
                },
                '^/k8s/cluster/[^/]+/pod/xterm': {
                    target: 'ws://127.0.0.1:3618',
                    ws: true,
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
             * 负载均衡代理插件（轮询转发）
             * 在多个后端服务之间进行简单的轮询负载均衡，支持 HTTP 与 WebSocket
             * 注意：不要将目标指向当前 Vite 开发端口（如 3000），避免形成自我代理循环
             */
            {
                name: 'lb-proxy',
                configureServer(server: any) {
                    // 后端目标列表：根据你的后端服务端口修改（示例为 3618 与 3619）
                    let httpTargets: string[] = ['http://127.0.0.1:3618', 'http://127.0.0.1:3619']
                    let rr = 0

                    /**
                     * 检测指定后端是否可用（通过 HEAD /，带超时）
                     * 返回 true 表示可用，false 表示不可用
                     */
                    function checkTarget(target: string, timeoutMs = 1000): Promise<boolean> {
                        return new Promise((resolve) => {
                            try {
                                const targetUrl = new URL(target)
                                const isHttps = targetUrl.protocol === 'https:'
                                const client = isHttps ? https : http
                                const options: http.RequestOptions = {
                                    hostname: targetUrl.hostname,
                                    port: targetUrl.port,
                                    method: 'HEAD',
                                    path: '/',
                                }
                                const req = client.request(options, (res) => {
                                    // 任意响应都视为端口可达（更细颗粒的健康检查可改为判断状态码或特定路径）
                                    resolve(true)
                                })
                                req.on('error', () => resolve(false))
                                req.setTimeout(timeoutMs, () => {
                                    req.destroy()
                                    resolve(false)
                                })
                                req.end()
                            } catch (e) {
                                resolve(false)
                            }
                        })
                    }

                    /**
                     * 探测所有目标，返回可用目标列表
                     */
                    async function probeTargets(targets: string[]): Promise<string[]> {
                        const results = await Promise.all(targets.map((t) => checkTarget(t)))
                        return targets.filter((t, i) => results[i])
                    }

                    // 启动时进行一次探测，过滤不可用端口
                    probeTargets(httpTargets).then((available) => {
                        if (available.length > 0) {
                            console.log(`[lb-proxy] 可用后端列表: ${available.join(', ')}`)
                            httpTargets = available
                        } else {
                            console.log(`[lb-proxy] 未探测到可用后端，保留原始列表: ${httpTargets.join(', ')}`)
                        }
                    })

                    /**
                     * forwardRequest 将请求转发到目标后端
                     * 使用 Node 内置 http/https 实现，避免额外依赖
                     */
                    function forwardRequest(req: any, res: any, target: string) {
                        const targetUrl = new URL(target)
                        const isHttps = targetUrl.protocol === 'https:'
                        const client = isHttps ? https : http
                        const options: http.RequestOptions = {
                            hostname: targetUrl.hostname,
                            port: targetUrl.port,
                            method: req.method,
                            path: req.url,
                            headers: req.headers,
                        }
                        const proxyReq = client.request(options, (proxyRes: http.IncomingMessage) => {
                            res.writeHead(proxyRes.statusCode || 502, proxyRes.headers as any)
                            proxyRes.pipe(res, { end: true })
                        })
                        proxyReq.on('error', (error: unknown) => {
                            console.log(`[lb-proxy] 代理错误: ${error instanceof Error ? error.message : String(error)}`)
                            res.statusCode = 502
                            res.end('代理错误')
                        })
                        req.pipe(proxyReq, { end: true })
                    }

                    /**
                     * HTTP 负载均衡中间件
                     * 将匹配到的业务接口请求轮询转发到后端
                     */
                    server.middlewares.use((req: any, res: any, next: any) => {
                        const url = req.url || ''
                        // 需要做负载均衡的路径前缀（根据现有代理统一处理）
                        const match = (
                            url.startsWith('/auth') ||
                            url.startsWith('/swagger') ||
                            url.startsWith('/k8s') ||
                            url.startsWith('/params') ||
                            url.startsWith('/mgm') ||
                            url.startsWith('/admin')
                        )
                        if (!match) return next()

                        // 修复路径中的 %2F%2F -> // 与原有逻辑一致
                        req.url = url.replace('%2F%2F', '//')

                        // 若当前无可用后端，交由下一中间件处理（避免除零错误）
                        if (!httpTargets.length) return next()
                        // 选择后端目标（轮询）
                        const target = httpTargets[rr++ % httpTargets.length]
                        console.log(`[lb-proxy] 代理到后端: ${target}  请求: ${req.method} ${req.url}`)
                        forwardRequest(req, res, target)
                    })
                }
            },
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
                                res.writeHead(200, {'Content-Type': 'application/json'});
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
