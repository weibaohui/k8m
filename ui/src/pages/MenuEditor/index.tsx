import React, { useState } from 'react';
import { Button, Form, Input, InputNumber, message, Modal, Select, Tree } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import type { DataNode } from 'antd/es/tree';

import IconPicker from '@/components/IconPicker';

interface MenuItem {
    key: string;
    title: string;
    icon?: string;
    url?: string;
    eventType?: 'url' | 'custom';
    order?: number;
    children?: MenuItem[];
}

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
    const [selectedKey, setSelectedKey] = useState<string | null>(null);
    const [form] = Form.useForm();
    const [editMode, setEditMode] = useState<'add' | 'edit' | null>(null);
    const [parentKey, setParentKey] = useState<string | null>(null);
    const [showIconModal, setShowIconModal] = useState(false);
    const [isPreview, setIsPreview] = useState(false);

    // 处理菜单项点击
    const handleMenuClick = (key: string) => {
        const item = findMenuItem(menuData, key);
        if (item) {
            if (item.eventType === 'url' && item.url) {
                window.open(item.url, '_blank');
            } else if (item.eventType === 'custom') {
                message.info(`触发自定义事件: ${item.title}`);
            }
        }
    };

    // 处理图标选择
    const handleIconSelect = (iconValue: string) => {
        form.setFieldsValue({ icon: iconValue });
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
        if (selectedKeys.length > 0) {
            setSelectedKey(selectedKeys[0] as string);
            const item = findMenuItem(menuData, selectedKeys[0] as string);
            console.log("onSelect:", item);
            if (item) {
                form.setFieldsValue(item);
                setEditMode('edit');
            }
        } else {
            setSelectedKey(null);
            form.resetFields();
            setEditMode(null);
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
                message.info(`已添加为 ${item.title} 的子菜单`);
            });
        } else if (
            (info.node.children || []).length > 0 && info.node.expanded && dropPosition === 1
        ) {
            // 拖到有子节点的节点底部
            loop(data, dropKey, (item) => {
                item.children = item.children || [];
                item.children.push(dragObj!);
                message.info(`已添加为 ${item.title} 的子菜单`);
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
            message.info(`已添加为 ${info.node.title.props.children[1]} 的同级菜单`);
        }
        setMenuData(data);
    };

    // 新增菜单项
    const handleAdd = (parentKey: string | null = null) => {
        setEditMode('add');
        setParentKey(parentKey);
        setSelectedKey(null);
        form.resetFields();
        form.setFieldsValue({ title: '', icon: '', url: '', eventType: 'url', order: 1 });
    };

    // 编辑菜单项
    const handleEdit = (key?: string) => {
        const editKey = key || selectedKey;
        if (!editKey) return;
        setEditMode('edit');
        setParentKey(null);
        setSelectedKey(editKey);
        const item = findMenuItem(menuData, editKey);
        console.log("handleEdit:", item);
        if (item) {
            form.setFieldsValue(item);
        }
    };

    // 删除菜单项
    const handleDelete = (key?: string) => {
        const delKey = key || selectedKey;
        if (!delKey) return;
        Modal.confirm({
            title: '确认删除该菜单项？',
            onOk: () => {
                setMenuData(deleteMenuItem(menuData, delKey));
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
                setMenuData(prev => addMenuItem(prev, parentKey, newItem));
                message.success('添加成功');
                setSelectedKey(newKey);
                setEditMode('edit');
            } else if (editMode === 'edit' && selectedKey) {
                const existingItem = findMenuItem(menuData, selectedKey);
                const newItem = { ...existingItem, ...values, key: selectedKey };
                setMenuData(updateMenuItem(menuData, selectedKey, newItem));
                message.success('保存成功');
            }
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

    return (
        <div style={{ display: 'flex', height: '80vh', border: '1px solid #eee', borderRadius: 8, overflow: 'hidden' }}>
            {/* 左侧菜单树 */}
            <div style={{ width: 350, borderRight: '1px solid #eee', padding: 16, overflow: 'auto' }}>
                <div style={{ marginBottom: 16, fontWeight: 'bold', fontSize: 18 }}>
                    菜单树
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
                    <Tree
                        treeData={menuData.map(item => ({
                            key: item.key,
                            title: (
                                <span onClick={() => handleMenuClick(item.key)}>
                                    {item.icon && <i className={`fa-solid ${item.icon}`} style={{ marginRight: '4px' }}></i>}
                                    {item.title}
                                </span>
                            ),
                            children: item.children?.map(child => ({
                                key: child.key,
                                title: (
                                    <span onClick={() => handleMenuClick(child.key)}>
                                        {child.icon && <i className={`fa-solid ${child.icon}`} style={{ marginRight: '4px' }}></i>}
                                        {child.title}
                                    </span>
                                ),
                                children: child.children?.map(grandChild => ({
                                    key: grandChild.key,
                                    title: (
                                        <span onClick={() => handleMenuClick(grandChild.key)}>
                                            {grandChild.icon && <i className={`fa-solid ${grandChild.icon}`} style={{ marginRight: '4px' }}></i>}
                                            {grandChild.title}
                                        </span>
                                    )
                                }))
                            }))
                        }))}
                        defaultExpandAll
                        showLine
                        blockNode
                    />
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
            {/* 右侧表单 */}
            <div style={{ flex: 1, padding: 32, display: isPreview ? 'none' : 'block' }}>
                <div style={{ fontWeight: 'bold', fontSize: 18, marginBottom: 16 }}>
                    {editMode === 'add' ? '新增菜单项' : '菜单项编辑'}
                </div>
                {editMode || selectedKey ? (
                    <>
                        <Form
                            form={form}
                            layout="vertical"
                            // 移除静态initialValues
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
                                    {form.getFieldValue('icon') ? (
                                        <span style={{ display: 'flex', alignItems: 'center' }}>
                                            <i className={`fa-solid ${form.getFieldValue('icon')}`} style={{ marginRight: '8px' }}></i>
                                        </span>
                                    ) : (
                                        '选择图标'
                                    )}
                                </Button>
                            </Form.Item>
                            <Form.Item label="URL" name="url"> <Input /> </Form.Item>
                            <Form.Item label="点击事件" name="eventType"> <Select options={[{ label: 'url跳转', value: 'url' }, { label: '自定义', value: 'custom' }]} /> </Form.Item>
                            <Form.Item label="排序" name="order"> <InputNumber min={1} /> </Form.Item>
                        </Form>
                        <Button type="primary" onClick={handleSave} style={{ marginRight: 8 }}>保存</Button>
                        <Button onClick={() => {
                            setSelectedKey(null);
                            form.resetFields();
                            setEditMode(null);
                        }}>取消</Button>
                    </>
                ) : (
                    <div style={{ color: '#aaa', marginTop: 32 }}>请选择左侧菜单项进行编辑或点击"新增"按钮创建新菜单项</div>
                )}
            </div>

            <IconPicker
                visible={showIconModal}
                onCancel={() => setShowIconModal(false)}
                onSelect={handleIconSelect}
                selectedIcon={form.getFieldValue('icon')}
            />

        </div>
    );
};

export default MenuEditor;
