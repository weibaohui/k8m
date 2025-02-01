import {Outlet, useLocation, useNavigate} from 'react-router-dom'
import {Layout} from '@arco-design/web-react'
import {IconGithub, IconMenuFold, IconMenuUnfold,} from '@arco-design/web-react/icon'
import Sidebar from '@/components/Sidebar'
import Toolbar from '@/components/ToolBar'
import useStore from '@/store/layout'
import {useCallback, useEffect} from 'react'
import menuConfig from './menu'
import styles from './index.module.scss'

const App = () => {
    const {pathname} = useLocation()
    const navigate = useNavigate()
    const {collapse, updateField} = useStore(state => state)
    const onCollapse = useCallback((collapsed: boolean) => {
        updateField('collapse', collapsed)
    }, [updateField])
    const goHome = useCallback(() => {
        navigate('/')
    }, [navigate])
    useEffect(() => {
        const token = localStorage.getItem('token')
        if (!pathname.includes('login') && token === null) {
            navigate('/login')
        }
    }, [navigate, pathname])
    return <Layout className={styles.container}>
        <Layout.Header>
            <div className={styles.navbar}>
                <div className={styles.logo} onClick={goHome}>
                    <h1>
                        <span>k8m</span>
                        <IconGithub
                            className='pointer' style={{marginLeft: '10px'}}
                            fontSize={18}
                            onClick={
                                () => window.open('https://github.com/weibaohui/k8m', '_blank')}/></h1>

                </div>
                <Toolbar/>
            </div>
        </Layout.Header>
        <Layout>
            <Layout.Sider
                width={160}
                defaultCollapsed={collapse}
                collapsible={true}
                onCollapse={onCollapse}
                trigger={<div className={styles.collapse}>
                    {collapse ? <IconMenuUnfold/> : <IconMenuFold/>}
                </div>}
            >
                <Sidebar config={menuConfig}/>
            </Layout.Sider>
            <Layout.Content className={styles.content}>
                <Outlet/>

            </Layout.Content>
        </Layout>
    </Layout>
}

export default App
