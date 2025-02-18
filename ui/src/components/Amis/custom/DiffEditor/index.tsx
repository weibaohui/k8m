import React from 'react';
import * as monaco from 'monaco-editor';
import { DiffEditor } from '@monaco-editor/react';

interface DiffEditorProps {
    originalValue?: string; // 左侧编辑器的内容（历史版本）
    modifiedValue?: string; // 右侧编辑器的内容（最新版本）
    height?: string | number; // 编辑器高度
    width?: string | number; // 编辑器宽度
    readOnly?: boolean; // 是否只读
    originalLabel?: string; // 左侧标签文字
    modifiedLabel?: string; // 右侧标签文字
}

// 用 forwardRef 让组件兼容 AMIS
const DiffEditorComponent = React.forwardRef<HTMLDivElement, DiffEditorProps>(({ 
    originalValue = 'hello',
    modifiedValue = 'hello world',
    height = 'calc(100vh - 100px)',
    width = '100%',
    readOnly = true,
    originalLabel = '历史版本',
    modifiedLabel = '最新版本',
}, _) => {
    // 配置编辑器选项
    const options: monaco.editor.IDiffEditorConstructionOptions = {
        readOnly: readOnly,
        renderSideBySide: true, // 左右分栏显示
        automaticLayout: true, // 自动布局
        scrollBeyondLastLine: false,
        minimap: { enabled: false }, // 禁用小地图
        folding: true, // 启用代码折叠
        lineNumbers: 'on',
        wordWrap: 'on',
        diffWordWrap: 'on'
    };

    return (
        <div style={{ width, height }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
                <div style={{ width: '50%', display: 'flex', justifyContent: 'center' }}>
                    <div style={{ padding: '4px 8px', backgroundColor: '#f0f0f0', borderRadius: '4px', fontSize: '14px' }}>{originalLabel}</div>
                </div>
                <div style={{ width: '50%', display: 'flex', justifyContent: 'center' }}>
                    <div style={{ padding: '4px 8px', backgroundColor: '#f0f0f0', borderRadius: '4px', fontSize: '14px' }}>{modifiedLabel}</div>
                </div>
            </div>
            <div style={{ border: '1px solid #e5e6eb', borderRadius: '4px', boxShadow: '0 2px 4px rgba(0,0,0,0.05)' }}>
                <DiffEditor
                    height="calc(100vh)"
                    language="yaml"
                    original={originalValue}
                    modified={modifiedValue}
                    options={options}
                />
            </div>
        </div>
    );
});

export default DiffEditorComponent;