import {Menu} from '@arco-design/web-react'
import {useNavigate} from 'react-router-dom'
import {useCallback} from 'react';
import useStore from '@/store/layout'

const MenuItem = Menu.Item;
const SubMenu = Menu.SubMenu;

type MenuItemType = {
    label: string
    path?: string
    icon?: string
    key: string
    children?: MenuItemType[]
}

interface Props {
    config: MenuItemType[]
}

const Sidebar = ({config}: Props) => {
    const navigate = useNavigate()
    const {collapse} = useStore(state => state)
    const renderIcon = useCallback((icon?: string) => {
        return (
            <>
                <i className={icon}></i>
            </>
        )

    }, [])
    const onMenuClick = (item: MenuItemType) => {
        if (item.path) {
            navigate(item.path)
        }
    }
    return <Menu
        collapse={collapse}
        defaultOpenKeys={['home']}
        defaultSelectedKeys={[]}
    >
        {
            config.map((item: MenuItemType) => {
                if (item.children && item.children.length) {
                    return <SubMenu
                        key={item.key}
                        title={
                            <>
                                {renderIcon(item.icon)} {item.label}
                            </>
                        }
                    >
                        {
                            item.children.map(sub => {
                                return <MenuItem  key={sub.key}
                                                 onClick={() => onMenuClick(sub)}>{sub.label}</MenuItem>
                            })
                        }
                    </SubMenu>
                } else {
                    return <MenuItem key={item.key}   onClick={() => onMenuClick(item)}>
                        {renderIcon(item.icon)} {item.label}
                    </MenuItem>
                }
            })
        }
    </Menu>
}

export default Sidebar
