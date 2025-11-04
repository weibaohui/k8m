import { Navigate, Route, Routes } from 'react-router-dom'
import Loading from '@/components/Loading'
import { lazy, Suspense } from 'react'
import Layout from '@/layout'
import Login from '@/pages/Login/index.tsx'
import PodExec from '@/pages/PodExec'
import PodLog from '@/pages/PodLog'
import NodeExec from '@/pages/NodeExec'
import MenuEditor from '@/pages/MenuEditor'

const lazyLoad = (Component: React.LazyExoticComponent<() => JSX.Element>) => {
    return (
        <Suspense fallback={<Loading />}>
            <Component />
        </Suspense>
    )
}
const Router = () => {
    return (
        <Routes>
            <Route path='/login' element={<Login />}></Route>
            <Route path='/k/:cluster/NodeExec' element={<NodeExec />}></Route>
            <Route path='/k/:cluster/PodExec' element={<PodExec />}></Route>
            <Route path='/k/:cluster/PodLog' element={<PodLog />}></Route>
            <Route path='/k/:cluster/MenuEditor' element={<MenuEditor />}></Route>
            <Route path='/' element={<Layout />}>
                <Route path='/' element={<Navigate to="/user/cluster/cluster_user" />}></Route>
                <Route path='/*' element={
                    lazyLoad(
                        lazy(
                            async () => import('@/pages/Admin/index.tsx')
                        )
                    )
                }></Route>
            </Route>
        </Routes>
    )
}

export default Router
