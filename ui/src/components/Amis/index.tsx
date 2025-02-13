import { registerFilter, registerRenderer, render as renderAmis, Schema } from 'amis'
import { AlertComponent, ToastComponent } from 'amis-ui'
import axios from 'axios'
import { fetcher } from "@/components/Amis/fetcher";
import k8sTextConditionsComponent from "@/components/Amis/custom/K8sTextConditions.tsx";
import NodeRolesComponent from '@/components/Amis/custom/NodeRoles.tsx';
import AutoConvertMemory from "@/components/Amis/custom/AutoConvertMemory.ts";
import FilterAnnotations from "@/components/Amis/custom/FilterAnnotations.ts";
import ShowAnnotationIcon from "@/components/Amis/custom/ShowAnnotationIcon.ts";
import simpleImageName from "@/components/Amis/custom/SimpleImageName.ts";
import FormatBytes from "@/components/Amis/custom/FormatBytes.ts";
import FormatLsShortDate from "@/components/Amis/custom/FormatLsShortDate.ts";
import K8sDate from '@/components/Amis/custom/K8sDate.ts';
import XTermComponent from "@/components/Amis/custom/XTerm.tsx";
import K8sAgeComponent from "@/components/Amis/custom/K8sAge.tsx";
import K8sPodReadyComponent from "@/components/Amis/custom/K8sPodReady.tsx";
import HighlightHtmlComponent from "@/components/Amis/custom/HighlightHtml.tsx";
import WebSocketMarkdownViewerComponent from "@/components/Amis/custom/WebSocketMarkdownViewer.tsx";
import SSELogDownloadComponent from "@/components/Amis/custom/SSELogDownload.tsx";
import SSELogDisplayComponent from "@/components/Amis/custom/SSELogDisplay.tsx";
import WebSocketViewerComponent from "@/components/Amis/custom/WebSocketViewer.tsx";
import WebSocketChatGPT from "@/components/Amis/custom/WebSocketChatGPT.tsx";
import MonacoEditorWithForm from "@/components/Amis/custom/MonacoEditorWithForm.tsx";
import GlobalTextSelector from '@/layout/TextSelectionPopover';
import HistoryRecordsComponent from '@/components/Amis/custom/applyer.tsx';
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
// @ts-ignore
registerRenderer({ type: 'xterm', component: XTermComponent })
// @ts-ignore
registerRenderer({ type: 'chatgpt', component: WebSocketChatGPT })
// @ts-ignore
registerRenderer({ type: 'mEditor', component: MonacoEditorWithForm })
// @ts-ignore
registerRenderer({ type: 'historyRecord', component: HistoryRecordsComponent })
// 注册过滤器
registerFilter("autoConvertMemory", AutoConvertMemory)
registerFilter("filterAnnotations", FilterAnnotations)
registerFilter("showAnnotationIcon", ShowAnnotationIcon)
registerFilter("simpleImageName", simpleImageName)
registerFilter("formatBytes", FormatBytes)
registerFilter("formatLsShortDate", FormatLsShortDate)
registerFilter("k8sDate", K8sDate)

interface Props {
    schema: Schema
}


const Amis = ({ schema }: Props) => {
    const theme = 'cxd';
    const locale = 'zh-CN';


    return <>
        <GlobalTextSelector />

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
                    fetcher,
                    isCancel: value => axios.isCancel(value),
                })
        }
    </>
}
export default Amis
