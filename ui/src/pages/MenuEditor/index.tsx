import React, { useEffect, useState } from 'react';
import { Button, Form, Input, InputNumber, message, Modal, Select, Space, Tag, Tabs, Tree } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import type { DataNode } from 'antd/es/tree';

import IconPicker from '@/components/IconPicker';
import { MenuItem } from '@/types/menu';
import CustomEventTags from './CustomEventTags';
import MenuPreviewTree from './MenuPreviewTree';

const initialMenu: MenuItem[] = [
    {
        key: '1',
        title: '首页',
        icon: 'fa-home',
        order: 1,
        children: [
            {
                key: '1-1',
                title: '仪表盘',
                icon: 'fa-tachometer-alt',
                order: 1,
            },
            {
                key: '1-2',
                title: '项目管理',
                icon: 'fa-project-diagram',
                order: 2,
            },
            {
                key: '1-3',
                title: '设置',
                icon: 'fa-cog',
                order: 3,
                children: [
                    {
                        key: '1-3-1',
                        title: '用户设置',
                        icon: 'fa-user',
                        eventType: 'url',
                        url: 'http://www.baidu.com',
                        order: 1,
                    },
                    {
                        key: '1-3-2',
                        title: '权限设置',
                        icon: 'fa-shield-alt',
                        order: 2,
                    },
                ],
            },
        ],
    },
];

const MenuEditor: React.FC = () => {
    const [menuData, setMenuData] = useState<MenuItem[]>(initialMenu);
    const [history, setHistory] = useState<{ data: MenuItem[], time: string }[]>([]);
    const [historyIndex, setHistoryIndex] = useState(-1);
    const [showHistory, setShowHistory] = useState(false);
    const [selectedKey, setSelectedKey] = useState<string | null>(null);
    const [form] = Form.useForm();

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

    const handleMenuClick = (key: string) => {
        const item = findMenuItem(menuData, key);
        if (item) {
            if (item.eventType === 'url' && item.url) {
                window.open(item.url, '_blank');
            } else if (item.eventType === 'custom' && item.customEvent) {
                try {
                    // 创建一个函数执行上下文
                    const context = {
                        onMenuClick: (path: string) => {
                            // 这里实现onMenuClick的逻辑
                            console.log(`执行自定义菜单点击: ${path}`);
                            // 可以根据需要添加路由跳转或其他逻辑
                        },
                        message
                    };

                    // 构建并执行自定义函数
                    const func = new Function(...Object.keys(context), `return ${item.customEvent}`);
                    const result = func(...Object.values(context));

                    // 如果是函数，执行它
                    if (typeof result === 'function') {
                        result();
                    }
                } catch (error) {
                    console.error('自定义事件执行错误:', error);
                    message.error('自定义事件执行错误');
                }
            }
        }
    };

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
            loop(data, dropKey, (item, idx, arr) => {
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
            title: '',
            icon: '',
            url: '',
            eventType: 'url',
            order: 1,
            customEvent: ''
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
                title: item.title,
                icon: item.icon || '',
                url: item.url || '',
                eventType: item.eventType || 'url',
                customEvent: item.customEvent || '',
                order: item.order || 1
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
                const newKey = Date.now().toString();
                const newItem: MenuItem = { ...values, key: newKey };
                const newData = addMenuItem(menuData, parentKey, newItem);
                setMenuData(newData);
                saveHistory(newData);
                message.success('添加成功');
                setSelectedKey(newKey);
            } else if (editMode === 'edit' && selectedKey) {
                const existingItem = findMenuItem(menuData, selectedKey);
                const newItem = { ...existingItem, ...values, key: selectedKey };
                const newData = updateMenuItem(menuData, selectedKey, newItem);
                setMenuData(newData);
                saveHistory(newData);
                message.success('保存成功');
            }
            // 关闭Modal并重置表单
            setEditMode(null);
            form.resetFields();
            // 输出最终菜单JSON
            console.log("Final Menu JSON:", JSON.stringify(menuData, null, 2));
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
                    <Button size="small" type="link" icon={<EditOutlined />} onClick={e => { e.stopPropagation(); handleEdit(item.key); }} title="编辑" />
                    <Button size="small" type="link" icon={<DeleteOutlined />} danger onClick={e => { e.stopPropagation(); handleDelete(item.key); }} title="删除" />
                    <Button size="small" type="link" icon={<PlusOutlined />} onClick={e => { e.stopPropagation(); handleAdd(item.key); }} title="新增" />
                </span>
            ),
            children: item.children ? convertToTreeData(item.children) : undefined,
            isLeaf: !item.children || item.children.length === 0
        }));
    };

    // 保存历史记录
    const saveHistory = (data: MenuItem[]) => {
        const newHistory = [...history];
        // 如果当前不是最新历史记录，则丢弃后面的记录
        if (historyIndex < newHistory.length - 1) {
            newHistory.splice(historyIndex + 1);
        }
        newHistory.push({
            data: JSON.parse(JSON.stringify(data)), // 深拷贝
            time: new Date().toLocaleString()
        });
        setHistory(newHistory);
        setHistoryIndex(newHistory.length - 1);
    };

    // 恢复历史记录
    const restoreHistory = (index: number) => {
        if (index >= 0 && index < history.length) {
            setMenuData(JSON.parse(JSON.stringify(history[index].data)));
            setHistoryIndex(index);
        }
    };

    return (
        <>
            <div style={{ display: 'flex', height: '80vh', border: '1px solid #eee', borderRadius: 8, overflow: 'hidden' }}>
                {/* 左侧菜单树 */}
                <div style={{ width: 350, borderRight: '1px solid #eee', padding: 16, overflow: 'auto' }}>
                    <div style={{ marginBottom: 16, fontWeight: 'bold', fontSize: 18 }}>
                        菜单树
                        <Button
                            type={showHistory ? "primary" : "default"}
                            onClick={() => setShowHistory(!showHistory)}
                            style={{ marginLeft: 8, float: 'right' }}
                        >
                            历史记录
                        </Button>
                        <Button
                            type={isPreview ? "primary" : "default"}
                            onClick={() => setIsPreview(!isPreview)}
                            style={{ marginLeft: 8, float: 'right' }}
                        >
                            {isPreview ? "返回编辑" : "预览"}
                        </Button>
                    </div>
                    {!isPreview && (
                        <Button
                            type="primary"
                            icon={<PlusOutlined />}
                            onClick={() => handleAdd(null)}
                            style={{ marginBottom: 12 }}
                        >
                            新增根菜单
                        </Button>
                    )}
                    {isPreview ? (
                        <MenuPreviewTree menuData={menuData} onMenuClick={handleMenuClick} />
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
                                    <th style={{ padding: '8px', border: '1px solid #ddd' }}>时间</th>
                                    <th style={{ padding: '8px', border: '1px solid #ddd' }}>操作</th>
                                </tr>
                            </thead>
                            <tbody>
                                {[...history].reverse().map((record, index) => (
                                    <tr key={index} style={{ borderBottom: '1px solid #ddd' }}>
                                        <td style={{ padding: '8px', border: '1px solid #ddd' }}>{record.time}</td>
                                        <td style={{ padding: '8px', border: '1px solid #ddd' }}>
                                            <Button
                                                type="link"
                                                onClick={() => {
                                                    restoreHistory(index);
                                                    setShowHistory(false);
                                                }}
                                            >
                                                恢复到此版本
                                            </Button>
                                            <Button
                                                type="link"
                                                onClick={() => {
                                                    Modal.info({
                                                        title: '菜单预览',
                                                        content: (
                                                            <div>
                                                                <Tabs defaultActiveKey="1">
                                                                    <Tabs.TabPane tab="菜单JSON配置" key="1">
                                                                        <pre style={{ maxHeight: '400px', overflow: 'auto' }}>
                                                                            {JSON.stringify(record.data, null, 2)}
                                                                        </pre>
                                                                    </Tabs.TabPane>
                                                                    <Tabs.TabPane tab="菜单预览" key="2">
                                                                        <MenuPreviewTree menuData={record.data} onMenuClick={handleMenuClick} />
                                                                    </Tabs.TabPane>
                                                                </Tabs>
                                                            </div>
                                                        ),
                                                        width: 800,
                                                    });
                                                }}
                                            >
                                                预览此版本
                                            </Button>
                                        </td>
                                    </tr>
                                ))}
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
                        <Button key="cancel" onClick={() => {
                            setEditMode(null);
                            setSelectedKey(null);
                            form.resetFields();
                        }}>
                            取消
                        </Button>,
                        <Button key="submit" type="primary" onClick={handleSave}>
                            保存
                        </Button>
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
                    </Form>
                </Modal>
            </div>

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

