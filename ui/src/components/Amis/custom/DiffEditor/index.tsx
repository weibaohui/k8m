import React from 'react';
import { loader } from '@monaco-editor/react';

import * as monaco from 'monaco-editor';
import { DiffEditor } from '@monaco-editor/react';
import yaml from "js-yaml";
interface DiffEditorProps {
    originalValue?: string; // 左侧编辑器的内容（历史版本）
    modifiedValue?: string; // 右侧编辑器的内容（最新版本）
    height?: string | number; // 编辑器高度
    width?: string | number; // 编辑器宽度
    readOnly?: boolean; // 是否只读
    originalLabel?: string; // 左侧标签文字
    modifiedLabel?: string; // 右侧标签文字
    data?: Record<string, any>;
    language?: string;
}

// 用 forwardRef 让组件兼容 AMIS

const DiffEditorComponent = React.forwardRef<HTMLDivElement, DiffEditorProps>((props, _) => {

    let originalValue = props.originalValue || 'hello';
    let modifiedValue = props.modifiedValue || 'hello world';
    let height = props.height || 'calc(100vh - 100px)';
    let width = props.width || '100%';
    let readOnly = props.readOnly || true;
    let originalLabel = props.originalLabel || '历史版本';
    let modifiedLabel = props.modifiedLabel || '最新版本';
    let language = props.language || 'yaml';

    function getByPath(data: Record<string, any>, path: string) {
        const keys = path.split(".");
        let result = data;

        for (let key of keys) {
            // 支持数组下标取值
            const match = key.match(/(.*?)\[(\d+)\]/);
            if (match) {
                key = match[1];
                const index = parseInt(match[2], 10);
                result = result?.[key]?.[index];
            } else {
                result = result?.[key];
            }

            if (result === undefined) {
                return null;
            }
        }

        return result;
    }

    function extractValue(data: Record<string, any>, expr: string) {
        const regex = /^\$\{(.+?)\}$/;
        const match = expr.match(regex);
        if (match) {
            const path = match[1];
            return getByPath(data, path);
        }
        return expr; // 直接返回字符串
    }

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
        diffWordWrap: 'on',
        hideUnchangedRegions: {
            enabled: true,
            minimumLineCount: 3,    // 当未更改区域超过3行时触发折叠
            contextLineCount: 1     // 折叠时保留1行上下文
        }
    };
    if (props.data) {
        let originalValueObj = extractValue(props.data, originalValue,);
        let modifiedValueObj = extractValue(props.data, modifiedValue,);
        if (language === 'yaml') {
            try {
                originalValue = yaml.dump(originalValueObj, {
                    indent: 2,
                    lineWidth: -1,  // 禁用自动换行
                    noRefs: true    // 避免引用标记
                });
                modifiedValue = yaml.dump(modifiedValueObj, {
                    indent: 2,
                    lineWidth: -1,
                    noRefs: true
                });
            } catch (error) {
                console.error('YAML转换错误:', error);
                // 转换失败时使用原始值
                originalValue = String(originalValueObj);
                modifiedValue = String(modifiedValueObj);
            }
        } else if (language === 'json') {
            originalValue = JSON.stringify(originalValueObj, null, 4);
            modifiedValue = JSON.stringify(modifiedValueObj, null, 4);
        }

    }
    loader.config({
        paths: {
            vs: '/monacoeditorwork'
        }
    })
    return (
        <div style={{ width, height }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
                <div style={{ width: '50%', display: 'flex', justifyContent: 'center' }}>
                    <div style={{
                        padding: '4px 8px',
                        backgroundColor: '#f0f0f0',
                        borderRadius: '4px',
                        fontSize: '14px'
                    }}>{originalLabel}</div>
                </div>
                <div style={{ width: '50%', display: 'flex', justifyContent: 'center' }}>
                    <div style={{
                        padding: '4px 8px',
                        backgroundColor: '#f0f0f0',
                        borderRadius: '4px',
                        fontSize: '14px'
                    }}>{modifiedLabel}</div>
                </div>
            </div>
            <div style={{ border: '1px solid #e5e6eb', borderRadius: '4px', boxShadow: '0 2px 4px rgba(0,0,0,0.05)' }}>
                <DiffEditor
                    height="calc(100vh)"
                    language={language}
                    original={originalValue}
                    modified={modifiedValue}
                    options={options}

                />
            </div>
        </div>
    );
});

export default DiffEditorComponent;