import {useState, useEffect, useRef} from 'react';
import {Tabs, Button, List, Input, Modal, Space} from 'antd';
import * as monaco from 'monaco-editor';
import React from 'react';
import JSZip from 'jszip';
import {saveAs} from 'file-saver';
import {fetcher} from "@/components/Amis/fetcher.ts";
import {DeleteFilled, EditFilled, StarFilled, StarOutlined} from '@ant-design/icons';
import type {TabsProps} from 'antd';

interface RecordItem {
    id: string;
    content: string;
    isFavorite: boolean;
    customName?: string;
}

interface HistoryRecordsProps {
    data: any;
}

// 用 forwardRef 让组件兼容 AMIS
const HistoryRecordsComponent = React.forwardRef<HTMLDivElement, HistoryRecordsProps>(() => {
    // 初始化记录数据
    const [historyRecords, setHistoryRecords] = useState<RecordItem[]>([]);
    const [favoriteRecords, setFavoriteRecords] = useState<RecordItem[]>([]);
    const [selectedRecord, setSelectedRecord] = useState<string>('');
    const [currentPage, setCurrentPage] = useState(1);
    const [currentFavoritePage, setCurrentFavoritePage] = useState(1);
    const [editingId, setEditingId] = useState<string>();
    const [editingName, setEditingName] = useState('');
    const [activeTab, setActiveTab] = useState('history');
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
                theme: 'vs',
                language: "yaml",
                wordWrap: "on",
                scrollbar: {
                    vertical: "auto"
                },
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
            // 添加到收藏记录的最前面
            setFavoriteRecords(prevRecords => [{...record, isFavorite: true}, ...prevRecords]);
            updateLocalStorage();
        } else {
            // 从收藏记录中移除
            const favoriteRecord = favoriteRecords.find(r => r.id === recordId);
            if (favoriteRecord) {
                Modal.confirm({
                    title: '确认取消收藏',
                    content: '确定要取消收藏这条记录吗？',
                    onOk: () => {
                        setFavoriteRecords(prevRecords => prevRecords.filter(r => r.id !== recordId));
                        // 生成新的ID并添加到历史记录
                        setHistoryRecords(prevRecords => [...prevRecords, {
                            ...favoriteRecord,
                            id: Math.random().toString(36).substring(2, 15),
                            isFavorite: false
                        }]);
                        updateLocalStorage();
                    }
                });
            }
        }
    };

    // 保存记录到 localStorage并发送到服务器
    const handleSave = async () => {
        if (!editorValue) return;

        try {
            const response = await fetcher({
                url: '/k8s/yaml/apply',
                method: 'post',
                data: {
                    yaml: editorValue
                }
            });
            const responseData = response.data;
            if (responseData?.status !== 0) {
                if (response.data?.msg.includes("please apply your changes to the latest version and try again")) {
                    Modal.error({
                        title: '应用失败',
                        content: '资源已被更新，请刷新后再试。'
                    });
                    return;
                } else {
                    Modal.error({
                        title: '应用失败',
                        content: `请尝试刷新后重试。${response.data?.msg}`
                    });
                    return;
                }
            }

            // 解析结果并展示详细信息
            //@ts-ignore
            const resultList = responseData.data.result || [];
            Modal.success({
                title: '应用状态',
                content: (
                    <List
                        style={{maxHeight: '400px', overflow: 'auto'}}
                        dataSource={resultList}
                        renderItem={(item, index) => {
                            const resultItem = item as string;
                            return (
                                <List.Item key={index} style={{padding: '8px'}}>
                                    <div style={{display: 'flex', alignItems: 'center', gap: '8px'}}>
                                        <div>{resultItem}</div>
                                    </div>
                                </List.Item>
                            );
                        }}
                    />
                )
            });

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
            setHistoryRecords(prevRecords => [newRecord, ...prevRecords]);
            updateLocalStorage();
            setActiveTab('history');
        } catch (error) {
            Modal.error({
                title: '应用失败',
                content: error instanceof Error ? error.message : '未知错误'
            });
        }
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
                        ? {...record, customName: editingName.trim()}
                        : record
                )
            );
            // 同步更新收藏列表中的记录名称
            setFavoriteRecords(prevRecords =>
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
            }
        });
    };

    const renderRecord = (record: RecordItem) => (
        <List.Item key={record.id} data-record-id={record.id} className="list-item" style={{cursor: 'pointer'}}
                   onClick={() => setSelectedRecord(record.content)}>
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
                    <div
                        style={{
                            flex: 1,
                            overflow: 'hidden',
                            textOverflow: 'ellipsis',
                            whiteSpace: 'nowrap',
                            marginRight: '10px'
                        }}
                    >
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
                            icon={activeTab === 'favorites' ? <StarFilled style={{color: '#FFD700'}}/> :
                                <StarOutlined/>}
                            onClick={(e) => {
                                e.stopPropagation();
                                toggleFavorite(record.id);
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
    const items: TabsProps['items'] = [
        {
            key: 'history',
            id: 'history',
            label: "历史记录",
            children: [
                <div key={'history-list'}>
                    <div style={{marginBottom: '10px'}}>
                        <Space.Compact>
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
                        </Space.Compact>
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
            ]
        },
        {
            key: 'favorites',
            id: 'favorites',
            label: "收藏",
            children: [
                <div key={'history-list'}>
                    <div style={{marginBottom: '10px'}}>
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
                                    const blob = await zip.generateAsync({type: 'blob'});
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
                    <div style={{marginTop: '16px', textAlign: 'right'}}>
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

            ]
        }
    ];
    return (
        <div style={{display: 'flex', height: '100vh', backgroundColor: '#FFFFFF'}}>
            <div style={{width: '350px', padding: '10px', backgroundColor: '#FFFFFF', flexShrink: 0}}>
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
                <Tabs items={items} onChange={setActiveTab}/>
                <div className="bg-blue-50	" style={{padding: '5px', margin: '5px', fontSize: '12px'}}>
                    历史记录存储在浏览器本地，清空浏览器缓存后会删除所有历史记录
                </div>
            </div>

            <div style={{padding: '10px', backgroundColor: '#FFFFFF', width: '100%'}}>
                <div style={{marginBottom: '10px', display: 'flex', gap: '10px'}}>
                    <Button
                        variant='outlined'
                        onClick={() => {
                            const input = document.createElement('input');
                            input.type = 'file';
                            input.accept = '.yaml,.yml';
                            input.onchange = async (e) => {
                                const file = (e.target as HTMLInputElement).files?.[0];
                                if (file) {
                                    try {
                                        const reader = new FileReader();
                                        reader.onload = (e) => {
                                            const content = e.target?.result as string;
                                            if (monacoInstance.current) {
                                                monacoInstance.current.setValue(content);
                                                setEditorValue(content);
                                            }
                                        };
                                        reader.readAsText(file);
                                    } catch (error) {
                                        Modal.error({
                                            title: '导入失败',
                                            content: '无法读取YAML文件'
                                        });
                                    }
                                }
                            };
                            input.click();
                        }}
                    >
                        导入YAML
                    </Button>
                </div>
                <div ref={editorRef} style={{
                    minWidth: '500px',
                    height: 'calc(100vh - 200px)',
                    border: '1px solid #e5e6eb',
                    borderRadius: '4px'
                }}/>
                <Button
                    type="primary"
                    style={{marginTop: '10px'}}
                    onClick={handleSave}
                >
                    应用到集群
                </Button>
                <Button
                    type="primary"
                    danger={true}
                    style={{marginTop: '10px', marginLeft: '10px'}}
                    onClick={async () => {
                        if (!editorValue) return;
                        try {
                            const response = await fetcher({
                                url: '/k8s/yaml/delete',
                                method: 'post',
                                data: {
                                    yaml: editorValue
                                }
                            });
                            const responseData = response.data;
                            if (responseData?.status !== 0) {
                                Modal.error({
                                    title: '删除失败',
                                    content: `请尝试刷新后重试。${response.data?.msg}`
                                });
                                return;
                            }

                            // 解析结果并展示详细信息
                            //@ts-ignore
                            const resultList = responseData.data.result || [];
                            Modal.success({
                                title: '删除状态',
                                content: (
                                    <List
                                        style={{maxHeight: '400px', overflow: 'auto'}}
                                        dataSource={resultList}
                                        renderItem={(item, index) => {
                                            const resultItem = item as string;
                                            return (
                                                <List.Item key={index} style={{padding: '8px'}}>
                                                    <div style={{display: 'flex', alignItems: 'center', gap: '8px'}}>
                                                        <div>{resultItem}</div>
                                                    </div>
                                                </List.Item>
                                            );
                                        }}
                                    />
                                )
                            });

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
                            setHistoryRecords(prevRecords => [newRecord, ...prevRecords]);
                            updateLocalStorage();
                            setActiveTab('history');
                        } catch (error) {
                            Modal.error({
                                title: '删除失败',
                                content: error instanceof Error ? error.message : '未知错误'
                            });
                        }
                    }}
                >
                    从集群删除
                </Button>
            </div>
        </div>
    );
});

export default HistoryRecordsComponent;
