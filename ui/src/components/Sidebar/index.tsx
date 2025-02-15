import useStore from '@/store/layout'
import { Menu } from 'antd';
import menu from '@/components/Sidebar/menu';
const Sidebar = () => {
    const { collapse } = useStore(state => state)

    return <Menu style={{ minWidth: 0, flex: "auto" }}
        mode="inline"
        inlineCollapsed={collapse}
        defaultOpenKeys={['home']}
        defaultSelectedKeys={[]}
        items={menu()}
    >

    </Menu>
}

export default Sidebar
