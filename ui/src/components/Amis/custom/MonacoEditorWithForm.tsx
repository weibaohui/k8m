import React, { useEffect, useRef, useState } from 'react';
import * as monaco from 'monaco-editor';

import { replacePlaceholders } from "@/utils/utils.ts";
import { fetcher } from "@/components/Amis/fetcher.ts";
import { Button, Input, message } from 'antd';


//保存如需传递更多参数，请参考MonacoEditorWithFormProps
//saveApi={`/k8s/file/save`}
// data = {{
//     params: {
//         containerName: selectedContainer,
//         podName: podName,
//         namespace: namespace,
//         path: selected?.path || '',
// }}
//
interface MonacoEditorWithFormProps {
    text: string;
    saveApi: string;
    componentId: string;
    data: Record<string, any>
    options?: monaco.editor.IStandaloneEditorConstructionOptions;
}

const MonacoEditorWithForm: React.FC<MonacoEditorWithFormProps> = ({
    text,
    saveApi,
    data,
    options,
    componentId
}) => {
    const editorRef = useRef<HTMLDivElement>(null);
    const monacoInstance = useRef<monaco.editor.IStandaloneCodeEditor | null>(null);
    const [editorValue, setEditorValue] = useState(text);
    const [loading, setLoading] = useState(false);
    const [messageApi, contextHolder] = message.useMessage();

    text = replacePlaceholders(text, data)
    if (saveApi) {
        saveApi = replacePlaceholders(saveApi, data)
    }
    useEffect(() => {
        if (editorRef.current) {
            monacoInstance.current = monaco.editor.create(editorRef.current, {
                value: text,
                theme: 'vs',
                automaticLayout: true,
                minimap: {
                    enabled: false // 关闭小地图
                },
                ...options,
            });

            monacoInstance.current.onDidChangeModelContent(() => {
                setEditorValue(monacoInstance.current?.getValue() || '');
            });
        }
        return () => monacoInstance.current?.dispose();
    }, []);

    useEffect(() => {
        if (monacoInstance.current && text !== monacoInstance.current.getValue()) {
            monacoInstance.current.setValue(text);
        }
    }, [text]);


    const handleSave = async () => {
        if (!saveApi) return;
        setLoading(true);
        // 构造请求数据，将编辑器的值和额外参数合并
        const requestData = {
            [componentId]: editorValue,
            ...(data.params || {}) // 如果存在params属性，将其展开并合并到请求数据中
        };

        const response = await fetcher({
            url: saveApi,
            method: 'post',
            data: requestData
        });

        if (response.data?.status !== 0) {
            if (response.data?.msg.includes("please apply your changes to the latest version and try again")) {
                messageApi.error("保存失败: 资源已被更新，请刷新后再试。");
            } else {
                messageApi.error(`保存失败:请尝试刷新后重试。 ${response.data?.msg}`);
            }
        } else {
            messageApi.info('保存成功！');
        }
        setLoading(false);
    };

    return (
        <>
            {contextHolder}
            <div style={{ width: '100%', height: 'calc(100vh - 200px)', display: 'flex', flexDirection: 'column' }}>
                <div style={{ padding: '10px', display: 'flex', justifyContent: 'flex-end' }}>
                    <Input.TextArea value={editorValue} readOnly
                        hidden={true} style={{ flexGrow: 1, marginRight: '10px' }} />
                    {saveApi && <Button type="primary" onClick={handleSave} loading={loading}>保存</Button>}
                </div>
                <div style={{ flexGrow: 1 }} ref={editorRef} />
            </div>
        </>
    );
};

export default MonacoEditorWithForm;
