import React, { useState, useEffect } from 'react';
import { Button, List, Input, Modal, Space } from 'antd';
import { DeleteFilled, EditFilled, StarFilled, StarOutlined } from '@ant-design/icons';
import JSZip from 'jszip';
import { saveAs } from 'file-saver';

interface RecordItem {
    id: string;
    content: string;
    isFavorite: boolean;
    customName?: string;
}

interface HistoryPanelProps {
    onSelectRecord: (content: string) => void;
    historyRecords: RecordItem[];
    setHistoryRecords: React.Dispatch<React.SetStateAction<RecordItem[]>>;
}

const HistoryPanel: React.FC<HistoryPanelProps> = ({ onSelectRecord, historyRecords, setHistoryRecords }) => {
    const [favoriteRecords, setFavoriteRecords] = useState<RecordItem[]>([]);
    const [editingId, setEditingId] = useState<string>();
    const [editingName, setEditingName] = useState('');
    const [activeTab, setActiveTab] = useState('history');
    const [currentPage, setCurrentPage] = useState(1);
    const [currentFavoritePage, setCurrentFavoritePage] = useState(1);

    const pageSize = 10;

    useEffect(() => {
        const savedFavoriteRecords = localStorage.getItem('favoriteRecords');
        setFavoriteRecords(savedFavoriteRecords ? JSON.parse(savedFavoriteRecords) : []);
    }, []);

    const updateLocalStorage = () => {
        localStorage.setItem('historyRecords', JSON.stringify(historyRecords));
        localStorage.setItem('favoriteRecords', JSON.stringify(favoriteRecords));
    };

    useEffect(() => {
        updateLocalStorage();
    }, [historyRecords, favoriteRecords]);

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
        Modal.confirm({
            title: '确认删除',
            content: '确定要删除这条记录吗？',
            onOk: () => {
                if (activeTab === 'favorites') {
                    const record = favoriteRecords.find(r => r.id === recordId);
                    if (record) {
                        setFavoriteRecords(prevRecords => prevRecords.filter(r => r.id !== recordId));
                        setHistoryRecords(prevRecords => [{
                            ...record,
                            id: Math.random().toString(36).substring(2, 15),
                            isFavorite: false
                        }, ...prevRecords]);
                    }
                } else {
                    setHistoryRecords(prevRecords => prevRecords.filter(r => r.id !== recordId));
                }
                updateLocalStorage();
            }
        });
    };

    const toggleFavorite = (recordId: string) => {
        const record = historyRecords.find(r => r.id === recordId);
        if (record) {
            setHistoryRecords(prevRecords => prevRecords.filter(r => r.id !== recordId));
            setFavoriteRecords(prevRecords => [{ ...record, isFavorite: true }, ...prevRecords]);
            updateLocalStorage();
        } else {
            const favoriteRecord = favoriteRecords.find(r => r.id === recordId);
            if (favoriteRecord) {
                Modal.confirm({
                    title: '确认取消收藏',
                    content: '确定要取消收藏这条记录吗？',
                    onOk: () => {
                        setFavoriteRecords(prevRecords => prevRecords.filter(r => r.id !== recordId));
                        setHistoryRecords(prevRecords => [{
                            ...favoriteRecord,
                            id: Math.random().toString(36).substring(2, 15),
                            isFavorite: false
                        }, ...prevRecords]);
                        updateLocalStorage();
                    }
                });
            }
        }
    };

    const renderRecord = (record: RecordItem) => (
        <List.Item key={record.id} data-record-id={record.id} className="list-item" style={{ cursor: 'pointer' }}
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
                        {record.customName || record.content}
                    </div>
                )}
                {editingId !== record.id && (
                    <div style={{ display: 'flex', gap: '8px', zIndex: 10 }}>
                        <Button
                            type="text"
                            icon={<EditFilled style={{ color: '#1890ff' }} />}
                            onClick={(e) => {
                                e.stopPropagation();
                                handleNameEdit(record.id);
                            }}
                        />
                        <Button
                            type="text"
                            icon={activeTab === 'favorites' ? <StarFilled style={{ color: '#FFD700' }} /> :
                                <StarOutlined />}
                            onClick={(e) => {
                                e.stopPropagation();
                                toggleFavorite(record.id);
                            }}
                        />
                        <Button
                            type="text"
                            icon={<DeleteFilled style={{ color: '#f23034' }} />}
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
            <Space.Compact>
                <Button
                    variant="outlined"
                    onClick={() => setActiveTab('history')}
                    type={activeTab === 'history' ? 'primary' : 'default'}
                >
                    历史记录
                </Button>
                <Button
                    variant="outlined"
                    onClick={() => setActiveTab('favorites')}
                    type={activeTab === 'favorites' ? 'primary' : 'default'}
                >
                    收藏
                </Button>
            </Space.Compact>

            {activeTab === 'history' ? (
                <div>
                    <div style={{ marginTop: '10px', marginBottom: '10px' }}>
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
            ) : (
                <div>
                    <div style={{ marginTop: '10px', marginBottom: '10px' }}>
                        <Space.Compact>
                            <Button
                                variant="outlined"
                                onClick={() => {
                                    const input = document.createElement('input');
                                    input.type = 'file';
                                    input.accept = '.zip';
                                    input.onchange = async (e) => {
                                        const file = (e.target as HTMLInputElement).files?.[0];
                                        if (file) {
                                            try {
                                                const zip = await JSZip.loadAsync(file);
                                                const yamlFiles = [];

                                                for (const [fileName, fileData] of Object.entries(zip.files)) {
                                                    if (fileName.endsWith('.yaml') || fileName.endsWith('.yml')) {
                                                        const content = await fileData.async('text');
                                                        const customName = fileName.replace(/\.(yaml|yml)$/, '');
                                                        yamlFiles.push({
                                                            id: Math.random().toString(36).substring(2, 15),
                                                            content,
                                                            isFavorite: true,
                                                            customName
                                                        });
                                                    }
                                                }

                                                if (yamlFiles.length > 0) {
                                                    const newRecords = yamlFiles.filter(newRecord =>
                                                        !favoriteRecords.some(existingRecord =>
                                                            existingRecord.content === newRecord.content
                                                        )
                                                    );

                                                    if (newRecords.length > 0) {
                                                        setFavoriteRecords(prev => [...prev, ...newRecords]);
                                                        Modal.success({
                                                            title: '导入成功',
                                                            content: `成功导入 ${newRecords.length} 个YAML文件`
                                                        });
                                                        updateLocalStorage();
                                                    } else {
                                                        Modal.warning({
                                                            title: '导入提示',
                                                            content: '没有新的YAML文件需要导入'
                                                        });
                                                    }
                                                } else {
                                                    Modal.warning({
                                                        title: '导入提示',
                                                        content: 'ZIP文件中没有找到YAML文件'
                                                    });
                                                }
                                            } catch (error) {
                                                Modal.error({
                                                    title: '导入失败',
                                                    content: '无法解析ZIP文件或文件格式错误'
                                                });
                                            }
                                        }
                                    };
                                    input.click();
                                }}
                            >
                                导入备份
                            </Button>
                            <Button
                                variant="outlined"
                                onClick={async () => {
                                    const zip = new JSZip();
                                    favoriteRecords.forEach((record, index) => {
                                        const fileName = record.customName || `favorite_${index + 1}.yaml`;
                                        zip.file(fileName.endsWith('.yaml') ? fileName : `${fileName}.yaml`, record.content);
                                    });
                                    const blob = await zip.generateAsync({ type: 'blob' });
                                    saveAs(blob, 'favorites.zip');
                                }}
                            >
                                导出备份
                            </Button>
                        </Space.Compact>
                    </div>
                    <List
                        dataSource={favoriteRecords.slice((currentFavoritePage - 1) * pageSize, currentFavoritePage * pageSize)}
                        renderItem={renderRecord}
                        bordered={true}
                    />
                    <div style={{ marginTop: '16px', textAlign: 'right' }}>
                        <Space.Compact>
                            <Button
                                type="default"
                                disabled={currentFavoritePage === 1}
                                onClick={() => setCurrentFavoritePage(prev => Math.max(1, prev - 1))}
                            >
                                上一页
                            </Button>
                            <Button type="default" disabled>
                                {currentFavoritePage}/{Math.ceil(favoriteRecords.length / pageSize)}
                            </Button>
                            <Button
                                type="default"
                                disabled={currentFavoritePage >= Math.ceil(favoriteRecords.length / pageSize)}
                                onClick={() => setCurrentFavoritePage(prev => Math.min(Math.ceil(favoriteRecords.length / pageSize), prev + 1))}
                            >
                                下一页
                            </Button>
                        </Space.Compact>
                    </div>
                </div>
            )}
        </div>
    );
};

export default HistoryPanel;