import {Avatar, Dropdown, Menu, Space} from '@arco-design/web-react'
import {useNavigate} from 'react-router-dom'
import {IconUser} from '@arco-design/web-react/icon'
import styles from './index.module.scss'

const Toolbar = () => {
    const navigate = useNavigate()
    const handleClick = (key: string) => {
        switch (key) {
            case 'logout':
                localStorage.clear()
                navigate('/login')
                break
        }
    }
    const dropList = (
        <Menu onClickMenuItem={handleClick}>
            <Menu.Item key='logout'>退出</Menu.Item>
        </Menu>
    );
    return <div className={styles.toolbar}>
        <Space size="medium">

            <li>
                <Dropdown droplist={dropList} trigger='click' position='br'>
                    <Avatar size={18} style={{backgroundColor: '#3370ff', cursor: 'pointer'}}>
                        <IconUser style={{fontSize: 18}}/>
                    </Avatar>
                </Dropdown>
            </li>
            <li>

            </li>
        </Space>

    </div>
}

export default Toolbar
