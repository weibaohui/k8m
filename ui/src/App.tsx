import { HashRouter } from 'react-router-dom'
import routes from '@/routes'

const App = () => {
    return <HashRouter>{routes()}
    </HashRouter>
    // return <BrowserRouter basename={'/'}>{routes()}</BrowserRouter>
}
export default App
