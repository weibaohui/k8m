import React from 'react';
import { Menu, MenuProps } from 'antd';
import {
    CopyOutlined,
    DeleteOutlined,
    DownloadOutlined,
    EditOutlined,
    FileZipOutlined,
    ReloadOutlined,
    UploadOutlined
} from '@ant-design/icons';
import { FileNode } from './FileTree';

type MenuItem = Required<MenuProps>['items'][number];

interface ContextMenuProps {
    visible: boolean;
    x: number;
    y: number;
    node: FileNode | null;
    onMenuClick: MenuProps['onClick'];
}

const ContextMenu: React.FC<ContextMenuProps> = ({
    visible,
    x,
    y,
    node,
    onMenuClick
}) => {
    if (!visible || !node) return null;

    const items: MenuItem[] = [
        {
            label: '刷新',
            key: 'refresh',
            icon: <ReloadOutlined style={{ color: '#1890ff' }} />
        },
        {
            label: '复制路径',
            key: 'copy',
            icon: <CopyOutlined style={{ color: '#1890ff' }} />
        },
        {
            label: '删除',
            key: 'delete',
            icon: <DeleteOutlined style={{ color: '#ff4d4f' }} />
        },
        {
            label: '编辑',
            key: 'edit',
            disabled: !(node?.type === 'file'),
            icon: <EditOutlined style={{ color: '#52c41a' }} />
        },
        {
            label: '下载',
            key: 'download',
            disabled: !(node?.type === 'file'),
            icon: <DownloadOutlined style={{ color: '#722ed1' }} />
        },
        {
            label: '压缩下载',
            key: 'downloadZip',
            disabled: !(node?.type === 'file'),
            icon: <FileZipOutlined style={{ color: '#faad14' }} />
        },
        {
            label: '上传',
            key: 'upload',
            disabled: !node?.isDir,
            icon: <UploadOutlined style={{ color: '#13c2c2' }} />
        }
    ];

    return (
        <Menu
            style={{
                position: 'absolute',
                top: y,
                left: x,
                zIndex: 1000,
                minWidth: '120px',
                backgroundColor: '#ffffff',
                boxShadow: '0 3px 6px -4px rgba(0, 0, 0, 0.12), 0 6px 16px 0 rgba(0, 0, 0, 0.08), 0 9px 28px 8px rgba(0, 0, 0, 0.05)',
                border: '1px solid #f0f0f0',
                borderRadius: '2px'
            }}
            onClick={onMenuClick}
            items={items}
        />
    );
};

export default ContextMenu;