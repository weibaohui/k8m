import { Avatar, Dropdown, Menu, Space } from 'antd'
import { useNavigate } from 'react-router-dom'
import styles from './index.module.scss'
import { UserOutlined } from '@ant-design/icons'

const Toolbar = () => {
    const navigate = useNavigate()
    const handleClick = (e: any) => {
        switch (e.key) {
            case 'logout':
                localStorage.removeItem("token")
                navigate('/login')
                break
        }
    }
    const dropList = (
        <Menu onClick={handleClick}>
            <Menu.Item key='logout'>退出</Menu.Item>
        </Menu>
    );
    return <div className={styles.toolbar}>
        <Space>
            <li>
                <Dropdown overlay={dropList} placement='bottomRight'>
                    <Avatar size="small" style={{ backgroundColor: '#1677ff', cursor: 'pointer' }}>
                        <UserOutlined style={{ fontSize: 14 }} />
                    </Avatar>
                </Dropdown>
            </li>
        </Space>
    </div>
}

export default Toolbar
