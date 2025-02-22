import React, {useState} from 'react';
import {message} from 'antd';
import HistoryPanel from './components/HistoryPanel';
import EditorPanel from './components/EditorPanel';
import TemplatePanel from './components/TemplatePanel';
import {fetcher} from '@/components/Amis/fetcher';

// 用 forwardRef 让组件兼容 AMIS
const YamlApplyer = React.forwardRef<HTMLDivElement>(() => {
    const [editorContent, setEditorContent] = useState<string>('');
    const [historyRecords, setHistoryRecords] = useState<any[]>([]);
    const [templateRefreshKey, setTemplateRefreshKey] = useState<number>(0);

    // 初始化时从localStorage加载历史记录
    React.useEffect(() => {
        const savedHistoryRecords = localStorage.getItem('historyRecords');
        if (savedHistoryRecords) {
            setHistoryRecords(JSON.parse(savedHistoryRecords));
        }
    }, []);

    const handleRecordSelect = (content: string) => {
        setEditorContent(content);
    };

    const handleSaveTemplate = async (content: string) => {
        try {
            const newTemplate = {
                name: `模板-${Math.random().toString(36).substring(2, 8)}`,
                content: content,
                kind: ''
            };
            const response = await fetcher({
                url: '/mgm/custom/template/save',
                method: 'post',
                data: newTemplate
            });
            if (response.data?.status === 0) {
                message.success('已保存为模板');
                // 触发模板列表刷新
                setTemplateRefreshKey(prev => prev + 1);
            } else {
                throw new Error(response.data?.msg || '保存模板失败');
            }
        } catch (error) {
            message.error('保存模板失败：' + (error instanceof Error ? error.message : '未知错误'));
        }
    };

    const handleSaveSuccess = (content: string) => {
        setEditorContent(content);

        // 检查是否已存在相同内容的记录
        const existingRecord = historyRecords.find(record => record.content === content);
        if (existingRecord) {
            message.success('已保存到历史记录');
            return; // 如果已存在相同内容的记录，则不添加新记录
        }

        // 创建新的记录
        const newRecord = {
            id: Math.random().toString(36).substring(2, 15),
            content: content,
            isFavorite: false
        };

        // 更新状态和本地存储
        const updatedRecords = [newRecord, ...historyRecords];
        setHistoryRecords(updatedRecords);
        localStorage.setItem('historyRecords', JSON.stringify(updatedRecords));
        message.success('已保存到历史记录');
    };

    return (
        <div style={{height: '100%', display: 'flex'}}>
            <div style={{width: '25%', borderRight: '1px solid #e5e6eb', padding: '10px', overflowY: 'auto'}}>
                <TemplatePanel onSelectTemplate={handleRecordSelect} refreshKey={templateRefreshKey}/>
            </div>
            <div style={{width: '50%', padding: '10px', overflowY: 'auto'}}>
                <EditorPanel onSaveSuccess={handleSaveSuccess} initialContent={editorContent}/>
            </div>
            <div style={{width: '25%', borderLeft: '1px solid #e5e6eb', padding: '10px', overflowY: 'auto'}}>
                <HistoryPanel
                    onSelectRecord={handleRecordSelect}
                    historyRecords={historyRecords}
                    setHistoryRecords={setHistoryRecords}
                    onSaveTemplate={handleSaveTemplate}/>
            </div>
        </div>
    );
});

export default YamlApplyer;
