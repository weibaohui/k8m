import { useState, useEffect, useMemo, useRef } from 'react';
import { Tabs, Button, Tooltip, List, Input, Modal } from '@arco-design/web-react';
import { IconEye, IconStar, IconDelete, IconExport } from '@arco-design/web-react/icon';
import axios from 'axios';
import * as monaco from 'monaco-editor';

interface RecordItem {
    id: string;
    content: string;
    isFavorite: boolean;
    customName?: string;
}

const HistoryRecords = () => {
    // 初始化记录数据
    const [allRecords, setAllRecords] = useState<RecordItem[]>([]);
    const [favoriteRecords, setFavoriteRecords] = useState<RecordItem[]>([]);
    const [selectedRecord, setSelectedRecord] = useState<string>('');
    const [viewRecord, setViewRecord] = useState<RecordItem>();
    const [isModalVisible, setIsModalVisible] = useState(false);
    const [currentPage, setCurrentPage] = useState(1);
    const [currentFavoritePage, setCurrentFavoritePage] = useState(1);
    const [editingId, setEditingId] = useState<string>();
    const [editingName, setEditingName] = useState('');
    const [activeTab, setActiveTab] = useState('all');
    const monacoInstance = useRef<monaco.editor.IStandaloneCodeEditor | null>(null);
    const [editorValue, setEditorValue] = useState<string>();
    const editorRef = useRef<HTMLDivElement>(null);

    const pageSize = 10;

    // 从 localStorage 获取记录数据
    useEffect(() => {
        const savedAllRecords = localStorage.getItem('allRecords');
        const savedFavoriteRecords = localStorage.getItem('favoriteRecords');

        setAllRecords(savedAllRecords ? JSON.parse(savedAllRecords) : []);
        setFavoriteRecords(savedFavoriteRecords ? JSON.parse(savedFavoriteRecords) : []);

    }, []);
    useEffect(() => {
        if (editorRef.current) {
            monacoInstance.current = monaco.editor.create(editorRef.current, {
                value: '',
                theme: 'vs',
                automaticLayout: true,
                minimap: {
                    enabled: false // 关闭小地图
                },
            });
            monacoInstance.current.onDidChangeModelContent(() => {
                setEditorValue(monacoInstance.current?.getValue() || '');
            });
        }
        return () => monacoInstance.current?.dispose();
    }, []);
    useEffect(() => {
        if (monacoInstance.current && selectedRecord !== monacoInstance.current.getValue()) {
            monacoInstance.current.setValue(selectedRecord);
        }
    }, [selectedRecord]);

    // 更新 localStorage 中的数据
    const updateLocalStorage = useMemo(() => {
        return () => {
            localStorage.setItem('allRecords', JSON.stringify(allRecords));
            localStorage.setItem('favoriteRecords', JSON.stringify(favoriteRecords));
        };
    }, [allRecords, favoriteRecords]);

    // 收藏某条记录
    const toggleFavorite = (recordId: string) => {
        const record = allRecords.find(r => r.id === recordId);
        if (record) {
            // 从所有记录中移除
            setAllRecords(prevRecords => prevRecords.filter(r => r.id !== recordId));
            // 添加到收藏记录
            setFavoriteRecords(prevRecords => [...prevRecords, { ...record, isFavorite: true }]);
        } else {
            // 从收藏记录中移除
            const favoriteRecord = favoriteRecords.find(r => r.id === recordId);
            if (favoriteRecord) {
                setFavoriteRecords(prevRecords => prevRecords.filter(r => r.id !== recordId));
                setAllRecords(prevRecords => [...prevRecords, { ...favoriteRecord, isFavorite: false }]);
            }
        }
        updateLocalStorage();
    };

    // 保存记录到 localStorage
    const handleSave = () => {
        if (!editorValue) return;

        // 检查记录是否已存在
        const existingRecord = allRecords.find(record => record.content === editorValue);
        if (existingRecord) {
            const element = document.querySelector(`[data-record-id="${existingRecord.id}"]`);
            if (element) {
                element.classList.add('highlight-animation');
                setTimeout(() => {
                    element.classList.remove('highlight-animation');
                }, 1000);
            }
            return;
        }

        const newRecord: RecordItem = {
            id: Math.random().toString(36).substring(2, 15),
            content: editorValue,
            isFavorite: false
        };
        setAllRecords(prevRecords => [...prevRecords, newRecord]);
        updateLocalStorage();
    };

    const handleNameEdit = (recordId: string) => {
        const record = activeTab === 'all'
            ? allRecords.find(r => r.id === recordId)
            : favoriteRecords.find(r => r.id === recordId);

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
            setAllRecords(prevRecords =>
                prevRecords.map(record =>
                    record.id === recordId
                        ? { ...record, customName: editingName.trim() }
                        : record
                )
            );
            // 同步更新收藏列表中的记录名称
            setFavoriteRecords(prevRecords =>
                prevRecords.map(record =>
                    record.id === recordId
                        ? { ...record, customName: editingName.trim() }
                        : record
                )
            );
        }
        setEditingId('');
        setEditingName('');
        updateLocalStorage();
    };

    const renderRecord = (record: RecordItem) => (
        <List.Item key={record.id} data-record-id={record.id} className="list-item">
            <div style={{ display: 'flex', justifyContent: 'space-between', width: '100%', position: 'relative', backgroundColor: '#FFFFFF' }}>
                {editingId === record.id ? (
                    <Input
                        autoFocus
                        value={editingName}
                        onChange={setEditingName}
                        onBlur={() => handleNameSubmit(record.id)}
                        onPressEnter={() => handleNameSubmit(record.id)}
                        style={{ maxWidth: '150px' }}
                    />
                ) : (
                    <div
                        onClick={() => handleNameEdit(record.id)}
                        style={{
                            maxWidth: '150px',
                            overflow: 'hidden',
                            textOverflow: 'ellipsis',
                            whiteSpace: 'nowrap',
                            cursor: 'pointer',
                            flex: 1,
                            zIndex: 2,
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
                                setViewRecord(record);
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
                                setAllRecords(prevRecords => prevRecords.filter(item => item.id !== record.id));
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
                <Tabs defaultActiveTab="all" onChange={setActiveTab}>
                    <Tabs.TabPane title="所有" key="all" >
                        <List
                            dataSource={allRecords.slice((currentPage - 1) * pageSize, currentPage * pageSize)}
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
                                    {currentPage}/{Math.ceil(allRecords.length / pageSize)}
                                </Button>
                                <Button
                                    type="secondary"
                                    disabled={currentPage >= Math.ceil(allRecords.length / pageSize)}
                                    onClick={() => setCurrentPage(prev => Math.min(Math.ceil(allRecords.length / pageSize), prev + 1))}
                                >
                                    下一页
                                </Button>
                            </Button.Group>
                        </div>
                    </Tabs.TabPane>
                    <Tabs.TabPane title="收藏" key="favorites">
                        <List
                            dataSource={favoriteRecords.slice((currentFavoritePage - 1) * pageSize, currentFavoritePage * pageSize)}
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
                                    {currentFavoritePage}/{Math.ceil(favoriteRecords.length / pageSize)}
                                </Button>
                                <Button
                                    type="secondary"
                                    disabled={currentFavoritePage >= Math.ceil(favoriteRecords.length / pageSize)}
                                    onClick={() => setCurrentFavoritePage(prev => Math.min(Math.ceil(favoriteRecords.length / pageSize), prev + 1))}
                                >
                                    下一页
                                </Button>
                            </Button.Group>
                        </div>
                    </Tabs.TabPane>
                </Tabs>
            </div>

            <div style={{ padding: '10px', backgroundColor: '#FFFFFF' }}>
                <div ref={editorRef} style={{
                    minWidth: '500px',
                    width: 'calc(100vh - 200px)',
                    height: '80vh',
                }} />
                <Button
                    type="primary"
                    style={{ marginTop: '10px' }}
                    onClick={handleSave}
                >
                    应用
                </Button>
            </div>

            <Modal
                title={`查看记录${viewRecord?.customName || ''}`}
                visible={isModalVisible}
                onOk={() => setIsModalVisible(false)}
                onCancel={() => setIsModalVisible(false)}
                style={{ width: '600px', backgroundColor: '#FFFFFF' }}
            >
                <div style={{ maxHeight: '400px', overflowY: 'auto', backgroundColor: '#FFFFFF' }}>
                    <pre style={{ whiteSpace: 'pre-wrap', wordWrap: 'break-word' }}>
                        {viewRecord?.content}
                    </pre>
                </div>
            </Modal>
        </div >
    );
};

export default HistoryRecords;