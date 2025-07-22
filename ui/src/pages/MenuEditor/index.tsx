import React, { useState } from 'react';
import { Tree, Button, Form, Input, Select, InputNumber, message, Modal } from 'antd'; // 移除 Modal 导入
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import type { DataNode } from 'antd/es/tree';

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
        icon: 'home',
        order: 1,
        children: [
            {
                key: '1-1',
                title: '仪表盘',
                icon: 'dashboard',
                order: 1,
            },
            {
                key: '1-2',
                title: '项目管理',
                icon: 'project',
                order: 2,
            },
            {
                key: '1-3',
                title: '设置',
                icon: 'setting',
                order: 3,
                children: [
                    {
                        key: '1-3-1',
                        title: '用户设置',
                        icon: 'user',
                        order: 1,
                    },
                    {
                        key: '1-3-2',
                        title: '权限设置',
                        icon: 'safety',
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
    // 移除 modalForm、modalVisible 状态
    const [editMode, setEditMode] = useState<'add' | 'edit' | null>(null); // 修改为可空类型
    const [parentKey, setParentKey] = useState<string | null>(null);

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

    // 递归添加菜单项，新增后自动展开父节点
    const addMenuItem = (data: MenuItem[], parentKey: string | null, newItem: MenuItem): MenuItem[] => {
        if (!parentKey) {
            return [...data, newItem];
        }
        return data.map(item => {
            if (item.key === parentKey) {
                // 新增后自动展开父节点
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
            if (item) {
                form.setFieldsValue(item);
                setEditMode('edit'); // 选中时切换到编辑模式
            }
        } else {
            setSelectedKey(null);
            form.resetFields();
            setEditMode(null); // 取消选择时重置模式
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
                item.children.unshift(dragObj!);
            });
        } else if (
            (info.node.children || []).length > 0 && info.node.expanded && dropPosition === 1
        ) {
            // 拖到有子节点的节点底部
            loop(data, dropKey, (item) => {
                item.children = item.children || [];
                item.children.unshift(dragObj!);
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
        }
        setMenuData(data);
    };

    // 新增菜单项 - 修改为直接操作右侧表单
    const handleAdd = (parentKey: string | null = null) => {
        setEditMode('add');
        setParentKey(parentKey);
        setSelectedKey(null); // 清除选中状态
        form.resetFields();
        form.setFieldsValue({ title: '', icon: '', url: '', eventType: 'url', order: 1 });
    };

    // 编辑菜单项 - 修改为直接操作右侧表单
    const handleEdit = (key?: string) => {
        const editKey = key || selectedKey;
        if (!editKey) return;
        setEditMode('edit');
        setParentKey(null);
        setSelectedKey(editKey);
        const item = findMenuItem(menuData, editKey);
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

    // 保存菜单项 - 修改为处理右侧表单
    const handleSave = () => {
        form.validateFields().then(values => {
            if (editMode === 'add') {
                const newKey = Date.now().toString();
                const newItem: MenuItem = { ...values, key: newKey };
                setMenuData(prev => addMenuItem(prev, parentKey, newItem));
                message.success('添加成功');
                setSelectedKey(newKey);
                setEditMode('edit'); // 新增后切换到编辑模式
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
                    {item.icon ? <span style={{ marginRight: 4 }}><i className={`anticon anticon-${item.icon}`} /></span> : null}
                    {item.title}
                    <Button size="small" type="link" icon={<EditOutlined />} onClick={e => { e.stopPropagation(); handleEdit(item.key); }}>编辑</Button>
                    <Button size="small" type="link" icon={<DeleteOutlined />} danger onClick={e => { e.stopPropagation(); handleDelete(item.key); }}>删除</Button>
                    <Button size="small" type="link" icon={<PlusOutlined />} onClick={e => { e.stopPropagation(); handleAdd(item.key); }}>新增</Button>
                </span>
            ),
            children: item.children ? convertToTreeData(item.children) : undefined,
        }));
    };

    return (
        <div style={{ display: 'flex', height: '80vh', border: '1px solid #eee', borderRadius: 8, overflow: 'hidden' }}>
            {/* 左侧菜单树 */}
            <div style={{ width: 350, borderRight: '1px solid #eee', padding: 16, overflow: 'auto' }}>
                <div style={{ marginBottom: 16, fontWeight: 'bold', fontSize: 18 }}>菜单树</div>
                <Button type="primary" icon={<PlusOutlined />} onClick={() => handleAdd(null)} style={{ marginBottom: 12 }}>新增根菜单</Button>
                <Tree
                    treeData={convertToTreeData(menuData)}
                    defaultExpandAll
                    selectedKeys={selectedKey ? [selectedKey] : []}
                    onSelect={onSelect}
                    draggable
                    onDrop={onDrop}
                    blockNode
                />
            </div>
            {/* 右侧表单 */}
            <div style={{ flex: 1, padding: 32 }}>
                <div style={{ fontWeight: 'bold', fontSize: 18, marginBottom: 16 }}>
                    {editMode === 'add' ? '新增菜单项' : '菜单项编辑'}
                </div>
                {editMode || selectedKey ? (
                    <> // 有编辑模式或选中项时显示表单
                        <Form
                            form={form}
                            layout="vertical"
                            initialValues={{ title: '', eventType: 'url', order: 1 }} // 添加title初始值
                        >
                            <Form.Item label="菜单名称" name="title" rules={[
                                { required: true, whitespace: true, message: '请输入菜单名称' } // 添加whitespace验证
                            ]}>
                                <Input />
                            </Form.Item>
                            <Form.Item label="图标" name="icon"> <Input placeholder="如 home, dashboard, setting..." /> </Form.Item>
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
            {/* 移除弹窗组件 */}
        </div>
    );
};

export default MenuEditor;
