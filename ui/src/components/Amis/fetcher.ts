import {FetcherConfig} from "amis-core/lib/factory";
import {fetcherResult} from "amis-core/lib/types";
import {message} from "antd";
import axios from "axios";
import {ProcessK8sUrlWithCluster} from "@/utils/utils.ts";


export const fetcher = ({url, method = 'get', data, config}: FetcherConfig): Promise<fetcherResult> => {
    const token = localStorage.getItem('token') || '';

    const ajax = axios.create({
        baseURL: '/',
        headers: {
            ...config?.headers,
            Authorization: token ? `Bearer ${token}` : ''
        }
    });

    // 请求发送之前的拦截
    ajax.interceptors.request.use(
        config => {
            if (config.url) {
                config.url = ProcessK8sUrlWithCluster(config.url);
            }
            return config;
        },
        error => {
            return Promise.reject(error);
        }
    );

    // 请求发送之前的拦截
    ajax.interceptors.response.use(
        response => response, // 请求成功
        error => {
            if (error.response && error.response.status === 401) {
                // 如果是401，跳转到登录页面
                window.location.href = '/#/login';
            }
            if (error.response && error.response.status === 512) {
                var cluster = error.response.data.msg;
                message.info(`${cluster}。如有疑问请联系管理员。`)
                window.location.href = '/#/user/cluster/cluster_user';
            }
            if (error.response && error.response.status === 403) {
                message.error(`权限不足，请联系管理员。`)
            }
            return Promise.reject(error); // 继续处理其他错误
        }
    );

    switch (method.toLowerCase()) {
        case 'get':
            return ajax.get(url, config);
        case 'post':
            return ajax.post(url, data || null, config);
        default:
            return ajax.post(url, data || null, config);
    }
};
