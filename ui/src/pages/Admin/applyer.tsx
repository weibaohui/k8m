import { useState, useEffect, useMemo } from 'react';
import { Tabs, Button, Tooltip, List, Input, Modal } from '@arco-design/web-react';
import { IconEye, IconStar, IconDelete, IconExport } from '@arco-design/web-react/icon';
import axios from 'axios';

interface RecordItem {
    id: string;
    content: string;
    isFavorite: boolean;
    customName?: string;
}

const HistoryRecords = () => {
    // 初始化记录数据
    const [records, setRecords] = useState<RecordItem[]>([]);
    const [selectedRecord, setSelectedRecord] = useState<string>('');
    const [viewRecord, setViewRecord] = useState<string>('');
    const [isModalVisible, setIsModalVisible] = useState(false);
    const [currentPage, setCurrentPage] = useState(1);
    const [currentFavoritePage, setCurrentFavoritePage] = useState(1);
    const [editingId, setEditingId] = useState<string | null>(null);
    const [editingName, setEditingName] = useState('');
    const pageSize = 10;

    // 从 localStorage 获取记录数据
    useEffect(() => {
        const savedRecords = localStorage.getItem('records');
        if (savedRecords) {
            setRecords(JSON.parse(savedRecords));
        }
    }, []);

    // 更新 localStorage 中的数据
    const updateLocalStorage = useMemo(() => {
        return () => {
            localStorage.setItem('records', JSON.stringify(records));
        };
    }, [records]);

    // 收藏某条记录
    const toggleFavorite = (recordId: string) => {
        setRecords(prevRecords =>
            prevRecords.map(record =>
                record.id === recordId
                    ? { ...record, isFavorite: !record.isFavorite }
                    : record
            )
        );
        updateLocalStorage();
    };

    // 保存记录到 localStorage
    const handleSave = () => {
        if (!selectedRecord) return;

        // 检查记录是否已存在
        const existingRecord = records.find(record => record.content === selectedRecord);
        if (existingRecord) {
            // 如果记录已存在，为对应元素添加闪亮动画
            const element = document.querySelector(`[data-record-id="${existingRecord.id}"]`);
            if (element) {
                element.classList.add('highlight-animation');
                // 动画结束后移除类名
                setTimeout(() => {
                    element.classList.remove('highlight-animation');
                }, 1000);
            }
            return;
        }

        // 如果记录不存在，则添加到记录中
        const newRecord: RecordItem = {
            id: Date.now().toString(),
            content: selectedRecord,
            isFavorite: false
        };
        setRecords(prevRecords => [...prevRecords, newRecord]);
        updateLocalStorage();
    };

    const handleNameEdit = (recordId: string) => {
        const record = records.find(r => r.id === recordId);
        if (record) {
            console.log('edit', record);
            setEditingId(record.id);
            setEditingName(record.customName || '');
        } else {
            console.log('edit null', recordId);
            setEditingId(null); // 如果记录不存在，设置为 null
            setEditingName(''); // 清空编辑名称
        }
    };

    const handleNameSubmit = (recordId: string) => {
        if (editingName.trim()) {
            setRecords(prevRecords =>
                prevRecords.map(record =>
                    record.id === recordId
                        ? { ...record, customName: editingName.trim() }
                        : record
                )
            );
        }
        setEditingId(null);
        setEditingName('');
        updateLocalStorage();
    };

    const renderRecord = (record: RecordItem) => (
        <List.Item key={record.id} data-record-id={record.id} className="list-item">
            <div style={{ display: 'flex', justifyContent: 'space-between', width: '100%', position: 'relative', backgroundColor: '#FFFFFF' }}>
                {
                    <>
                        <div>{(record.id === editingId) ? "true" : 'false'} xxxx </div>
                    </>
                }
                {editingId === record.id ? (
                    <Input
                        autoFocus
                        value={editingName}
                        onChange={setEditingName}
                        onBlur={() => handleNameSubmit(record.id)}
                        onPressEnter={() => handleNameSubmit(record.id)}
                        style={{ maxWidth: '120px' }}
                    />
                ) : (
                    <div
                        onClick={() => handleNameEdit(record.id)}
                        style={{
                            maxWidth: '200px',
                            overflow: 'hidden',
                            textOverflow: 'ellipsis',
                            whiteSpace: 'nowrap',
                            cursor: 'pointer',
                            flex: 1,
                            zIndex: 2
                        }}
                    >
                        {record.customName || record.content}
                    </div>
                )}
                <div style={{ position: 'absolute', right: '4px', top: '50%', transform: 'translateY(-50%)', zIndex: 1, pointerEvents: 'none', backgroundColor: '#FFFFFF' }}>
                    {record.isFavorite && (
                        <IconStar style={{ color: '#FFB400', fill: '#FFB400' }} />
                    )}
                </div>
                <div className="button-group" style={{ position: 'absolute', right: 0, top: '50%', transform: 'translateY(-50%)', zIndex: 3, padding: '0 5px', backgroundColor: '#FFFFFF' }} onClick={(e) => e.stopPropagation()}>
                    <Button.Group>
                        <Button
                            type="text"
                            icon={<IconEye style={{ fontSize: '14px' }} />}
                            onClick={() => {
                                setViewRecord(record.content);
                                setIsModalVisible(true);
                            }}
                        />
                        <Button
                            type="text"
                            icon={<IconStar style={{ color: record.isFavorite ? '#FFB400' : '#86909C', fill: record.isFavorite ? '#FFB400' : 'none', fontSize: '14px' }} />}
                            onClick={() => {
                                if (record.isFavorite) {
                                    Modal.confirm({
                                        title: '确认取消收藏',
                                        content: '确定要取消收藏这条记录吗？',
                                        onOk: () => toggleFavorite(record.id)
                                    });
                                } else {
                                    toggleFavorite(record.id);
                                }
                            }}
                        />
                        <Button
                            type="text"
                            icon={<IconDelete style={{ fontSize: '14px' }} />}
                            onClick={() => {
                                setRecords(prevRecords => prevRecords.filter(item => item.id !== record.id));
                                updateLocalStorage();
                            }}
                        />
                        <Button
                            type="text"
                            icon={<IconExport style={{ fontSize: '14px' }} />}
                            onClick={() => setSelectedRecord(record.content)}
                        />
                    </Button.Group>
                </div>
            </div>
        </List.Item>
    );

    return (
        <div style={{ display: 'flex', height: '100vh', backgroundColor: '#FFFFFF' }}>
            <div style={{ width: '350px', padding: '10px', backgroundColor: '#FFFFFF' }}>
                <style>
                    {`
                    .highlight-animation {
                        animation: highlight 1s ease;
                    }
                    @keyframes highlight {
                        0% { box-shadow: inset 0 0 10px rgba(24, 144, 255, 0.7); }
                        50% { box-shadow: inset 0 0 20px rgba(24, 144, 255, 0.9); }
                        100% { box-shadow: inset 0 0 10px rgba(24, 144, 255, 0.7); }
                    }
                    .button-group {
                        opacity: 0;
                        transition: opacity 0.3s ease;
                    }
                    .list-item:hover .button-group {
                        opacity: 1;
                    }
                    `}
                </style>
                <Tabs defaultActiveTab="all">
                    <Tabs.TabPane title="所有" key="all">
                        <List
                            dataSource={records.slice((currentPage - 1) * pageSize, currentPage * pageSize)}
                            render={renderRecord}
                        />
                        <div style={{ marginTop: '16px', textAlign: 'right' }}>
                            <Button.Group>
                                <Button
                                    type="secondary"
                                    disabled={currentPage === 1}
                                    onClick={() => setCurrentPage(prev => Math.max(1, prev - 1))}
                                >
                                    上一页
                                </Button>
                                <Button type="secondary" disabled>
                                    {currentPage}/{Math.ceil(records.length / pageSize)}
                                </Button>
                                <Button
                                    type="secondary"
                                    disabled={currentPage >= Math.ceil(records.length / pageSize)}
                                    onClick={() => setCurrentPage(prev => Math.min(Math.ceil(records.length / pageSize), prev + 1))}
                                >
                                    下一页
                                </Button>
                            </Button.Group>
                        </div>
                    </Tabs.TabPane>
                    <Tabs.TabPane title="收藏" key="favorites">
                        <List
                            dataSource={records.filter(record => record.isFavorite).slice((currentFavoritePage - 1) * pageSize, currentFavoritePage * pageSize)}
                            render={renderRecord}
                        />
                        <div style={{ marginTop: '16px', textAlign: 'right' }}>
                            <Button.Group>
                                <Button
                                    type="secondary"
                                    disabled={currentFavoritePage === 1}
                                    onClick={() => setCurrentFavoritePage(prev => Math.max(1, prev - 1))}
                                >
                                    上一页
                                </Button>
                                <Button type="secondary" disabled>
                                    {currentFavoritePage}/{Math.ceil(records.filter(record => record.isFavorite).length / pageSize)}
                                </Button>
                                <Button
                                    type="secondary"
                                    disabled={currentFavoritePage >= Math.ceil(records.filter(record => record.isFavorite).length / pageSize)}
                                    onClick={() => setCurrentFavoritePage(prev => Math.min(Math.ceil(records.filter(record => record.isFavorite).length / pageSize), prev + 1))}
                                >
                                    下一页
                                </Button>
                            </Button.Group>
                        </div>
                    </Tabs.TabPane>
                </Tabs>
            </div>

            <div style={{ flex: 1, padding: '10px', backgroundColor: '#FFFFFF' }}>
                <Input.TextArea
                    value={selectedRecord}
                    onChange={(value) => setSelectedRecord(value)}
                    rows={10}
                />
                <Button
                    type="primary"
                    style={{ marginTop: '10px' }}
                    onClick={handleSave}
                >
                    应用
                </Button>
            </div>

            <Modal
                title="查看记录"
                visible={isModalVisible}
                onOk={() => setIsModalVisible(false)}
                onCancel={() => setIsModalVisible(false)}
                style={{ width: '600px', backgroundColor: '#FFFFFF' }}
            >
                <div style={{ maxHeight: '400px', overflowY: 'auto', backgroundColor: '#FFFFFF' }}>
                    <pre style={{ whiteSpace: 'pre-wrap', wordWrap: 'break-word' }}>
                        {viewRecord}
                    </pre>
                </div>
            </Modal>
        </div>
    );
};

export default HistoryRecords;