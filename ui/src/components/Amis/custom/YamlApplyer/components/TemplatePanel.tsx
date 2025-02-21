import React, { useState } from 'react';
import { Button, List, Input, Modal, Space } from 'antd';
import { DeleteFilled, EditFilled } from '@ant-design/icons';

interface TemplateItem {
    id: string;
    name: string;
    content: string;
}

interface TemplatePanelProps {
    onSelectTemplate: (content: string) => void;
}

const TemplatePanel: React.FC<TemplatePanelProps> = ({ onSelectTemplate }) => {
    const [templates, setTemplates] = useState<TemplateItem[]>([]);
    const [editingId, setEditingId] = useState<string>();
    const [editingName, setEditingName] = useState('');
    const [currentPage, setCurrentPage] = useState(1);

    const pageSize = 10;

    const handleNameEdit = (templateId: string) => {
        const template = templates.find(t => t.id === templateId);
        if (template) {
            setEditingId(template.id);
            setEditingName(template.name);
        }
    };

    const handleNameSubmit = (templateId: string) => {
        if (editingName.trim()) {
            setTemplates(prevTemplates =>
                prevTemplates.map(template =>
                    template.id === templateId
                        ? { ...template, name: editingName.trim() }
                        : template
                )
            );
        }
        setEditingId('');
        setEditingName('');
    };

    const handleDelete = (templateId: string) => {
        Modal.confirm({
            title: '确认删除',
            content: '确定要删除这个模板吗？',
            onOk: () => {
                setTemplates(prevTemplates => prevTemplates.filter(t => t.id !== templateId));
            }
        });
    };

    const renderTemplate = (template: TemplateItem) => (
        <List.Item key={template.id} className="list-item" style={{ cursor: 'pointer' }}
                   onClick={() => onSelectTemplate(template.content)}>
            <div style={{
                display: 'flex',
                justifyContent: 'space-between',
                width: '100%',
                position: 'relative',
                backgroundColor: '#FFFFFF'
            }}>
                {editingId === template.id ? (
                    <Input
                        autoFocus
                        value={editingName}
                        onChange={(e) => setEditingName(e.target.value)}
                        onBlur={() => handleNameSubmit(template.id)}
                        onPressEnter={() => handleNameSubmit(template.id)}
                        placeholder="请输入新的名称"
                        style={{ maxWidth: '100px' }}
                    />
                ) : (
                    <div style={{
                        flex: 1,
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                        whiteSpace: 'nowrap',
                        marginRight: '10px'
                    }}>
                        {template.name}
                    </div>
                )}
                {editingId !== template.id && (
                    <div style={{ display: 'flex', gap: '8px', zIndex: 10 }}>
                        <Button
                            type="text"
                            icon={<EditFilled style={{ color: '#1890ff' }} />}
                            onClick={(e) => {
                                e.stopPropagation();
                                handleNameEdit(template.id);
                            }}
                        />
                        <Button
                            type="text"
                            icon={<DeleteFilled style={{ color: '#f23034' }} />}
                            onClick={(e) => {
                                e.stopPropagation();
                                handleDelete(template.id);
                            }}
                        />
                    </div>
                )}
            </div>
        </List.Item>
    );

    return (
        <div>
            <div style={{ marginBottom: '10px' }}>
                <Space.Compact>
                    <Button
                        variant="outlined"
                        onClick={() => {
                            // 添加新模板的逻辑
                            const newTemplate: TemplateItem = {
                                id: Math.random().toString(36).substring(2, 15),
                                name: `模板 ${templates.length + 1}`,
                                content: ''
                            };
                            setTemplates(prev => [...prev, newTemplate]);
                        }}
                    >
                        新建模板
                    </Button>
                </Space.Compact>
            </div>
            <List
                dataSource={templates.slice((currentPage - 1) * pageSize, currentPage * pageSize)}
                renderItem={renderTemplate}
                bordered={true}
            />
            <div style={{ marginTop: '16px', textAlign: 'right' }}>
                <Space.Compact>
                    <Button
                        type="default"
                        disabled={currentPage === 1}
                        onClick={() => setCurrentPage(prev => Math.max(1, prev - 1))}
                    >
                        上一页
                    </Button>
                    <Button type="default" disabled>
                        {currentPage}/{Math.ceil(templates.length / pageSize)}
                    </Button>
                    <Button
                        type="default"
                        disabled={currentPage >= Math.ceil(templates.length / pageSize)}
                        onClick={() => setCurrentPage(prev => Math.min(Math.ceil(templates.length / pageSize), prev + 1))}
                    >
                        下一页
                    </Button>
                </Space.Compact>
            </div>
        </div>
    );
};

export default TemplatePanel;