import { useEffect } from 'react';
// @ts-ignore
import i18nTranslate from 'i18n-jsautotranslate';
import enTranslation from './translations/en';

const I18nTranslateProvider = () => {
    useEffect(() => {
        // i18nTranslate.service.use('none'); // 设置翻译通道
        i18nTranslate.service.use('client.edge'); // 设置翻译通道
        // i18nTranslate.service.use('giteeAI'); // 设置翻译通道
        i18nTranslate.whole.enableAll(); // 启用整体翻译
        i18nTranslate.listener.start();
        // i18nTranslate.office.showPanel();//翻译管理面板
        i18nTranslate.office.fullExtract.isUse = true;
        i18nTranslate.language.setLocal('chinese_simplified'); //设置本地语种（当前网页的语种）

        //读取离线配置
        i18nTranslate.office.append('english', enTranslation);

        i18nTranslate.execute();
        // 解决 input placeholder 延迟渲染问题
        const timer = setTimeout(() => {
            //@ts-ignore
            i18nTranslate.execute();
        }, 500);
        //@ts-ignore
        window.translate = i18nTranslate; // 控制台调试方便
        // 清理定时器 & 监听器（如果需要）
        return () => {
            clearTimeout(timer);
            //@ts-ignore
            i18nTranslate.listener.stop?.(); // 如果有 stop 方法
        };
    }, []);
    return null;
};

export default I18nTranslateProvider;
