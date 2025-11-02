import {create} from 'zustand'
import {Schema} from 'amis'
import ajax from '@/utils/ajax'
import {getCurrentClusterId, setCurrentClusterId} from '@/utils/utils'

interface Store {
    schema: Schema
    loading: boolean
    initPage: (path: string) => void
}

const useStore = create<Store>((set) => ({
    schema: {
        type: 'page'
    },
    loading: false,
    initPage(path) {
        set({loading: true})

        if (path.startsWith('/crd/namespaced_cr') || path.startsWith('/crd/cluster_cr')) {
            try {
                const hash = window.location.hash || '';
                const queryString = hash.includes('?') ? hash.split('?')[1] : '';
                const urlParams = new URLSearchParams(queryString);
                const group = urlParams.get('group') || '';
                const kind = urlParams.get('kind') || '';
                const version = urlParams.get('version') || '';
                const scope = urlParams.get('scope') || (path.startsWith('/crd/namespaced_cr') ? 'Namespaced' : 'Cluster');
                const cols = urlParams.get('cols') || '';
                const mode = urlParams.get('mode') || 'append';
                const ns = urlParams.get('ns') || '';
                const linkCluster = urlParams.get('cluster') || '';

                // 如果链接中携带 cluster，则切换并刷新
                if (linkCluster) {
                    const originCluster = getCurrentClusterId();
                    if (originCluster !== linkCluster) {
                        setCurrentClusterId(linkCluster);
                        // 刷新后会重新构建页面
                        window.location.reload();
                        return;
                    }
                }
                const defaultNs = ns || (localStorage.getItem('selectedNs') || (scope === 'Namespaced' ? 'default' : ''));

                if (!group || !kind || !version) {
                    set({
                        schema: {
                            type: 'page',
                            body: [{
                                type: 'alert',
                                level: 'danger',
                                body: '缺少必要参数：group、kind、version'
                            }]
                        }
                    });
                    return;
                }

                // 构建本地schema
                const pageSchema: Schema = buildDynamicCRDSchema({
                    group, kind, version, scope, ns: defaultNs, cols, mode
                }) as unknown as Schema;

                set({ schema: pageSchema });
            } catch (error: any) {
                console.error('Failed to build CRD page schema:', error);
                set({
                    schema: {
                        type: 'page',
                        body: [{
                            type: 'alert',
                            level: 'danger',
                            body: `构建CRD页面失败: ${error?.message || '未知错误'}`
                        }]
                    }
                });
            } finally {
                set({loading: false})
            }
        } else {
            let page = path.slice(1);
            page = page + '.json';
            const url = `/public/pages/${page}`;
            ajax.get(url).then(res => {
                set({
                    schema: res.data
                })
            }).finally(() => {
                set({loading: false})
            })
        }
    }
}))

export default useStore
function safeParseJSON<T = any>(input: string): T | null {
    try {
        return JSON.parse(input) as T;
    } catch {
        return null;
    }
}

function tryDecodeBase64JSON<T = any>(input: string): T | null {
    try {
        const normalized = input.replace(/\s/g, '');
        const base64 = normalized.replace(/-/g, '+').replace(/_/g, '/');
        const padLen = (4 - (base64.length % 4)) % 4;
        const padded = base64 + '='.repeat(padLen);
        const decoded = atob(padded);
        return JSON.parse(decoded) as T;
    } catch {
        return null;
    }
}

function parseCustomColumns(cols: string | null | undefined): any[] {
    if (!cols) return [];
    // 先尝试直接解析 JSON
    const asJson = safeParseJSON<any[]>(cols);
    if (Array.isArray(asJson)) return asJson;
    // 尝试 Base64 解码后解析
    const asB64 = tryDecodeBase64JSON<any[]>(cols);
    if (Array.isArray(asB64)) return asB64;
    // 兜底：返回空
    return [];
}

function buildDynamicCRDSchema(props: {
    group: string;
    kind: string;
    version: string;
    scope: string; // Namespaced | Cluster
    ns: string; // 逗号分隔可多选
    cols?: string;
    mode?: string; // append | replace
}): any {
    const { group, kind, version, scope, ns, cols, mode } = props;
    const customColumns = parseCustomColumns(cols);
    const replaceMode = mode === 'replace';

    const defaultColumns: any[] = [];
    // 操作列始终放在最前
    const operationColumn: any = {
        type: 'operation',
        label: '操作',
        width: 120,
        buttons: [
            {
                type: 'button',
                icon: 'fas fa-eye text-primary',
                actionType: 'drawer',
                tooltip: '资源描述',
                drawer: {
                    closeOnEsc: true,
                    closeOnOutside: true,
                    size: 'xl',
                    title: 'Describe: ${metadata.name}  (ESC 关闭)',
                    body: [
                        {
                            type: 'service',
                            api: `post:/k8s/${kind}/group/${group}/version/${version}/describe/ns/$metadata.namespace/name/$metadata.name`,
                            body: [
                                {
                                    type: 'highlightHtml',
                                    keywords: ['Error', 'Warning'],
                                    html: '${result}'
                                }
                            ]
                        }
                    ]
                }
            },
            {
                type: 'button',
                icon: 'fa fa-edit text-primary',
                tooltip: 'Yaml编辑',
                actionType: 'drawer',
                drawer: {
                    closeOnEsc: true,
                    closeOnOutside: true,
                    size: 'lg',
                    title: 'Yaml管理',
                    body: [
                        {
                            type: 'tabs',
                            tabsMode: 'tiled',
                            tabs: [
                                {
                                    title: '查看编辑',
                                    body: [
                                        {
                                            type: 'service',
                                            api: `get:/k8s/${kind}/group/${group}/version/${version}/ns/$metadata.namespace/name/$metadata.name`,
                                            body: [
                                                {
                                                    type: 'mEditor',
                                                    text: '${yaml}',
                                                    componentId: 'yaml',
                                                    saveApi: `/k8s/${kind}/group/${group}/version/${version}/update/ns/${'${metadata.namespace}'}/name/${'${metadata.name}'}`,
                                                    options: {
                                                        language: 'yaml',
                                                        wordWrap: 'on',
                                                        scrollbar: { vertical: 'auto' }
                                                    }
                                                }
                                            ]
                                        }
                                    ]
                                }
                            ]
                        }
                    ]
                }
            },
            {
                type: 'button',
                icon: 'fa fa-trash text-danger',
                label: '',
                confirmText: '确定要删除该资源?',
                actionType: 'ajax',
                api: `post:/k8s/${kind}/group/${group}/version/${version}/remove/ns/$metadata.namespace/name/$metadata.name`
            }
        ]
    };

    // 默认列
    defaultColumns.push({ name: 'metadata.name', label: '名称', type: 'text', sortable: true, width: '220px' });
    if (scope === 'Namespaced') {
        defaultColumns.push({ name: 'metadata.namespace', label: '命名空间', type: 'text', width: '160px' });
    }
    defaultColumns.push({ name: 'metadata.creationTimestamp', label: '存在时长', type: 'k8sAge' });

    const mergedColumns = replaceMode ? customColumns : [...defaultColumns, ...customColumns];

    // 批量动作
    const bulkActions = [
        {
            label: '批量删除',
            actionType: 'ajax',
            confirmText: '确定要批量删除?',
            api: {
                url: `/k8s/${kind}/group/${group}/version/${version}/batch/remove`,
                method: 'post',
                data: {
                    name_list: "${selectedItems | pick:metadata.name }",
                    ns_list: "${selectedItems | pick:metadata.namespace }"
                }
            }
        },
        {
            label: '强制删除',
            actionType: 'ajax',
            confirmText: '确定要批量强制删除?',
            api: {
                url: `/k8s/${kind}/group/${group}/version/${version}/force_remove`,
                method: 'post',
                data: {
                    name_list: "${selectedItems | pick:metadata.name }",
                    ns_list: "${selectedItems | pick:metadata.namespace }"
                }
            }
        }
    ];

    // API 构建，统一走 list/ns/$ns（后端已注册空 ns 的路由）
    const listUrl = `/k8s/${kind}/group/${group}/version/${version}/list/ns/$ns`;

    return {
        type: 'page',
        data: {
            kind,
            group,
            version,
            scope,
            // 初始 ns：URL 提供优先，否则沿用本地选择
            ns: ns || "${ls:selectedNs||''}"
        },
        body: [
            {
                type: 'container',
                className: 'floating-toolbar-right',
                body: [
                    {
                        type: 'wrapper',
                        style: { display: 'inline-flex' },
                        body: [
                            {
                                type: 'form',
                                mode: 'inline',
                                wrapWithPanel: false,
                                body: [
                                    {
                                        label: '集群',
                                        type: 'select',
                                        multiple: false,
                                        name: 'cluster',
                                        id: 'cluster',
                                        searchable: true,
                                        source: '/params/cluster/option_list',
                                        value: '${ls:cluster}',
                                        onEvent: {
                                            change: {
                                                actions: [
                                                    { actionType: 'custom', script: "window.setCurrentClusterId(event.data.value)" },
                                                    { actionType: 'custom', script: 'window.location.reload();' }
                                                ]
                                            }
                                        }
                                    },
                                    ...(scope === 'Namespaced' ? [
                                        {
                                            label: '命名空间',
                                            type: 'select',
                                            name: 'ns',
                                            searchable: true,
                                            clearable: true,
                                            multiple: true,
                                            maxTagCount: 1,
                                            source: '/k8s/ns/option_list',
                                            onEvent: {
                                                change: {
                                                    actions: [
                                                        { actionType: 'custom', script: "const v = Array.isArray(event.data.value)?event.data.value.join(','):String(event.data.value||''); doAction('listCRUD','setValue',{ns:v}); localStorage.setItem('selectedNs', v); doAction('listCRUD','reload');" }
                                                    ]
                                                }
                                            }
                                        }
                                    ] : [])
                                ]
                            }
                        ]
                    }
                ]
            },
            {
                type: 'crud',
                id: 'listCRUD',
                name: 'listCRUD',
                autoFillHeight: true,
                autoGenerateFilter: { columnsNum: 4, showBtnToolbar: false, defaultCollapsed: false },
                headerToolbar: [
                    { type: 'columns-toggler', align: 'right', draggable: true, icon: 'fas fa-cog', overlay: true, footerBtnSize: 'sm' },
                    { type: 'tpl', tpl: '共${count}条', align: 'right', visibleOn: '${count}' },
                    'reload',
                    'bulkActions'
                ],
                loadDataOnce: false,
                syncLocation: false,
                perPage: 10,
                bulkActions,
                api: { url: listUrl, method: 'post' },
                columns: [operationColumn, ...mergedColumns]
            }
        ]
    };
}
