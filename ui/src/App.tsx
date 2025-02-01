import {BrowserRouter, HashRouter} from 'react-router-dom'
import routes from '@/routes'
import {isDev} from './utils/utils'

const App = () => {
    if (!isDev) {
        return <HashRouter>{routes()}</HashRouter>
    }
    return <BrowserRouter basename={'/'}>{routes()}</BrowserRouter>
}
export default App