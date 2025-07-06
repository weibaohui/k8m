import axios from 'axios'

const ajax = axios.create({
    baseURL: '/k8m/api',
    timeout: 30000
})

ajax.interceptors.request.use(function (config) {
    // 在发送请求之前做些什么
    // 如果url是.json结尾
    if (config.url?.endsWith('.json')) {
        // 修改baseurl
        config.baseURL = '/k8m/ui'
    }
    return config;
}, function (error) {
    // 对请求错误做些什么
    return Promise.reject(error);
});

export default ajax