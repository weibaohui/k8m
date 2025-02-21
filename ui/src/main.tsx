import ReactDOM from 'react-dom/client'
import 'amis/lib/themes/cxd.css'
import 'amis/lib/helper.css'
import 'antd/dist/reset.css'
import '@/styles/global.scss'
import '@fortawesome/fontawesome-free/css/all.css';
import App from './App.tsx'
import { ConfigProvider } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import dayjs from 'dayjs';
import 'dayjs/locale/zh-cn';

dayjs.locale('zh-cn');

ReactDOM.createRoot(document.getElementById('root')!).render(
    <ConfigProvider locale={zhCN}>
        <App />
    </ConfigProvider>
)
