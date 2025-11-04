import Amis from "@/components/Amis"
import { useEffect } from "react"
import { useLocation } from 'react-router-dom'
import useStore from "@/store"
import Loading from "@/components/Loading"

/**
 * Admin 页面组件
 * 负责根据当前路由加载对应的页面 Schema。
 * 注意：此处需要去掉路由中的集群前缀（如：#/k/<clusterID>/...），
 * 只保留后续的业务页面路径传入 initPage。
 */
const Admin = () => {
    const { schema, loading, initPage } = useStore(state => state)
    const { pathname, hash } = useLocation()

    /**
     * 去除集群前缀，保留业务页面路径
     * 规则：
     * - 优先使用 Hash 路由（形如 #/k/<clusterID>/xxx），否则回退到 pathname
     * - 若路径以 /k/<clusterID>/ 开头，则剥离前两段（k 与 clusterID）
     */
    const stripClusterPrefix = (rawPath: string): string => {
        const hasQuery = rawPath.includes('?')
        const pathPart = hasQuery ? rawPath.substring(0, rawPath.indexOf('?')) : rawPath
        const queryPart = hasQuery ? rawPath.substring(rawPath.indexOf('?')) : ''

        const segments = pathPart.split('/').filter(Boolean)
        let cleanedSegments = segments
        if (segments.length >= 2 && segments[0] === 'k') {
            cleanedSegments = segments.slice(2)
        }

        const cleaned = '/' + cleanedSegments.join('/')
        return cleaned + queryPart
    }

    useEffect(() => {
        const raw = hash && hash.startsWith('#/') ? hash.slice(1) : pathname
        const cleaned = stripClusterPrefix(raw)
        initPage(cleaned)
    }, [initPage, pathname, hash])

    return loading ? <Loading /> : <Amis schema={schema} />
}

export default Admin