import { Avatar, Dropdown, Space, MenuProps } from 'antd';
import { useNavigate } from 'react-router-dom';
import styles from './index.module.scss';
import { UserOutlined } from '@ant-design/icons';
import { useEffect, useState } from 'react';
import { jwtDecode } from 'jwt-decode';

interface DecodedToken {
    username: string;
    role: string;
    // JWT标准字段
    exp: number;
    iat: number;
}

const Toolbar = () => {
    const navigate = useNavigate();
    const [userInfo, setUserInfo] = useState<{ username: string; role: string }>({ username: '', role: '' });

    useEffect(() => {
        const token = localStorage.getItem('token');
        if (token) {
            try {
                const decoded = jwtDecode<DecodedToken>(token);


                // 角色映射表
                const roleMap: Record<string, string> = {
                    'cluster_admin': '集群管理员',
                    'cluster_readonly': '集群只读',
                    'platform_admin': '平台管理员'
                };

                setUserInfo({
                    username: decoded.username || '',
                    role: roleMap[decoded.role] || ''
                });
            } catch (error) {
                console.error('Failed to decode token:', error);
            }
        }
    }, []);

    const handleLogout = () => {
        localStorage.removeItem("token");
        navigate('/login');
    };

    const menuItems: MenuProps['items'] = [
        {
            key: 'username',
            label: (
                <div style={{ padding: '4px 0', display: 'flex', alignItems: 'center', gap: '8px' }}>
                    <span>{userInfo.username}</span>
                    <span style={{
                        backgroundColor: '#f0f0f0',
                        padding: '2px 8px',
                        borderRadius: '4px',
                        fontSize: '12px'
                    }}>
                        {userInfo.role}
                    </span>
                </div>
            )
        },
        {
            key: 'divider-1',
            type: 'divider'
        },
        {
            key: 'user-mgm',
            label: "用户管理",
            onClick: () => navigate('/user/user')
        },
        {
            key: 'user-group-mgm',
            label: "用户组管理",
            onClick: () => navigate('/user/user_group')
        },
        {
            key: 'divider-2',
            type: 'divider'
        },
        {
            key: 'logout',
            label: '退出',
            onClick: handleLogout
        }
    ];

    return (
        <div className={styles.toolbar}>
            <Space>
                <li>
                    <Dropdown menu={{ items: menuItems }} placement='bottomRight'>
                        <Avatar size="small" style={{ backgroundColor: '#1677ff', cursor: 'pointer' }}>
                            <UserOutlined style={{ fontSize: 14 }} />
                        </Avatar>
                    </Dropdown>
                </li>
            </Space>
        </div>
    );
};

export default Toolbar;