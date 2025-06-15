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
import WebSocketViewerComponent from "@/components/Amis/custom/WebSocketViewer.tsx";
import WebSocketChatGPT from "@/components/Amis/custom/WebSocketChatGPT.tsx";
import MonacoEditorWithForm from "@/components/Amis/custom/MonacoEditorWithForm.tsx";
import GlobalTextSelector from '@/layout/TextSelectionPopover';
import HistoryRecordsComponent from '@/components/Amis/custom/YamlApplyer/YamlApplyer.tsx';
import DiffEditorComponent from '@/components/Amis/custom/DiffEditor/index.tsx';
import DeploymentRevisionDiffEditor from '@/components/Amis/custom/DiffEditor/DeploymentRevisonDiffEditor.tsx';
import PodLogViewerComponent from '@/components/Amis/custom/LogView/PodLogViewer';
import PodsLogViewerComponent from '@/components/Amis/custom/LogView/PodsLogViewer';
import KubeConfigEditorComponent from '@/components/Amis/custom/KubeConfigEditor.tsx'
import PasswordEditorWithForm from "@/components/Amis/custom/PasswordEditorWithForm/PasswordEditorWithForm.tsx";
import K8sPodStatusComponent from "@/components/Amis/custom/K8sPodStatus.tsx";
import HPAMetricsComponent from '@/components/Amis/custom/HPAMetrics';
import HPABehaviorComponent from '@/components/Amis/custom/HPABehavior';
import HelmUpdateRelease from '@/components/Amis/custom/Helm/HelmUpdateRealease.tsx';
import K8sGPTComponent from '@/components/Amis/custom/K8sGPT';
import InspectionSummaryComponent from '@/components/Amis/custom/InspectionSummary.tsx'
import InspectionEventListComponent from '@/components/Amis/custom/InspectionEventList.tsx'
// 注册自定义组件
registerRenderer({ type: 'k8sTextConditions', component: k8sTextConditionsComponent })
registerRenderer({ type: 'nodeRoles', component: NodeRolesComponent })
// @ts-ignore
registerRenderer({ type: 'k8sAge', component: K8sAgeComponent })
// @ts-ignore
registerRenderer({ type: 'k8sPodReady', component: K8sPodReadyComponent })
// @ts-ignore
registerRenderer({ type: 'highlightHtml', component: HighlightHtmlComponent })
// @ts-ignore
registerRenderer({ type: 'webSocketMarkdownViewer', component: WebSocketMarkdownViewerComponent })
// @ts-ignore
registerRenderer({ type: 'podLogViewer', component: PodLogViewerComponent })
// @ts-ignore
registerRenderer({ type: 'podsLogViewer', component: PodsLogViewerComponent })
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

// @ts-ignore
registerRenderer({ type: 'diffEditor', component: DiffEditorComponent })
// @ts-ignore
registerRenderer({ type: 'deploymentRevisionDiffEditor', component: DeploymentRevisionDiffEditor })
//@ts-ignore
registerRenderer({ type: 'kubeConfigEditor', component: KubeConfigEditorComponent })
//@ts-ignore
registerRenderer({ type: 'passwordEditor', component: PasswordEditorWithForm })
//@ts-ignore
registerRenderer({ type: 'k8sPodStatus', component: K8sPodStatusComponent })
//@ts-ignore
registerRenderer({ type: 'hpaMetrics', component: HPAMetricsComponent })
//@ts-ignore
registerRenderer({ type: 'hpaBehavior', component: HPABehaviorComponent })
//@ts-ignore
registerRenderer({ type: 'helmUpdateRelease', component: HelmUpdateRelease })
//@ts-ignore
registerRenderer({ type: 'k8sGPT', component: K8sGPTComponent })
//@ts-ignore
registerRenderer({ type: 'inspectionSummary', component: InspectionSummaryComponent })
//@ts-ignore
registerRenderer({ type: 'inspectionEventList', component: InspectionEventListComponent })


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
                    updateLocation: () => {
                    },
                    fetcher,
                    isCancel: value => axios.isCancel(value),
                })
        }
    </>
}
export default Amis
