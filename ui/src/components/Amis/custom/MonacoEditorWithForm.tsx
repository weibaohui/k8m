import React, {useEffect, useRef, useState} from 'react';
import * as monaco from 'monaco-editor';
import {Button, Input} from "@arco-design/web-react";
import {replacePlaceholders} from "@/utils/utils.ts";
import {fetcher} from "@/components/Amis/fetcher.ts";

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
    text = replacePlaceholders(text, data)
    if (saveApi) {
        saveApi = replacePlaceholders(saveApi, data)
    }
    useEffect(() => {
        if (editorRef.current) {
            monacoInstance.current = monaco.editor.create(editorRef.current, {
                value: text,
                language: 'yaml',
                theme: 'vs-dark',
                automaticLayout: true,
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
        try {
            await fetcher({url: saveApi, method: 'post', data: {[componentId]: editorValue}});
            alert('保存成功！');
        } catch (error) {
            alert('保存失败，请检查日志');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div style={{width: '100%', height: 'calc(100vh - 200px)', display: 'flex', flexDirection: 'column'}}>

            <div style={{padding: '10px', display: 'flex', justifyContent: 'space-between', background: '#222'}}>
                <Input.TextArea value={editorValue} readOnly
                                hidden={true} style={{flexGrow: 1, marginRight: '10px'}}/>
                {saveApi && <Button type="primary" onClick={handleSave} loading={loading}>保存</Button>}
            </div>
            <div style={{flexGrow: 1}} ref={editorRef}/>
        </div>
    );
};

export default MonacoEditorWithForm;
