import {Navigate, Route, Routes} from 'react-router-dom'
import Loading from '@/components/Loading'
import {lazy, Suspense} from 'react'
import Layout from '@/layout'
import Login from '@/pages/Login/index.tsx'
import PodExec from '@/pages/PodExec'
import PodLog from '@/pages/PodLog'
import NodeExec from '@/pages/NodeExec'

const lazyLoad = (Component: React.LazyExoticComponent<() => JSX.Element>) => {
    return (
        <Suspense fallback={<Loading/>}>
            <Component/>
        </Suspense>
    )
}
const Router = () => {
    return (
        <Routes>
            <Route path='/login' element={<Login/>}></Route>
            <Route path='/NodeExec' element={<NodeExec/>}></Route>
            <Route path='/PodExec' element={<PodExec/>}></Route>
            <Route path='/PodLog' element={<PodLog/>}></Route>
            <Route path='/' element={<Layout/>}>
                <Route path='/' element={<Navigate to="/user/cluster/cluster_user"/>}></Route>
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
