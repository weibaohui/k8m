import React, {useState, useEffect} from 'react';
import {Button, List, Input, Modal, Space} from 'antd';
import {DeleteFilled, EditFilled, StarOutlined} from '@ant-design/icons';

interface RecordItem {
    id: string;
    content: string;
    customName?: string;
}

interface HistoryPanelProps {
    onSelectRecord: (content: string) => void;
    historyRecords: RecordItem[];
    setHistoryRecords: React.Dispatch<React.SetStateAction<RecordItem[]>>;
    onSaveTemplate: (content: string) => void;
}

const HistoryPanel: React.FC<HistoryPanelProps> = ({
                                                       onSelectRecord,
                                                       historyRecords,
                                                       setHistoryRecords,
                                                       onSaveTemplate
                                                   }) => {
    const [editingId, setEditingId] = useState<string>();
    const [editingName, setEditingName] = useState('');
    const [currentPage, setCurrentPage] = useState(1);

    const pageSize = 10;
    const updateLocalStorage = () => {
        localStorage.setItem('historyRecords', JSON.stringify(historyRecords));
    };


    useEffect(() => {
        if (historyRecords.length === 0) {
            return;
        }
        updateLocalStorage();
    }, [historyRecords]);

    const handleNameEdit = (recordId: string) => {
        const record = historyRecords.find(r => r.id === recordId);
        if (record) {
            setEditingId(record.id);
            setEditingName(record.customName || '');
        } else {
            setEditingId('');
            setEditingName('');
        }
    };

    const handleNameSubmit = (recordId: string) => {
        if (editingName.trim()) {
            setHistoryRecords(prevRecords =>
                prevRecords.map(record =>
                    record.id === recordId
                        ? {...record, customName: editingName.trim()}
                        : record
                )
            );
        }
        setEditingId('');
        setEditingName('');
        updateLocalStorage();
    };

    const handleDelete = (recordId: string) => {
        Modal.confirm({
            title: '确认删除',
            content: '确定要删除这条记录吗？',
            onOk: () => {
                setHistoryRecords(prevRecords => prevRecords.filter(r => r.id !== recordId));
                updateLocalStorage();
            }
        });
    };

    const handleSaveTemplate = (recordId: string) => {
        const record = historyRecords.find(r => r.id === recordId);
        if (record) {
            onSaveTemplate(record.content);
        }
    };

    const renderRecord = (record: RecordItem) => (
        <List.Item key={record.id} data-record-id={record.id} className="list-item" style={{cursor: 'pointer'}}
                   onClick={() => onSelectRecord(record.content)}>
            <div style={{
                display: 'flex',
                justifyContent: 'space-between',
                width: '100%',
                position: 'relative',
                backgroundColor: '#FFFFFF'
            }}>
                {editingId === record.id ? (
                    <Input
                        autoFocus
                        value={editingName}
                        onChange={(e) => setEditingName(e.target.value)}
                        onBlur={() => handleNameSubmit(record.id)}
                        onPressEnter={() => handleNameSubmit(record.id)}
                        placeholder="请输入新的名称"
                        style={{maxWidth: '100px'}}
                    />
                ) : (
                    <div style={{
                        flex: 1,
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                        whiteSpace: 'nowrap',
                        marginRight: '10px'
                    }}>
                        {record.customName || record.content}
                    </div>
                )}
                {editingId !== record.id && (
                    <div style={{display: 'flex', gap: '8px', zIndex: 10}}>
                        <Button
                            type="text"
                            icon={<EditFilled style={{color: '#1890ff'}}/>}
                            onClick={(e) => {
                                e.stopPropagation();
                                handleNameEdit(record.id);
                            }}
                        />
                        <Button
                            type="text"
                            icon={<StarOutlined/>}
                            onClick={(e) => {
                                e.stopPropagation();
                                handleSaveTemplate(record.id);
                            }}
                        />
                        <Button
                            type="text"
                            icon={<DeleteFilled style={{color: '#f23034'}}/>}
                            onClick={(e) => {
                                e.stopPropagation();
                                handleDelete(record.id);
                            }}
                        />
                    </div>
                )}
            </div>
        </List.Item>
    );

    return (
        <div>


            <div>
                <div style={{marginTop: '10px', marginBottom: '10px'}}>
                    <Button
                        variant="outlined"
                        onClick={() => {
                            if (historyRecords.length === 0) {
                                Modal.warning({
                                    title: '提示',
                                    content: '暂无历史记录可删除'
                                });
                                return;
                            }
                            Modal.confirm({
                                title: '确认删除',
                                content: '确定要删除所有历史记录吗？此操作不可恢复。',
                                onOk: () => {
                                    setHistoryRecords([]);
                                    updateLocalStorage();
                                    Modal.success({
                                        title: '删除成功',
                                        content: '已清空所有历史记录'
                                    });
                                }
                            });
                        }}
                    >
                        全部删除
                    </Button>
                </div>
                <List
                    dataSource={historyRecords.slice((currentPage - 1) * pageSize, currentPage * pageSize)}
                    renderItem={renderRecord}
                    bordered={true}
                />
                <div style={{marginTop: '16px', textAlign: 'right'}}>
                    <Space.Compact>
                        <Button
                            type="default"
                            disabled={currentPage === 1}
                            onClick={() => setCurrentPage(prev => Math.max(1, prev - 1))}
                        >
                            上一页
                        </Button>
                        <Button type="default" disabled>
                            {currentPage}/{Math.ceil(historyRecords.length / pageSize)}
                        </Button>
                        <Button
                            type="default"
                            disabled={currentPage >= Math.ceil(historyRecords.length / pageSize)}
                            onClick={() => setCurrentPage(prev => Math.min(Math.ceil(historyRecords.length / pageSize), prev + 1))}
                        >
                            下一页
                        </Button>
                    </Space.Compact>
                </div>
            </div>
            <div style={{
                marginTop: '16px',
                padding: '12px',
                backgroundColor: '#f5f5f5',
                borderRadius: '4px',
                fontSize: '10px',
                color: '#666'
            }}>
                提示：历史记录数据存储在浏览器本地缓存中，清除浏览器缓存可能会导致数据丢失。
            </div>
        </div>
    );
};

export default HistoryPanel;