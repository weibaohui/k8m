import React from 'react';
import { Button, Input, message, Modal, Space, Tag } from 'antd';

interface CustomEventTagsProps {
    value?: string;
    onChange?: (value: string) => void;
}

const CustomEventTags: React.FC<CustomEventTagsProps> = ({ value, onChange }) => {
    const [customTags, setCustomTags] = React.useState<Array<{ label: string, value: string }>>(() => {
        try {
            const savedTags = localStorage.getItem('menuCustomTags');
            return savedTags ? JSON.parse(savedTags) : [];
        } catch (e) {
            console.error('Failed to load custom tags:', e);
            return [];
        }
    });

    // 内置的快捷输入选项
    const defaultTags = [
        { label: '页面跳转', value: '() => onMenuClick(\'/cluster/ns\')' },
        { label: '刷新页面', value: '() => window.location.reload()' },
        { label: '控制台日志', value: '() => console.log(\'菜单点击\')' },
    ];

    // 处理添加新的快捷输入
    const handleAddCustomTag = () => {
        const currentValue = value;
        Modal.confirm({
            title: '保存为快捷输入',
            content: (
                <Space.Compact>
                    <Input
                        style={{ width: '30%' }}
                        placeholder="输入标签名"
                        id="tagLabel"
                    />
                    <Input
                        style={{ width: '70%' }}
                        defaultValue={currentValue}
                        placeholder="输入代码"
                        id="tagValue"
                    />
                </Space.Compact>
            ),
            onOk: () => {
                const label = (document.getElementById('tagLabel') as HTMLInputElement)?.value;
                const value = (document.getElementById('tagValue') as HTMLInputElement)?.value;
                if (label && value) {
                    const newCustomTags = [...customTags, { label, value }];
                    setCustomTags(newCustomTags);
                    localStorage.setItem('menuCustomTags', JSON.stringify(newCustomTags));
                    message.success('保存成功');
                }
            }
        });
    };

    // 处理删除自定义标签
    const handleDeleteTag = (index: number) => {
        const newCustomTags = customTags.filter((_, i) => i !== index);
        setCustomTags(newCustomTags);
        localStorage.setItem('menuCustomTags', JSON.stringify(newCustomTags));
        message.success('删除成功');
    };

    return (
        <div style={{ marginBottom: 8 }}>
            <span style={{ marginRight: 8, color: '#666' }}>快捷输入:</span>
            <Space size={[4, 8]} wrap>
                {/* 默认标签 */}
                {defaultTags.map((tag, index) => (
                    <Tag
                        key={`default-${index}`}
                        color="blue"
                        style={{ cursor: 'pointer' }}
                        onClick={() => onChange?.(tag.value)}
                    >
                        {tag.label}
                    </Tag>
                ))}
                {/* 自定义标签 */}
                {customTags.map((tag, index) => (
                    <Tag
                        key={`custom-${index}`}
                        color="blue"
                        style={{ cursor: 'pointer' }}
                        onClick={(e) => {
                            // 防止点击关闭按钮时触发 tag 的点击事件
                            if ((e.target as HTMLElement).tagName.toLowerCase() !== 'svg') {
                                onChange?.(tag.value);
                            }
                        }}
                        closable
                        onClose={() => handleDeleteTag(index)}
                    >
                        {tag.label}
                    </Tag>
                ))}
            </Space>
            <Button
                type="link"
                size="small"
                style={{ padding: '2px 8px' }}
                onClick={handleAddCustomTag}
            >
                +添加快捷输入
            </Button>
        </div>
    );
};

export default CustomEventTags;
