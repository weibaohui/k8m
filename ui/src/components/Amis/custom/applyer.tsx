import { useState, useEffect, useRef } from 'react';
import { Tabs, Button, List, Input, Modal } from '@arco-design/web-react';
import { IconEye, IconStar, IconDelete, IconEdit } from '@arco-design/web-react/icon';
import * as monaco from 'monaco-editor';

interface RecordItem {
    id: string;
    content: string;
    isFavorite: boolean;
    customName?: string;
}

const HistoryRecords = () => {
    // 初始化记录数据
    const [historyRecords, setHistoryRecords] = useState<RecordItem[]>([]);
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
        const savedHistoryRecords = localStorage.getItem('historyRecords');
        const savedFavoriteRecords = localStorage.getItem('favoriteRecords');

        setHistoryRecords(savedHistoryRecords ? JSON.parse(savedHistoryRecords) : []);
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
    const updateLocalStorage = () => {
        localStorage.setItem('historyRecords', JSON.stringify(historyRecords));
        localStorage.setItem('favoriteRecords', JSON.stringify(favoriteRecords));
    };

    useEffect(() => {
        localStorage.setItem('historyRecords', JSON.stringify(historyRecords));
        localStorage.setItem('favoriteRecords', JSON.stringify(favoriteRecords));
    }, [historyRecords, favoriteRecords]);

    // 收藏某条记录
    const toggleFavorite = (recordId: string) => {
        const record = historyRecords.find(r => r.id === recordId);
        if (record) {
            // 从历史记录中移除
            setHistoryRecords(prevRecords => prevRecords.filter(r => r.id !== recordId));
            // 添加到收藏记录
            setFavoriteRecords(prevRecords => [...prevRecords, { ...record, isFavorite: true }]);
        } else {
            // 从收藏记录中移除
            const favoriteRecord = favoriteRecords.find(r => r.id === recordId);
            if (favoriteRecord) {
                setFavoriteRecords(prevRecords => prevRecords.filter(r => r.id !== recordId));
                // 生成新的ID并添加到历史记录
                setHistoryRecords(prevRecords => [...prevRecords, {
                    ...favoriteRecord,
                    id: Math.random().toString(36).substring(2, 15),
                    isFavorite: false
                }]);
            }
        }
        updateLocalStorage();
    };

    // 保存记录到 localStorage
    const handleSave = () => {
        if (!editorValue) return;

        // 检查记录是否已存在
        const existingRecord = historyRecords.find(record => record.content === editorValue);
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
        setHistoryRecords(prevRecords => [...prevRecords, newRecord]);
        updateLocalStorage();
        setActiveTab('history');
    };

    const handleNameEdit = (recordId: string) => {
        const record = activeTab === 'history'
            ? historyRecords.find(r => r.id === recordId)
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
            setHistoryRecords(prevRecords =>
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

    const handleDelete = (recordId: string) => {
        if (activeTab === 'favorites') {
            const record = favoriteRecords.find(r => r.id === recordId);
            if (record) {
                setFavoriteRecords(prevRecords => prevRecords.filter(r => r.id !== recordId));
                // 生成新的ID并添加到历史记录
                setHistoryRecords(prevRecords => [...prevRecords, {
                    ...record,
                    id: Math.random().toString(36).substring(2, 15),
                    isFavorite: false
                }]);
            }
        } else {
            setHistoryRecords(prevRecords => prevRecords.filter(r => r.id !== recordId));
        }
        updateLocalStorage();
    };

    const renderRecord = (record: RecordItem) => (
        <List.Item key={record.id} data-record-id={record.id} className="list-item" style={{ cursor: 'pointer' }} onClick={() => setSelectedRecord(record.content)}>
            <div style={{ display: 'flex', justifyContent: 'space-between', width: '100%', position: 'relative', backgroundColor: '#FFFFFF' }}>
                {editingId === record.id ? (
                    <Input
                        autoFocus
                        value={editingName}
                        onChange={setEditingName}
                        onBlur={() => handleNameSubmit(record.id)}
                        onPressEnter={() => handleNameSubmit(record.id)}
                        placeholder="请输入新的名称"
                        style={{ maxWidth: '100px' }}
                    />
                ) : (
                    <div
                        style={{
                            maxWidth: '150px',
                            overflow: 'hidden',
                            textOverflow: 'ellipsis',
                            whiteSpace: 'nowrap',
                            flex: 1,
                            zIndex: 2,
                        }}
                    >
                        {record.customName || record.content}
                    </div>
                )}

                <div className="button-group" style={{ position: 'absolute', right: 0, top: '50%', transform: 'translateY(-50%)', zIndex: 3, padding: '0 5px', backgroundColor: '#FFFFFF' }} onClick={(e) => e.stopPropagation()}>
                    <Button.Group>
                        <Button
                            type="text"
                            icon={<IconEdit style={{ fontSize: '14px' }} />}
                            onClick={() => handleNameEdit(record.id)}
                        />
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
                            onClick={() => handleDelete(record.id)}
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
                <Tabs defaultActiveTab="history" onChange={setActiveTab}>
                    <Tabs.TabPane title="历史记录" key="history" >
                        <div style={{ marginBottom: '10px' }}>
                            <Button.Group>
                                <Button
                                    type="outline"
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
                            </Button.Group>
                        </div>
                        <List
                            dataSource={historyRecords.slice((currentPage - 1) * pageSize, currentPage * pageSize)}
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
                                    {currentPage}/{Math.ceil(historyRecords.length / pageSize)}
                                </Button>
                                <Button
                                    type="secondary"
                                    disabled={currentPage >= Math.ceil(historyRecords.length / pageSize)}
                                    onClick={() => setCurrentPage(prev => Math.min(Math.ceil(historyRecords.length / pageSize), prev + 1))}
                                >
                                    下一页
                                </Button>
                            </Button.Group>
                        </div>
                    </Tabs.TabPane>
                    <Tabs.TabPane title="收藏" key="favorites">
                        <div style={{ marginBottom: '10px' }}>
                            <Button.Group>
                                <Button
                                    type="outline"
                                    onClick={() => {
                                        const dataStr = JSON.stringify(favoriteRecords);
                                        const blob = new Blob([dataStr], { type: 'application/json' });
                                        const url = URL.createObjectURL(blob);
                                        const a = document.createElement('a');
                                        a.href = url;
                                        a.download = 'favorites.json';
                                        document.body.appendChild(a);
                                        a.click();
                                        document.body.removeChild(a);
                                        URL.revokeObjectURL(url);
                                    }}
                                >
                                    导出收藏
                                </Button>
                                <Button
                                    type="outline"
                                    onClick={() => {
                                        const input = document.createElement('input');
                                        input.type = 'file';
                                        input.accept = '.json';
                                        input.onchange = (e) => {
                                            const file = (e.target as HTMLInputElement).files?.[0];
                                            if (file) {
                                                const reader = new FileReader();
                                                reader.onload = (e) => {
                                                    try {
                                                        const importedRecords = JSON.parse(e.target?.result as string);
                                                        if (Array.isArray(importedRecords)) {
                                                            const validRecords = importedRecords.filter(record =>
                                                                record.id && record.content && typeof record.isFavorite === 'boolean'
                                                            );
                                                            const newRecords = validRecords.filter(newRecord =>
                                                                !favoriteRecords.some(existingRecord =>
                                                                    existingRecord.id === newRecord.id ||
                                                                    existingRecord.content === newRecord.content
                                                                )
                                                            );
                                                            if (newRecords.length > 0) {
                                                                setFavoriteRecords(prev => [...prev, ...newRecords]);
                                                                Modal.success({
                                                                    title: '导入成功',
                                                                    content: `成功导入 ${newRecords.length} 条记录`
                                                                });
                                                                updateLocalStorage();
                                                            } else {
                                                                Modal.warning({
                                                                    title: '导入提示',
                                                                    content: '没有新的记录需要导入'
                                                                });
                                                            }
                                                        } else {
                                                            throw new Error('Invalid data format');
                                                        }
                                                    } catch (error) {
                                                        Modal.error({
                                                            title: '导入失败',
                                                            content: '文件格式错误或数据无效'
                                                        });
                                                    }
                                                };
                                                reader.readAsText(file);
                                            }
                                        };
                                        input.click();
                                    }}
                                >
                                    导入收藏
                                </Button>
                            </Button.Group>
                        </div>
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
                    border: '1px solid #e5e6eb',
                    borderRadius: '4px'
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