import { message, Modal } from 'antd';
import { fetcher } from '@/components/Amis/fetcher.ts';
import { FileNode } from './FileTree';
import { ProcessK8sUrlWithCluster } from '@/utils/utils';

interface FileOperationsProps {
    selectedContainer: string;
    podName: string;
    namespace: string;
}

export class FileOperations {
    private props: FileOperationsProps;

    constructor(props: FileOperationsProps) {
        this.props = props;
    }

    async handleCopy(node: FileNode) {
        await navigator.clipboard.writeText(node.path);
        message.success('路径已复制到剪贴板');
    }

    async handleRefresh(node: FileNode, callback: (children: FileNode[]) => void) {
        const children = await this.fetchData(node.path, node.isDir);
        if (node.isDir) {
            callback(children);
        }
        message.success('刷新成功');
    }

    async handleDelete(node: FileNode, onSuccess: () => void) {
        Modal.confirm({
            title: '请确认',
            content: `是否确认删除文件：${node.path} ？`,
            okText: '删除',
            cancelText: '取消',
            onOk: async () => {
                const response = await fetcher({
                    url: '/k8s/file/delete',
                    method: 'post',
                    data: {
                        containerName: this.props.selectedContainer,
                        podName: this.props.podName,
                        namespace: this.props.namespace,
                        path: node.path
                    }
                });
                message.success(response.data?.msg);
                if (response.data?.status === 0) {
                    onSuccess();
                }
            }
        });
    }

    async handleEditFile(node: FileNode, onEditorLoad: (content: string, language: string) => void) {
        if (node.type !== 'file') {
            message.error('只能编辑文件类型');
            return;
        }

        try {
            const response = await fetcher({
                url: '/k8s/file/show',
                method: 'post',
                data: {
                    containerName: this.props.selectedContainer,
                    podName: this.props.podName,
                    namespace: this.props.namespace,
                    path: node.path
                }
            });

            if (response.data?.status !== 0) {
                message.error(response.data?.msg || '非文本文件，不可在线编辑。请下载编辑后上传。');
                return;
            }

            //@ts-ignore
            const fileContent = response.data?.data?.content || '';
            let language = node.path?.split('.').pop() || 'plaintext';

            switch (language) {
                case 'yaml':
                case 'yml':
                    language = 'yaml';
                    break;
                case 'json':
                    language = 'json';
                    break;
                case 'py':
                    language = 'python';
                    break;
                default:
                    language = 'shell';
                    break;
            }

            onEditorLoad(fileContent, language);
        } catch (error) {
            message.error('获取文件内容失败');
        }
    }

    async downloadFile(node: FileNode, type?: string) {
        if (!node || !this.props.selectedContainer || !this.props.podName || !this.props.namespace) {
            message.error('缺少必要的参数，请检查输入');
            return;
        }

        try {
            const queryParams = new URLSearchParams({
                containerName: this.props.selectedContainer,
                podName: this.props.podName,
                namespace: this.props.namespace,
                path: node.path || "",
                token: localStorage.getItem('token') || "",
                type: type || ""
            }).toString();

            let url = `/k8s/file/download?${queryParams}`;
            url = ProcessK8sUrlWithCluster(url);
            const a = document.createElement('a');
            a.href = url;
            a.click();
            message.success('文件正在下载...');
        } catch (e) {
            message.error('下载失败，请重试');
        }
    }

    async handleUpload(node: FileNode, onUploadSuccess: () => void) {
        if (!node.isDir) {
            message.error('只能在目录下上传文件');
            return;
        }

        const uploadInput = document.createElement('input');
        uploadInput.type = 'file';
        uploadInput.onchange = async (e) => {
            const file = (e.target as HTMLInputElement).files?.[0];
            if (!file) return;

            const formData = new FormData();
            formData.append('file', file);
            formData.append('containerName', this.props.selectedContainer);
            formData.append('podName', this.props.podName);
            formData.append('namespace', this.props.namespace);
            formData.append('isDir', String(node.isDir));
            formData.append('path', String(node.path));
            formData.append('fileName', file.name);

            try {
                const url = ProcessK8sUrlWithCluster('/k8s/file/upload');
                const response = await fetch(url, {
                    method: 'POST',
                    headers: {
                        'Authorization': `Bearer ${localStorage.getItem('token')}`
                    },
                    body: formData
                });
                const result = await response.json();
                if (result.data?.file?.status === 'done') {
                    message.success('上传成功');
                    onUploadSuccess();
                } else {
                    message.error(result.data?.file?.error || '上传失败');
                }
            } catch (error) {
                message.error('上传失败');
            }
        };
        uploadInput.click();
    }

    async fetchData(path: string = '/', isDir: boolean): Promise<FileNode[]> {
        try {
            const response = await fetcher({
                url: `/k8s/file/list?path=${encodeURIComponent(path)}`,
                method: 'post',
                data: {
                    containerName: this.props.selectedContainer,
                    podName: this.props.podName,
                    namespace: this.props.namespace,
                    isDir: isDir,
                    path: path
                }
            });

            //@ts-ignore
            const rows = response.data?.data?.rows || [];
            return rows.map((item: any): FileNode => ({
                name: item.name || '',
                type: item.type || '',
                permissions: item.permissions || '',
                owner: item.owner || '',
                group: item.group || '',
                size: item.size || 0,
                modTime: item.modTime || '',
                path: item.path || '',
                isDir: item.isDir || false,
                isLeaf: !item.isDir,
                title: item.name,
                key: Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15),
            }));
        } catch (error) {
            console.error('Failed to fetch file tree:', error);
            return [];
        }
    }
}