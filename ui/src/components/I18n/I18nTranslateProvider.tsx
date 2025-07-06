import { useEffect } from 'react';
// @ts-ignore
import i18nTranslate from 'i18n-jsautotranslate';

const I18nTranslateProvider = () => {
    useEffect(() => {
        i18nTranslate.service.use('client.edge'); // 设置翻译通道
        i18nTranslate.whole.enableAll(); // 启用整体翻译
        i18nTranslate.listener.start();
        i18nTranslate.office.showPanel();//翻译管理面板
        i18nTranslate.office.fullExtract.isUse = true;
        i18nTranslate.language.setLocal('chinese_simplified'); //设置本地语种（当前网页的语种）


        //加入离线翻译-切换为繁体中文的配置
        i18nTranslate.office.append('chinese_traditional', `
    库=庫
    代码=代碼
    引入=引入
    版本，=版本，
    简介:=簡介:
    语言切换示例：=語言切換示例：
    当前为 =當前為
    选择框切换语言:=選擇框切換語言:
    国际化，网页自动翻译，同谷歌浏览器自动翻译的效果，适用于网站。=國際化，網頁自動翻譯，同谷歌瀏覽器自動翻譯的效果，適用於網站。
    进行翻译=進行翻譯
    按钮切换语言:=按鈕切換語言:
    如果你网页中有=如果你網頁中有
    版本参见：=版本參見：
    网页自动翻译，页面无需另行改造，加入两行 =網頁自動翻譯，頁面無需另行改造，加入兩行
    即可让你的网页快速具备多国语言切换能力！=即可讓你的網頁快速具備多國語言切換能力！
    使用方式:=使用方式:
    在页面最底部加入=在頁面最底部加入
    注意，要在页面最底部加。如果你在页面顶部加，那下面的是不会被翻译的=注意，要在頁面最底部加。如果你在頁面頂部加，那下面的是不會被翻譯的
    请求更新了数据，要对其更新的数据进行翻译时，可直接执行 =請求更新了數據，要對其更新的數據進行翻譯時，可直接執行
    js=js
    js =js
    v1 =v1
    v2 =v2
    ajax=阿賈克斯
    demo=演示
    hello, =你好，
    v1.html=v1.html
    select=選擇
    `);

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
