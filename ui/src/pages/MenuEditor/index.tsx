import React, { useEffect, useState } from 'react';
import { Button, Form, Input, InputNumber, message, Modal, Select, Tabs, Tree, Tooltip, Flex, Space } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, CopyOutlined, FileTextOutlined, EyeOutlined, HistoryOutlined, RollbackOutlined, SnippetsOutlined, ImportOutlined } from '@ant-design/icons';
import type { DataNode } from 'antd/es/tree';
import { useNavigate } from 'react-router-dom';

import IconPicker from '@/components/IconPicker';
import { MenuItem } from '@/types/menu';
import { initialMenu } from './menuData'; // 添加这一行导入语句
import CustomEventTags from './CustomEventTags';
import Preview from './Preview.tsx';
import { fetcher } from '@/components/Amis/fetcher';

interface ApiResponse {
    status: number;
    msg?: string;
    data?: {
        status: number;
        msg?: string;
        data: HistoryItem[];
    };
}
interface HistoryItem {
    menu_data: MenuItem[];
    id: number;
    updated_at: string;
    created_at: string;
}
const MenuEditor: React.FC = () => {
    const navigate = useNavigate();
    const [menuData, setMenuData] = useState<MenuItem[]>(initialMenu);
    const [history, setHistory] = useState<HistoryItem[]>([]);
    const [historyIndex, setHistoryIndex] = useState(-1);
    const [showHistory, setShowHistory] = useState(false);
    const [selectedKey, setSelectedKey] = useState<string | null>(null);
    const [form] = Form.useForm();
    const [isImportModalVisible, setIsImportModalVisible] = useState(false);
    const [importJson, setImportJson] = useState('');

    /**
     * 从API加载历史记录
     */
    const loadHistoryFromAPI = async () => {
        try {
            const response = await fetcher({
                url: '/admin/menu/history',
                method: 'get'
            }) as ApiResponse;

            if (response.data?.status !== 0) {
                message.error(`获取巡检结果失败:请尝试刷新后重试。 ${response.data?.msg}`);
            } else {
                const result = response.data.data;
                setHistory(result);
                setHistoryIndex(result.length - 1);
            }

        } catch (error) {
            console.error('从API加载历史记录失败:', error);
        }
    };





    // 组件初始化时加载历史记录
    useEffect(() => {
        loadHistoryFromAPI();
    }, []);

    const [editMode, setEditMode] = useState<'add' | 'edit' | null>(null);
    const [parentKey, setParentKey] = useState<string | null>(null);
    const [showIconModal, setShowIconModal] = useState(false);
    const [isPreview, setIsPreview] = useState(false);
    const [currentEventType, setCurrentEventType] = useState<'url' | 'custom'>('url');
    const [currentIcon, setCurrentIcon] = useState<string | null>(null);


    // eventType 的变化现在直接通过 Select 的 onChange 处理
    useEffect(() => {
        // 当表单初始化时，设置默认值
        const initialEventType = form.getFieldValue('eventType');
        setCurrentEventType(initialEventType || 'url');
        setCurrentIcon(form.getFieldValue('icon') || null);

    }, [form]);


    // 处理图标选择
    const handleIconSelect = (iconValue: string) => {
        setCurrentIcon(iconValue);
        form.setFieldValue('icon', iconValue)
        setShowIconModal(false);
    };

    // 递归查找菜单项
    const findMenuItem = (data: MenuItem[], key: string): MenuItem | null => {
        for (const item of data) {
            if (item.key === key) return item;
            if (item.children) {
                const found = findMenuItem(item.children, key);
                if (found) return found;
            }
        }
        return null;
    };

    // 递归更新菜单项
    const updateMenuItem = (data: MenuItem[], key: string, newItem: MenuItem): MenuItem[] => {
        return data.map(item => {
            if (item.key === key) return { ...newItem };
            if (item.children) {
                return { ...item, children: updateMenuItem(item.children, key, newItem) };
            }
            return item;
        });
    };

    // 递归删除菜单项
    const deleteMenuItem = (data: MenuItem[], key: string): MenuItem[] => {
        return data.filter(item => {
            if (item.key === key) return false;
            if (item.children) {
                item.children = deleteMenuItem(item.children, key);
            }
            return true;
        });
    };

    // 递归添加菜单项
    const addMenuItem = (data: MenuItem[], parentKey: string | null, newItem: MenuItem): MenuItem[] => {
        if (!parentKey) {
            return [...data, newItem];
        }
        return data.map(item => {
            if (item.key === parentKey) {
                return { ...item, children: [...(item.children || []), newItem] };
            }
            if (item.children) {
                return { ...item, children: addMenuItem(item.children, parentKey, newItem) };
            }
            return item;
        });
    };

    // 处理菜单树选择
    const onSelect = (selectedKeys: React.Key[]) => {
        // 仅设置选中项，不触发编辑模式
        if (selectedKeys.length > 0) {
            setSelectedKey(selectedKeys[0] as string);
        } else {
            setSelectedKey(null);
        }
    };

    // 拖拽排序和层级调整
    const onDrop = (info: any) => {
        const dropKey = info.node.key;
        const dragKey = info.dragNode.key;
        const dropPos = info.node.pos.split('-');
        const dropPosition = info.dropPosition - Number(dropPos[dropPos.length - 1]);

        const loop = (data: MenuItem[], key: string, callback: (item: MenuItem, idx: number, arr: MenuItem[]) => void) => {
            for (let i = 0; i < data.length; i++) {
                if (data[i].key === key) {
                    return callback(data[i], i, data);
                }
                if (data[i].children) {
                    loop(data[i].children!, key, callback);
                }
            }
        };

        const data = [...menuData];
        let dragObj: MenuItem;
        loop(data, dragKey, (item, idx, arr) => {
            arr.splice(idx, 1);
            dragObj = item;
        });

        if (!info.dropToGap) {
            // 拖到节点内部
            loop(data, dropKey, (item) => {
                item.children = item.children || [];
                item.children.push(dragObj!);
                message.info(`已经将 ${dragObj!.title} 添加为 ${item.title} 的子菜单`);
                // 更新order值
                item.children!.forEach((child, index) => {
                    child.order = index + 1;
                });
            });
        } else if (
            (info.node.children || []).length > 0 && info.node.expanded && dropPosition === 1
        ) {
            // 拖到有子节点的节点底部
            loop(data, dropKey, (item) => {
                item.children = item.children || [];
                item.children.push(dragObj!);
                message.info(`已经将 ${dragObj!.title} 添加为 ${item.title} 的子菜单`);
                // 更新order值
                item.children!.forEach((child, index) => {
                    child.order = index + 1;
                });
            });
        } else {
            // 拖到节点之间
            let ar: MenuItem[] = data;
            let i: number;
            loop(data, dropKey, (_, idx, arr) => {
                ar = arr;
                i = idx;
            });
            ar.splice(dropPosition === -1 ? i! : i! + 1, 0, dragObj!);
            message.info(`已经将 ${dragObj!.title} 添加为 ${info.node.title.props.children[1]} 的同级菜单`);
            // 更新order值
            ar.forEach((item, index) => {
                item.order = index + 1;
            });
        }
        setMenuData(data);
        saveHistory(data);
    };

    // 新增菜单项
    const handleAdd = (parentKey: string | null = null) => {
        setEditMode('add');
        setParentKey(parentKey);
        setSelectedKey(null);
        form.resetFields();
        form.setFieldsValue({
            key: '', // 新增菜单时key为空，保存时再生成
            title: '',
            icon: '',
            url: '',
            eventType: 'url',
            order: 1,
            customEvent: '',
            show: 'true' // 修改这里，只使用字符串值
        });
    };

    // 编辑菜单项
    const handleEdit = (key?: string) => {
        const editKey = key || selectedKey;
        if (!editKey) return;

        setEditMode('edit');
        setParentKey(null);
        setSelectedKey(editKey);

        const item = findMenuItem(menuData, editKey);
        if (item) {
            form.resetFields();
            const formValues = {
                key: item.key, // 添加key字段
                title: item.title,
                icon: item.icon || '',
                url: item.url || '',
                eventType: item.eventType || 'url',
                customEvent: item.customEvent || '',
                order: item.order || 1,
                show: typeof item.show === 'string' ? item.show : 'true' // 修改这里，只处理字符串
            };
            form.setFieldsValue(formValues);
            setCurrentEventType(formValues.eventType);
            setCurrentIcon(formValues.icon || null);
        }
    };

    // 删除菜单项
    const handleDelete = (key?: string) => {
        const delKey = key || selectedKey;
        if (!delKey) return;
        Modal.confirm({
            title: '确认删除该菜单项？',
            onOk: () => {
                const newData = deleteMenuItem(menuData, delKey);
                setMenuData(newData);
                saveHistory(newData);
                if (delKey === selectedKey) {
                    setSelectedKey(null);
                    form.resetFields();
                }
                message.success('删除成功');
            },
        });
    };

    // 保存菜单项
    const handleSave = () => {
        form.validateFields().then(values => {
            if (editMode === 'add') {
                // 生成唯一key
                const newKey = `menu_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
                const newItem: MenuItem = { ...values, key: newKey };
                const newData = addMenuItem(menuData, parentKey, newItem);
                setMenuData(newData);
                saveHistory(newData);
                message.success('添加成功');
                setSelectedKey(newKey);
            } else if (editMode === 'edit' && selectedKey) {
                const existingItem = findMenuItem(menuData, selectedKey);
                // 保留现有key不变
                const newItem = { ...existingItem, ...values, key: selectedKey };
                const newData = updateMenuItem(menuData, selectedKey, newItem);
                setMenuData(newData);
                saveHistory(newData);
            }
            // 关闭Modal并重置表单
            setEditMode(null);
            form.resetFields();

        });
    };

    // 菜单树数据转换
    const convertToTreeData = (data: MenuItem[]): DataNode[] => {
        return data.map(item => ({
            key: item.key,
            title: (
                <span>
                    {item.icon ? <i className={`fa-solid ${item.icon}`} style={{ marginRight: '4px' }}></i> : null}
                    {item.title}
                    <Tooltip title="编辑">
                        <Button size="small" type="link" icon={<EditOutlined />} onClick={e => {
                            e.stopPropagation();
                            handleEdit(item.key);
                        }} />
                    </Tooltip>
                    <Tooltip title="删除">
                        <Button size="small" type="link" icon={<DeleteOutlined />} danger onClick={e => {
                            e.stopPropagation();
                            handleDelete(item.key);
                        }} />
                    </Tooltip>
                    <Tooltip title="新增">
                        <Button size="small" type="link" icon={<PlusOutlined />} onClick={e => {
                            e.stopPropagation();
                            handleAdd(item.key);
                        }} />
                    </Tooltip>
                </span>
            ),
            children: item.children ? convertToTreeData(item.children) : undefined,
            isLeaf: !item.children || item.children.length === 0
        }));
    };

    /**
     * 保存菜单数据到后端API
     * @param data 菜单数据
     */
    const saveMenuToAPI = async (data: MenuItem[]) => {
        try {
            const response = await fetcher({
                url: '/admin/menu/save',
                method: 'post',
                data: {
                    menu_data: data
                }
            }) as ApiResponse;

            if (response.data?.status === 0) {
                message.success('菜单保存成功');
                return true;
            } else {
                message.error(response.data?.msg || '保存失败');
                return false;
            }
        } catch (error) {
            message.error('保存菜单失败，请检查网络连接');
            return false;
        }
    };

    /**
     * 保存历史记录到本地存储
     * @param data 菜单数据
     */
    const saveHistory = async (data: MenuItem[]) => {
        // 先尝试保存到后端API
        const saveSuccess = await saveMenuToAPI(data);

        if (saveSuccess) {
            const newHistory = [...history];
            // 如果当前不是最新历史记录，则丢弃后面的记录
            if (historyIndex < newHistory.length - 1) {
                newHistory.splice(historyIndex + 1);
            }
            newHistory.push(JSON.parse(JSON.stringify(data))); // 深拷贝

            // 限制历史记录数量，避免localStorage过大
            const maxHistoryCount = 50;
            if (newHistory.length > maxHistoryCount) {
                newHistory.splice(0, newHistory.length - maxHistoryCount);
            }

            setHistory(newHistory);
            setHistoryIndex(newHistory.length - 1);

            // 输出最终菜单JSON
            console.log("Final Menu JSON:", JSON.stringify(data, null, 2));
        }
    };

    // 恢复历史记录
    const restoreHistory = (index: number) => {
        if (index >= 0 && index < history.length) {
            setMenuData(JSON.parse(JSON.stringify(history[index])));
            setHistoryIndex(index);
        }
    };

    /**
     * 删除历史记录
     * @param id 历史记录ID
     */
    const handleDeleteHistory = async (id: number) => {
        Modal.confirm({
            title: '确认删除此历史记录？',
            onOk: async () => {
                try {
                    const response = await fetcher({
                        url: `/admin/menu/history/delete/${id}`,
                        method: 'delete'
                    }) as ApiResponse;

                    if (response.status === 0) {
                        message.success('历史记录删除成功');
                        loadHistoryFromAPI(); // 重新加载历史记录
                    } else {
                        message.error(response.msg || '删除失败');
                    }
                } catch (error) {
                    console.error('删除历史记录失败:', error);
                    message.error('删除历史记录失败，请检查网络连接');
                }
            },
        });
    };

    /**
     * @description 处理导入菜单的逻辑
     */
    const handleImport = () => {
        try {
            const importedMenu = JSON.parse(importJson);
            // 在这里可以添加对导入的JSON数据格式的校验
            setMenuData(importedMenu);
            saveHistory(importedMenu);
            message.success('菜单导入成功');
            setIsImportModalVisible(false);
            setImportJson('');
        } catch (error) {
            message.error('JSON格式错误，请检查后重试');
        }
    };

    /**
     * 复制历史记录数据到剪贴板
     * @param data 菜单数据
     */
    const handleCopyHistory = (data: MenuItem[]) => {
        navigator.clipboard.writeText(JSON.stringify(data, null, 2))
            .then(() => message.success('菜单数据已复制到剪贴板'))
            .catch(() => message.error('复制失败'));
    };


    return (
        <>
            <div style={{
                display: 'flex',
                height: '100vh',
                border: '1px solid #eee',
                borderRadius: 8,
                overflow: 'hidden'
            }}>
                {/* 左侧菜单树 */}
                <div style={{ width: 350, borderRight: '1px solid #eee', padding: 16, overflow: 'auto' }}>
                    <div style={{
                        marginBottom: 16,
                        fontWeight: 'bold',
                        fontSize: 18,
                        backgroundColor: '#f5f5f5',
                        padding: '8px 12px',
                        borderRadius: '6px',
                        border: '1px solid #e8e8e8',
                        display: 'flex',
                        justifyContent: 'space-between',
                        alignItems: 'center'
                    }}>
                        <span>菜单编辑器</span>
                        <div style={{ display: 'flex', gap: '4px' }}>
                            <Tooltip title="新增菜单项">
                                <Button
                                    type="text"
                                    size="small"
                                    icon={<PlusOutlined />}
                                    onClick={() => handleAdd(null)}
                                />
                            </Tooltip>
                            <Tooltip title={isPreview ? "返回编辑" : "预览菜单"}>
                                <Button
                                    type={isPreview ? "primary" : "text"}
                                    size="small"
                                    icon={<EyeOutlined />}
                                    onClick={() => setIsPreview(!isPreview)}
                                />
                            </Tooltip>
                            <Tooltip title="导入菜单">
                                <Button
                                    type="text"
                                    size="small"
                                    icon={<ImportOutlined />}
                                    onClick={() => setIsImportModalVisible(true)}
                                />
                            </Tooltip>
                            <Tooltip title="复制JSON配置">
                                <Button
                                    type="text"
                                    size="small"
                                    icon={<CopyOutlined />}
                                    onClick={() => {
                                        const jsonString = JSON.stringify(menuData, null, 2);
                                        navigator.clipboard.writeText(jsonString).then(() => {
                                            message.success('JSON配置已复制到剪贴板');
                                        }).catch(() => {
                                            message.error('复制失败，请手动复制');
                                        });
                                    }}
                                />
                            </Tooltip>
                            <Tooltip title="显示JSON配置">
                                <Button
                                    type="text"
                                    size="small"
                                    icon={<FileTextOutlined />}
                                    onClick={() => {
                                        Modal.info({
                                            title: '当前菜单JSON配置',
                                            content: (
                                                <pre style={{ maxHeight: '400px', overflow: 'auto' }}>
                                                    {JSON.stringify(menuData, null, 2)}
                                                </pre>
                                            ),
                                            width: 800,
                                        });
                                    }}
                                />
                            </Tooltip>
                            <Tooltip title="菜单修改历史">
                                <Button
                                    type={showHistory ? "primary" : "text"}
                                    size="small"
                                    icon={<HistoryOutlined />}
                                    onClick={() => {
                                        if (!showHistory) {
                                            // 打开历史面板时重新从API加载历史记录
                                            loadHistoryFromAPI();
                                        }
                                        setShowHistory(!showHistory);
                                    }}
                                />
                            </Tooltip>
                        </div>
                    </div>
                    {isPreview ? (
                        <Preview menuData={menuData} navigate={navigate} />
                    ) : (
                        <Tree
                            treeData={convertToTreeData(menuData)}
                            defaultExpandAll
                            showLine
                            selectedKeys={selectedKey ? [selectedKey] : []}
                            onSelect={onSelect}
                            draggable
                            onDrop={onDrop}
                            blockNode
                        />
                    )}
                </div>
                {/* 右侧使用说明面板 */}
                <div style={{ flex: 1, padding: 32, display: isPreview ? 'none' : 'block' }}>
                    <div style={{ fontWeight: 'bold', fontSize: 18, marginBottom: 16 }}>
                        使用说明
                    </div>
                    <div style={{ color: '#666', lineHeight: '1.8' }}>
                        <h3>基本操作：</h3>
                        <ul>
                            <li>点击"新增根菜单"按钮可以在顶层添加菜单项</li>
                            <li>点击菜单项后的<EditOutlined />图标可以编辑该菜单</li>
                            <li>点击菜单项后的<DeleteOutlined />图标可以删除该菜单</li>
                            <li>点击菜单项后的<PlusOutlined />图标可以添加子菜单</li>
                        </ul>

                        <h3>高级功能：</h3>
                        <ul>
                            <li>支持拖拽排序：直接拖动菜单项可以调整顺序或层级</li>
                            <li>支持两种菜单动作：URL跳转和自定义事件</li>
                            <li>历史记录：点击右上角"历史记录"按钮可以查看和恢复历史版本</li>
                            <li>预览模式：点击右上角"预览"按钮可以预览实际效果</li>
                        </ul>

                        <h3>菜单配置说明：</h3>
                        <ul>
                            <li>图标：支持 Font Awesome 图标</li>
                            <li>URL跳转：直接填写目标URL地址</li>
                            <li>自定义事件：可以编写 JavaScript 代码实现复杂交互</li>
                            <li>排序号：决定同级菜单的显示顺序</li>
                            <li>
                                显示表达式：控制菜单项是否显示，支持使用预定义函数的JavaScript表达式
                                <ul>
                                    <li><code>true</code> 或 <code>false</code>：直接控制显示</li>
                                    <li><code>contains('admin', user.role)</code>：检查用户角色是否包含指定字符串</li>
                                    <li><code>isGatewayAPISupported()==true</code>：检查集群是否支持Gateway API</li>
                                    <li><code>isIstioSupported()==true</code>：检查集群是否支持Istio</li>
                                    <li><code>isOpenKruiseSupported()==true</code>：检查集群是否支持OpenKruise</li>
                                    <li><code>isPlatformAdmin()==true</code>：检查用户是否为平台管理员</li>
                                </ul>
                            </li>
                        </ul>

                    </div>
                </div>

                {/* 历史记录面板 */}
                <Modal
                    title="菜单修改历史"
                    open={showHistory}
                    onCancel={() => setShowHistory(false)}
                    footer={null}
                    width={800}
                >
                    <div style={{ maxHeight: '60vh', overflow: 'auto' }}>
                        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                            <thead>
                                <tr style={{ backgroundColor: '#f0f0f0' }}>
                                    <th style={{ padding: '8px', border: '1px solid #ddd' }}>序号</th>
                                    <th style={{ padding: '8px', border: '1px solid #ddd' }}>时间</th>
                                    <th style={{ padding: '8px', border: '1px solid #ddd' }}>操作</th>
                                </tr>
                            </thead>
                            <tbody>
                                {[...history].map((record, index) => {
                                    const actualIndex = history.length - 1 - index; // 计算实际索引
                                    return (
                                        <tr key={index} style={{ borderBottom: '1px solid #ddd' }}>
                                            <td style={{ padding: '8px', border: '1px solid #ddd' }}>{record.id}</td>
                                            <td style={{ padding: '8px', border: '1px solid #ddd' }}>
                                                {new Date(record.created_at).toLocaleString('zh-CN', {
                                                    year: 'numeric',
                                                    month: '2-digit',
                                                    day: '2-digit',
                                                    hour: '2-digit',
                                                    minute: '2-digit',
                                                    second: '2-digit',
                                                    hour12: false
                                                })}
                                            </td>
                                            <td style={{ padding: '0px', border: '1px solid #ddd' }}>
                                                <Flex wrap gap="small">

                                                    <Button
                                                        icon={<RollbackOutlined />}
                                                        onClick={() => {
                                                            restoreHistory(actualIndex);
                                                            setShowHistory(false);
                                                            message.success('已恢复到选定版本');
                                                        }}
                                                    >
                                                        恢复
                                                    </Button>
                                                    <Button
                                                        danger
                                                        icon={<DeleteOutlined />}
                                                        onClick={() => handleDeleteHistory(record.id)}
                                                    >
                                                        删除
                                                    </Button>
                                                    <Button
                                                        icon={<EyeOutlined />}
                                                        onClick={() => {
                                                            Modal.info({
                                                                title: '菜单预览',
                                                                content: (
                                                                    <div>
                                                                        <Tabs defaultActiveKey="1">
                                                                            <Tabs.TabPane tab="菜单JSON配置" key="1">
                                                                                <pre style={{
                                                                                    maxHeight: '400px',
                                                                                    overflow: 'auto'
                                                                                }}>
                                                                                    {JSON.stringify(record.menu_data, null, 2)}
                                                                                </pre>
                                                                            </Tabs.TabPane>
                                                                            <Tabs.TabPane tab="菜单预览" key="2">
                                                                                <Preview menuData={record.menu_data} />
                                                                            </Tabs.TabPane>
                                                                        </Tabs>
                                                                    </div>
                                                                ),
                                                                width: 800,
                                                            });
                                                        }}
                                                    >
                                                        预览
                                                    </Button>
                                                    <Button
                                                        icon={<CopyOutlined />}
                                                        onClick={() => handleCopyHistory(record.menu_data)}
                                                    >
                                                        复制
                                                    </Button>

                                                </Flex>
                                            </td>
                                        </tr>
                                    );
                                })}
                            </tbody>
                        </table>
                    </div>
                </Modal>

                {/* 编辑表单弹窗 */}
                <Modal
                    title={editMode === 'add' ? '新增菜单项' : '编辑菜单项'}
                    open={!!editMode}
                    onCancel={() => {
                        setEditMode(null);
                        setSelectedKey(null);
                        form.resetFields();
                    }}
                    footer={[
                        <Tooltip key="cancel-tooltip" title="取消编辑">
                            <Button key="cancel" onClick={() => {
                                setEditMode(null);
                                setSelectedKey(null);
                                form.resetFields();
                            }}>
                                取消
                            </Button>
                        </Tooltip>,
                        <Tooltip key="submit-tooltip" title="保存菜单项">
                            <Button key="submit" type="primary" onClick={handleSave}>
                                保存
                            </Button>
                        </Tooltip>
                    ]}
                >
                    <Form
                        form={form}
                        layout="vertical"
                    >

                        <Form.Item label="菜单名称" name="title" rules={[
                            { required: true, whitespace: true, message: '请输入菜单名称' }
                        ]}>
                            <Input />
                        </Form.Item>
                        <Form.Item label="图标" name="icon">
                            <Tooltip title="点击选择图标">
                                <Button
                                    type="primary"
                                    onClick={() => setShowIconModal(true)}
                                    style={{ width: '100%', justifyContent: 'space-between' }}
                                >
                                    {currentIcon ? (
                                        <span style={{ display: 'flex', alignItems: 'center' }}>
                                            <i className={`fa-solid ${currentIcon}`} style={{ marginRight: '8px' }}></i>
                                        </span>
                                    ) : (
                                        '选择图标'
                                    )}
                                </Button>
                            </Tooltip>
                        </Form.Item>

                        <Form.Item label="点击事件" name="eventType" initialValue="url">
                            <Select
                                style={{ zIndex: 1000000 }}
                                options={[
                                    { label: 'URL跳转', value: 'url' },
                                    { label: '自定义', value: 'custom' }
                                ]}
                                onChange={(value) => setCurrentEventType(value as 'url' | 'custom')}
                            />
                        </Form.Item>

                        {currentEventType === 'url' && (
                            <Form.Item label="URL" name="url">
                                <Input />
                            </Form.Item>
                        )}

                        {currentEventType === 'custom' && (
                            <Form.Item label="自定义事件代码" rules={[
                                { required: true, message: '请输入自定义事件代码' }
                            ]}>
                                <>
                                    <CustomEventTags onChange={(value) => form.setFieldsValue({ customEvent: value })} />
                                    <Form.Item noStyle name="customEvent">
                                        <Input.TextArea rows={4} placeholder="请输入自定义事件代码" />
                                    </Form.Item>
                                </>
                            </Form.Item>
                        )}

                        <Form.Item label="排序" name="order">
                            <InputNumber min={1} />
                        </Form.Item>

                        {/* 修改显示控制部分 */}
                        <Form.Item
                            label="显示表达式"
                            name="show"
                            rules={[{ required: true, message: '请输入显示表达式' }]}
                        >
                            <Input.TextArea
                                rows={3}
                                placeholder="请输入JavaScript表达式，例如: true 或 user.role === 'admin'"
                            />
                        </Form.Item>
                    </Form>
                </Modal>
            </div>

            <Modal
                title="导入菜单配置"
                open={isImportModalVisible}
                onOk={handleImport}
                onCancel={() => {
                    setIsImportModalVisible(false);
                    setImportJson('');
                }}
                okText="导入"
                cancelText="取消"
            >
                <Input.TextArea
                    rows={10}
                    value={importJson}
                    onChange={(e) => setImportJson(e.target.value)}
                    placeholder='请在此处粘贴菜单的JSON配置'
                />
            </Modal>

            <IconPicker
                open={showIconModal}
                onCancel={() => setShowIconModal(false)}
                onSelect={handleIconSelect}
                selectedIcon={currentIcon || ''}
            />

        </>
    );
};

export default MenuEditor;

