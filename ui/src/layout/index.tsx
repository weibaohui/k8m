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
//@ts-ignore
import i18nTranslate from 'i18n-jsautotranslate';
//@ts-ignore
window.translate = i18nTranslate; // 控制台调试方便

i18nTranslate.service.use('client.edge'); // 设置翻译通道
i18nTranslate.whole.enableAll(); // 启用整体翻译

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
        // 初始翻译执行
        //@ts-ignore
        translate.execute();

        // 解决 input placeholder 延迟渲染问题
        const timer = setTimeout(() => {
            //@ts-ignore
            translate.execute();
        }, 500);

        // 开启监听 DOM 更新（例如 MutationObserver）
        //@ts-ignore
        translate.listener.start();
        //@ts-ignore
        translate.office.showPanel();
        //@ts-ignore
        translate.office.fullExtract.isUse = true;
        // 清理定时器 & 监听器（如果需要）
        return () => {
            clearTimeout(timer);
            //@ts-ignore
            translate.listener.stop?.(); // 如果有 stop 方法
        };
    }, []);

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
    function LangLink({ lang, label }: { lang: string, label: string }) {
        return (
            <a
                href="#"
                onClick={(e) => {
                    e.preventDefault();
                    //@ts-ignore
                    translate.changeLanguage(lang);
                }}
                className="ignore"
            >
                {label}
            </a>
        );
    }
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
