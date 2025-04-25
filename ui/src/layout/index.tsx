import { Outlet, useLocation, useNavigate } from 'react-router-dom'
import { Layout } from 'antd'
import { GithubOutlined, MenuFoldOutlined, MenuUnfoldOutlined } from '@ant-design/icons'
import Sidebar from '@/components/Sidebar'
import Toolbar from '@/components/ToolBar'
import useStore from '@/store/layout'
import { useCallback, useEffect, useState } from 'react'
import styles from './index.module.scss'
import FloatingChatGPTButton from './FloatingChatGPTButton'
import { fetcher } from '@/components/Amis/fetcher'

const App = () => {
    const { pathname } = useLocation()
    const navigate = useNavigate()
    const { collapse, updateField } = useStore(state => state)
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

    const [produtcName, setProdutcName] = useState("k8m");

    useEffect(() => {
        // 从后端获取配置
        fetcher({
            url: '/params/config/ProductName',
            method: 'get'
        })
            .then(response => {
                //@ts-ignore
                setProdutcName(response.data?.data);
            })
            .catch(error => {
                console.error('Error fetching ProductName config:', error);
                setProdutcName("k8m");
            });
    }, []);
    return <Layout className={styles.container}>
        <Layout.Header style={{
            padding: '0 0',
        }}>
            <div className={styles.navbar}>
                <div className={styles.logo} onClick={goHome}>
                    <h1>
                        <span>{produtcName}</span>
                        <GithubOutlined
                            className='pointer'
                            style={{ marginLeft: '10px', fontSize: '18px' }}
                            onClick={() => window.open('https://github.com/weibaohui/k8m', '_blank')}
                        />
                    </h1>
                </div>
                <Toolbar />
            </div>
        </Layout.Header>
        <Layout>
            <Layout.Sider
                width={220}
                collapsed={collapse}
                collapsible
                onCollapse={onCollapse}
                trigger={<div className={styles.collapse}>
                    {collapse ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
                </div>}
            >
                <Sidebar />
            </Layout.Sider>
            <Layout.Content className={styles.content}>
                <FloatingChatGPTButton></FloatingChatGPTButton>
                <Outlet />

            </Layout.Content>
        </Layout>
    </Layout>
}

export default App
