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
        build: {
          rollupOptions: {
            output: {
              manualChunks(id) {
                if (id.includes('node_modules')) {
                  if (id.includes('react') || id.includes('amis')) {
                    return 'vendor';
                  }
                  if (id.includes('monaco-editor')) {
                    return 'editor';
                  }
                  if (id.includes('tinymce')) {
                    return 'rich-text';
                  }
                  if (id.includes('echarts')) {
                    return 'charts';
                  }
                  if (id.includes('exceljs') || id.includes('xlsx')) {
                    return 'excel';
                  }
                  if (id.includes('pdf')) {
                    return 'pdf';
                  }
                  if (id.includes('mpegts') || id.includes('hls')) {
                    return 'video';
                  }
                  return 'deps';
                }
              }
            }
          },
          chunkSizeWarningLimit: 1000
        },
        plugins: [react(), monacoEditorPlugin({})],
    }
})
