import {Avatar, Dropdown, Space, MenuProps} from 'antd';
import {useNavigate} from 'react-router-dom';
import styles from './index.module.scss';
import {UserOutlined} from '@ant-design/icons';
import {useEffect, useState} from 'react';
import {jwtDecode} from 'jwt-decode';

interface DecodedToken {
    username: string;
    role: string;
    // JWT标准字段
    exp: number;
    iat: number;
}

const Toolbar = () => {
    const navigate = useNavigate();
    const [userInfo, setUserInfo] = useState<{ username: string; role: string }>({username: '', role: ''});

    useEffect(() => {
        const token = localStorage.getItem('token');
        if (token) {
            try {
                const decoded = jwtDecode<DecodedToken>(token);
                setUserInfo({
                    username: decoded.username || '',
                    role: decoded.role || ''
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
                <div style={{padding: '4px 0'}}>
                    <div>用户名：{userInfo.username}</div>
                    <div>角色：{userInfo.role}</div>
                </div>
            )
        },
        {
            key: 'divider',
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
                    <Dropdown menu={{items: menuItems}} placement='bottomRight'>
                        <Avatar size="small" style={{backgroundColor: '#1677ff', cursor: 'pointer'}}>
                            <UserOutlined style={{fontSize: 14}}/>
                        </Avatar>
                    </Dropdown>
                </li>
            </Space>
        </div>
    );
};

export default Toolbar;