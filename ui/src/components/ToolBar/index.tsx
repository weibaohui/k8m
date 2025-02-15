import { Avatar, Dropdown, Menu, Space } from 'antd'
import { useNavigate } from 'react-router-dom'
import styles from './index.module.scss'
import { UserOutlined } from '@ant-design/icons'

const Toolbar = () => {
    const navigate = useNavigate()
    const handleLogout = () => {
        localStorage.removeItem("token")
        navigate('/login')
    }

    const menuItems = [
        {
            key: 'logout',
            label: (
                <div onClick={() => handleLogout()} >
                    退出
                </div>
            )
        }
    ];
    return <div className={styles.toolbar}>
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
}

export default Toolbar
