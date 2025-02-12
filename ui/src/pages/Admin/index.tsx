import Amis from "@/components/Amis"
import { useEffect } from "react"
import { useLocation } from 'react-router-dom'
import useStore from "@/store"
import Loading from "@/components/Loading"

const Admin = () => {
    const { schema, loading, initPage } = useStore(state => state)
    const { pathname } = useLocation()
    useEffect(() => {
        initPage(pathname)
    }, [initPage, pathname])
    return loading ? <Loading /> : <Amis schema={schema} />
}

export default Admin
