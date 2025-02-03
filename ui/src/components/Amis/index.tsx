import { registerFilter, registerRenderer, render as renderAmis, Schema } from 'amis'
import { AlertComponent, ToastComponent } from 'amis-ui'
import axios from 'axios'
import k8sTextConditionsComponent from "@/components/Amis/custom/K8sTextConditions.tsx";
import NodeRolesComponent from '@/components/Amis/custom/NodeRoles.tsx';
import AutoConvertMemory from "@/components/Amis/custom/AutoConvertMemory.ts";
import FilterAnnotations from "@/components/Amis/custom/FilterAnnotations.ts";
import ShowAnnotationIcon from "@/components/Amis/custom/ShowAnnotationIcon.ts";
import simpleImageName from "@/components/Amis/custom/SimpleImageName.ts";
import FormatBytes from "@/components/Amis/custom/FormatBytes.ts";
import FormatLsShortDate from "@/components/Amis/custom/FormatLsShortDate.ts";
import K8sAgeComponent from "@/components/Amis/custom/K8sAge.tsx";
import K8sPodReadyComponent from "@/components/Amis/custom/K8sPodReady.tsx";
import HighlightHtmlComponent from "@/components/Amis/custom/HighlightHtml.tsx";
import WebSocketMarkdownViewerComponent from "@/components/Amis/custom/WebSocketMarkdownViewer.tsx";
import SSELogDownloadComponent from "@/components/Amis/custom/SSELogDownload.tsx";
import SSELogDisplayComponent from "@/components/Amis/custom/SSELogDisplay.tsx";
import WebSocketViewerComponent from "@/components/Amis/custom/WebSocketViewer.tsx";
// 注册自定义组件
registerRenderer({ type: 'k8sTextConditions', component: k8sTextConditionsComponent })
registerRenderer({ type: 'nodeRoles', component: NodeRolesComponent })
// @ts-ignore
registerRenderer({ type: 'k8sAge', component: K8sAgeComponent })
registerRenderer({ type: 'k8sPodReady', component: K8sPodReadyComponent })
// @ts-ignore
registerRenderer({ type: 'highlightHtml', component: HighlightHtmlComponent })
// @ts-ignore
registerRenderer({ type: 'webSocketMarkdownViewer', component: WebSocketMarkdownViewerComponent })
// @ts-ignore
registerRenderer({ type: 'log-download', component: SSELogDownloadComponent })
// @ts-ignore
registerRenderer({ type: 'log-display', component: SSELogDisplayComponent })
// @ts-ignore
registerRenderer({ type: 'websocketViewer', component: WebSocketViewerComponent })

// 注册过滤器
registerFilter("autoConvertMemory", AutoConvertMemory)
registerFilter("filterAnnotations", FilterAnnotations)
registerFilter("showAnnotationIcon", ShowAnnotationIcon)
registerFilter("simpleImageName", simpleImageName)
registerFilter("formatBytes", FormatBytes)
registerFilter("formatLsShortDate", FormatLsShortDate)

interface Props {
    schema: Schema
}

const Amis = ({ schema }: Props) => {
    const theme = 'cxd';
    const locale = 'zh-CN';

    return <>
        <ToastComponent
            theme={theme}

            position={'top-center'}
            locale={locale}
        />
        <AlertComponent theme={theme} key="alert" locale={locale} />
        {

            renderAmis(schema,
                {},
                {
                    theme: 'cxd',
                    updateLocation: (to: unknown, replace: unknown) => {
                        console.log(to)
                        console.log(replace)
                    },
                    fetcher: ({
                        url, // 接口地址
                        method, // 请求方法 get、post、put、delete
                        data, // 请求数据
                        config, // 其他配置
                    }) => {
                        const token = localStorage.getItem('token') || '';

                        const ajax = axios.create({
                            baseURL: '/',
                            headers: {
                                ...config?.headers,
                                Authorization: token ? `Bearer ${token}` : ''
                            }
                        });
                        switch (method) {
                            case 'get':
                                return ajax.get(url, config)
                            case 'post':
                                return ajax.post(url, data || null, config)
                            default:
                                return ajax.post(url, data || null, config)
                        }
                    },
                    isCancel: value => axios.isCancel(value),
                })
        }
    </>
}
export default Amis
