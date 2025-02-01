import axios from 'axios'

const ajax = axios.create({
    baseURL: '/',
    timeout: 30000
})
export default ajax