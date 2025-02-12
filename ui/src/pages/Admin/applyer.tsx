import { useState, useEffect, useMemo } from 'react';
import { Tabs, Button, Tooltip, List, Input, Modal } from '@arco-design/web-react';
import { IconEye, IconStar, IconDelete, IconExport } from '@arco-design/web-react/icon';
import axios from 'axios';

const HistoryRecords = () => {
    // 初始化历史记录和收藏记录
    const [allRecords, setAllRecords] = useState<string[]>([]);
    const [favoriteRecords, setFavoriteRecords] = useState<string[]>([]);
    const [selectedRecord, setSelectedRecord] = useState<string>('');
    const [viewRecord, setViewRecord] = useState<string>('');
    const [isModalVisible, setIsModalVisible] = useState(false);
    const [currentPage, setCurrentPage] = useState(1);
    const [currentFavoritePage, setCurrentFavoritePage] = useState(1);
    const [customNames, setCustomNames] = useState<Record<string, string>>({});
    const [editingIndex, setEditingIndex] = useState<number | null>(null);
    const [editingName, setEditingName] = useState('');
    const pageSize = 10;

    // 从 localStorage 获取历史记录和收藏记录
    useEffect(() => {
        const all = localStorage.getItem('allRecords');
        const favorites = localStorage.getItem('favoriteRecords');
        const names = localStorage.getItem('customNames');
        setAllRecords(all ? JSON.parse(all) : []);
        setFavoriteRecords(favorites ? JSON.parse(favorites) : []);
        setCustomNames(names ? JSON.parse(names) : {});
    }, []);

    // 更新 localStorage 中的数据
    const updateLocalStorage = useMemo(() => {
        return () => {
            localStorage.setItem('allRecords', JSON.stringify(allRecords));
            localStorage.setItem('favoriteRecords', JSON.stringify(favoriteRecords));
            localStorage.setItem('customNames', JSON.stringify(customNames));
        };
    }, [allRecords, favoriteRecords, customNames]);

    // 收藏某条记录
    const toggleFavorite = (record: string) => {
        const updatedFavoriteRecords = [...favoriteRecords];
        const index = updatedFavoriteRecords.findIndex(item => item === record);

        if (index !== -1) {
            // 如果已经在收藏中，移除收藏
            updatedFavoriteRecords.splice(index, 1);
        } else {
            // 否则添加到收藏
            updatedFavoriteRecords.push(record);
        }

        setFavoriteRecords(updatedFavoriteRecords);
        updateLocalStorage();
    };

    // 发送记录的API请求并保存到 localStorage
    const handleSave = () => {
        if (!selectedRecord) return;

        // 检查记录是否已存在
        const existingIndex = allRecords.findIndex(record => record === selectedRecord);
        if (existingIndex !== -1) {
            // 如果记录已存在，为对应元素添加闪亮动画
            const element = document.querySelector(`[data-record-index="${existingIndex}"]`);
            if (element) {
                element.classList.add('highlight-animation');
                // 动画结束后移除类名
                setTimeout(() => {
                    element.classList.remove('highlight-animation');
                }, 1000);
            }
            return;
        }

        // 如果记录不存在，则添加到历史记录中
        const updatedAllRecords = [...allRecords, selectedRecord];
        setAllRecords(updatedAllRecords);
        updateLocalStorage();
    };

    const handleNameEdit = (record: string, index: number) => {
        setEditingIndex(index);
        setEditingName(customNames[record] || '');
    };

    const handleNameSubmit = (record: string) => {
        if (editingName.trim()) {
            setCustomNames(prev => ({
                ...prev,
                [record]: editingName.trim()
            }));
        }
        setEditingIndex(null);
        setEditingName('');
        updateLocalStorage();
    };

    const renderRecord = (record: string, index: number, isFavorites: boolean = false) => (
        <List.Item key={index} data-record-index={index} className="list-item">
            <div style={{ display: 'flex', justifyContent: 'space-between', width: '100%', position: 'relative', backgroundColor: '#FFFFFF' }}>
                {editingIndex === index ? (
                    <Input
                        autoFocus
                        value={editingName}
                        onChange={setEditingName}
                        onBlur={() => handleNameSubmit(record)}
                        onPressEnter={() => handleNameSubmit(record)}
                        style={{ maxWidth: '120px' }}
                    />
                ) : (
                    <div
                        onClick={(e) => {
                            e.stopPropagation();
                            handleNameEdit(record, index);
                        }}
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
                        {customNames[record] || record}
                    </div>
                )}
                <div style={{ position: 'absolute', right: '4px', top: '50%', transform: 'translateY(-50%)', zIndex: 1 }}>
                    {favoriteRecords.includes(record) && (
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
                            icon={<IconStar style={{ color: favoriteRecords.includes(record) ? '#FFB400' : '#86909C', fill: favoriteRecords.includes(record) ? '#FFB400' : 'none', fontSize: '14px' }} />}
                            onClick={() => {
                                if (favoriteRecords.includes(record)) {
                                    Modal.confirm({
                                        title: '确认取消收藏',
                                        content: '确定要取消收藏这条记录吗？',
                                        onOk: () => toggleFavorite(record)
                                    });
                                } else {
                                    toggleFavorite(record);
                                }
                            }}
                        />
                        <Button
                            type="text"
                            icon={<IconDelete style={{ fontSize: '14px' }} />}
                            onClick={() => {
                                const updatedAllRecords = allRecords.filter(item => item !== record);
                                setAllRecords(updatedAllRecords);
                                if (favoriteRecords.includes(record)) {
                                    const updatedFavoriteRecords = favoriteRecords.filter(item => item !== record);
                                    setFavoriteRecords(updatedFavoriteRecords);
                                }
                                const newCustomNames = { ...customNames };
                                delete newCustomNames[record];
                                setCustomNames(newCustomNames);
                                updateLocalStorage();
                            }}
                        />
                        <Button
                            type="text"
                            icon={<IconExport style={{ fontSize: '14px' }} />}
                            onClick={() => setSelectedRecord(record)}
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
                            dataSource={allRecords.slice((currentPage - 1) * pageSize, currentPage * pageSize)}
                            render={(record, index) => renderRecord(record, index)}
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
                            render={(record, index) => renderRecord(record, index, true)}
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