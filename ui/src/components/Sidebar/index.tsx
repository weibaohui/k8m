import useStore from '@/store/layout'
import { Menu } from 'antd';
import menu from '@/components/Sidebar/menu';
const Sidebar = () => {
    const { collapse } = useStore(state => state)

    return (
        <div style={{ height: 'calc(100vh - 64px)', minWidth: 0, flex: "auto", overflow: 'auto' }}>
            <Menu
                mode="inline"
                inlineCollapsed={collapse}
                defaultOpenKeys={['home']}
                defaultSelectedKeys={[]}
                items={menu()}
                style={{ height: '100%', borderRight: 0 }}
            >
            </Menu>
        </div>
    )
}

export default Sidebar
