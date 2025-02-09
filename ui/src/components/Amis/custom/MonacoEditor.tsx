import React, {useEffect, useRef} from 'react';
import * as monaco from 'monaco-editor';
import {replacePlaceholders} from "@/utils/utils.ts";

interface MonacoEditorProps {
    text: string;
    options?: monaco.editor.IStandaloneEditorConstructionOptions;
    data: Record<string, any>
}


// 用 forwardRef 包装组件
const MonacoEditorComponent = React.forwardRef<HTMLSpanElement, MonacoEditorProps>(({
                                                                                        text,
                                                                                        options,
                                                                                        data,
                                                                                    }, _) => {
    const editorRef = useRef<HTMLDivElement>(null);
    const monacoInstance = useRef<monaco.editor.IStandaloneCodeEditor | null>(null);
    text = replacePlaceholders(text, data)
    useEffect(() => {
        if (editorRef.current) {
            monacoInstance.current = monaco.editor.create(editorRef.current, {
                value: text,
                language: 'yaml', // 默认语言，可根据 options 传入自定义
                theme: 'vs-dark',
                automaticLayout: true,
                ...options,
            });
        }

        return () => monacoInstance.current?.dispose();
    }, []);

    useEffect(() => {
        if (monacoInstance.current && text != undefined && text !== monacoInstance.current.getValue()) {
            monacoInstance.current.setValue(text);
        }
    }, [text]);

    return <div ref={editorRef} style={{width: '100%', height: '100vh'}}/>;
});


export default MonacoEditorComponent;
