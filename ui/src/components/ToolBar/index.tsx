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
                //role 可能为guest,platform_admin，也可能为guset
                setUserInfo({
                    username: decoded.username || '',
                    role: decoded.role.includes('platform_admin') ? '平台管理员' : '普通用户'
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
            key: "user_profile_login_settings",
            label: "登录设置",
            icon: <i className="fa-solid fa-key"></i>,
            onClick: () => navigate('/user/profile/login_settings')
        },
        {
            key: "user_profile_clusters",
            label: "我的集群",
            icon: <i className="fa-solid fa-server"></i>,
            onClick: () => navigate('/user/profile/my_clusters')
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
                        <span style={{ cursor: 'pointer' }}>
                            <Avatar size="small" style={{ backgroundColor: '#1677ff' }} >
                                <UserOutlined style={{ fontSize: 14 }} />
                            </Avatar>
                            {/* <span className='ml-1'>{userInfo.username}</span> */}
                        </span>
                    </Dropdown>
                </li>
            </Space>
        </div >
    );
};

export default Toolbar;