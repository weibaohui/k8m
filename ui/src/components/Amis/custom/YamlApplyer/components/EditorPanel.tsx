import React, { useEffect, useRef } from 'react';
import * as monaco from 'monaco-editor';
import { Button, Modal, List } from 'antd';
import { fetcher } from "@/components/Amis/fetcher.ts";

interface EditorPanelProps {
    onSaveSuccess: (content: string) => void;
    initialContent?: string;
}

const EditorPanel: React.FC<EditorPanelProps> = ({ onSaveSuccess, initialContent = '' }) => {
    const monacoInstance = useRef<monaco.editor.IStandaloneCodeEditor | null>(null);
    const editorRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        if (editorRef.current) {
            monacoInstance.current = monaco.editor.create(editorRef.current, {
                theme: 'vs',
                language: "yaml",
                wordWrap: "on",
                scrollbar: {
                    vertical: "auto"
                },
                automaticLayout: true,
                minimap: {
                    enabled: false
                },
                value: initialContent
            });
        }
        return () => monacoInstance.current?.dispose();
    }, []);

    useEffect(() => {
        if (monacoInstance.current && initialContent !== monacoInstance.current.getValue()) {
            monacoInstance.current.setValue(initialContent);
        }
    }, [initialContent]);

    const handleDelete = async () => {
        const content = monacoInstance.current?.getValue();
        if (!content) return;

        try {
            const response = await fetcher({
                url: '/k8s/yaml/delete',
                method: 'post',
                data: {
                    yaml: content
                }
            });
            const responseData = response.data;
            if (responseData?.status !== 0) {
                Modal.error({
                    title: '删除失败',
                    content: `操作失败：${response.data?.msg}`
                });
                return;
            }

            //@ts-ignore
            const resultList = responseData.data.result || [];
            Modal.success({
                title: '删除状态',
                content: (
                    <List
                        style={{ maxHeight: '400px', overflow: 'auto' }}
                        dataSource={resultList}
                        renderItem={(item, index) => {
                            const resultItem = item as string;
                            return (
                                <List.Item key={index} style={{ padding: '8px' }}>
                                    <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                                        <div>{resultItem}</div>
                                    </div>
                                </List.Item>
                            );
                        }}
                    />
                )
            });

            // 删除后保持编辑器内容不变
        } catch (error) {
            Modal.error({
                title: '删除失败',
                content: error instanceof Error ? error.message : '未知错误'
            });
        }
    };

    const handleSave = async () => {
        const content = monacoInstance.current?.getValue();
        if (!content) return;

        try {
            const response = await fetcher({
                url: '/k8s/yaml/apply',
                method: 'post',
                data: {
                    yaml: content
                }
            });
            const responseData = response.data;
            if (responseData?.status !== 0) {
                if (response.data?.msg.includes("please apply your changes to the latest version and try again")) {
                    Modal.error({
                        title: '应用失败',
                        content: '资源已被更新，请刷新后再试。'
                    });
                    return;
                } else {
                    Modal.error({
                        title: '应用失败',
                        content: `请尝试刷新后重试。${response.data?.msg}`
                    });
                    return;
                }
            }

            //@ts-ignore
            const resultList = responseData.data.result || [];
            Modal.success({
                title: '应用状态',
                content: (
                    <List
                        style={{ maxHeight: '400px', overflow: 'auto' }}
                        dataSource={resultList}
                        renderItem={(item, index) => {
                            const resultItem = item as string;
                            return (
                                <List.Item key={index} style={{ padding: '8px' }}>
                                    <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                                        <div>{resultItem}</div>
                                    </div>
                                </List.Item>
                            );
                        }}
                    />
                )
            });

            onSaveSuccess(content);
        } catch (error) {
            Modal.error({
                title: '应用失败',
                content: error instanceof Error ? error.message : '未知错误'
            });
        }
    };

    const handleFileUpload = () => {
        const input = document.createElement('input');
        input.type = 'file';
        input.accept = '.yaml,.yml';
        input.onchange = async (e) => {
            const file = (e.target as HTMLInputElement).files?.[0];
            if (file) {
                try {
                    const reader = new FileReader();
                    reader.onload = (e) => {
                        const content = e.target?.result as string;
                        if (monacoInstance.current) {
                            monacoInstance.current.setValue(content);
                        }
                    };
                    reader.readAsText(file);
                } catch (error) {
                    Modal.error({
                        title: '导入失败',
                        content: '无法读取YAML文件'
                    });
                }
            }
        };
        input.click();
    };

    return (
        <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
            <div style={{ marginBottom: '10px' }}>
                <Button onClick={handleSave} type="primary" style={{ marginRight: '8px' }}>
                    应用
                </Button>
                <Button onClick={() => {
                    Modal.confirm({
                        title: '确认删除',
                        content: '确定要从集群中删除这些资源吗？此操作不可恢复。',
                        onOk: handleDelete
                    });
                }} danger style={{ marginRight: '8px' }}>
                    从集群删除
                </Button>
                <Button onClick={handleFileUpload}>
                    导入文件
                </Button>
            </div>
            <div ref={editorRef} style={{ flex: 1, border: '1px solid #d9d9d9' }} />
        </div>
    );
};

export default EditorPanel;