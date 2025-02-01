import {Navigate, Route, Routes} from 'react-router-dom'
import Loading from '@/components/Loading'
import {lazy, Suspense} from 'react'
import Layout from '@/layout'
import Login from '@/pages/Login/index.tsx'

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
            <Route path='/' element={<Layout/>}>
                <Route path='/' element={<Navigate to="/dashboard"/>}></Route>
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
