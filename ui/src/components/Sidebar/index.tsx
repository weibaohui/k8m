import { Menu, Tooltip } from 'antd'
import { useNavigate } from 'react-router-dom'
import { useCallback } from 'react';
import useStore from '@/store/layout'

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

const Sidebar = ({ config }: Props) => {
    const navigate = useNavigate()
    const { collapse } = useStore(state => state)
    const renderIcon = useCallback((icon?: string) => {
        icon = icon + " mr-0.5 "
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
        mode="inline"
        inlineCollapsed={collapse}
        defaultOpenKeys={['home']}
        defaultSelectedKeys={[]}
    >
        {
            config.map((item: MenuItemType) => {
                if (item.children && item.children.length) {
                    return <Menu.SubMenu
                        key={item.key}
                        title={
                            <>
                                {renderIcon(item.icon)} {item.label}
                            </>
                        }
                    >
                        {
                            item.children.map(sub => {
                                return (
                                    <Menu.Item key={sub.key}
                                        onClick={() => onMenuClick(sub)}>
                                        <Tooltip placement='right' title={sub.label}>
                                            {renderIcon(sub.icon)}{sub.label}
                                        </Tooltip>
                                    </Menu.Item>)
                            })
                        }
                    </Menu.SubMenu>
                } else {
                    return <Menu.Item key={item.key} onClick={() => onMenuClick(item)}>
                        {renderIcon(item.icon)} {item.label}
                    </Menu.Item>
                }
            })
        }
    </Menu>
}

export default Sidebar
