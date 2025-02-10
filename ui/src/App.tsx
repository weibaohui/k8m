import { HashRouter } from 'react-router-dom'
import routes from '@/routes'
import GlobalTextSelector from './layout/TextSelectionPopover'

const App = () => {
    return <HashRouter>{routes()}
        <GlobalTextSelector />
    </HashRouter>
    // return <BrowserRouter basename={'/'}>{routes()}</BrowserRouter>
}
export default App
