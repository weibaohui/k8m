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
                const overrideCluster = (config.headers && (config.headers as any)['x-k8m-target-cluster']) || (config.params && (config.params as any).__cluster);
                config.url = ProcessK8sUrlWithCluster(config.url, overrideCluster as string | undefined);
                if (config.params && (config.params as any).__cluster) {
                    const { __cluster, ...rest } = config.params as any;
                    config.params = rest;
                }
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
            // 将 data 作为查询参数透传，合并已有的 config.params
            const finalConfig = {
                ...config,
                params: {
                    ...(config?.params || {}),
                    ...(data && typeof data === 'object' ? data : {}),
                }
            };
            return ajax.get(url, finalConfig).then(response => {
                // 检查响应是否为空或无效
                if (!response) {
                    console.error('Empty response from server for URL:', url);
                    throw new Error('服务器返回空响应');
                }
                if (response.data === undefined || response.data === null || response.data === '') {
                    console.error('Response data is null/undefined/empty for URL:', url, 'data:', response.data);
                    throw new Error('响应数据为空');
                }
                console.log('Fetcher response for', url, ':', response);
                return response;
            });
        case 'post':
            return ajax.post(url, data || null, config).then(response => {
                // 检查响应是否为空或无效
                if (!response) {
                    console.error('Empty response from server for URL:', url);
                    throw new Error('服务器返回空响应');
                }
                if (response.data === undefined || response.data === null || response.data === '') {
                    console.error('Response data is null/undefined/empty for URL:', url, 'data:', response.data);
                    throw new Error('响应数据为空');
                }
                console.log('Fetcher response for', url, ':', response);
                return response;
            });
        default:
            return ajax.post(url, data || null, config).then(response => {
                // 检查响应是否为空或无效
                if (!response) {
                    console.error('Empty response from server for URL:', url);
                    throw new Error('服务器返回空响应');
                }
                if (response.data === undefined || response.data === null || response.data === '') {
                    console.error('Response data is null/undefined/empty for URL:', url, 'data:', response.data);
                    throw new Error('响应数据为空');
                }
                console.log('Fetcher response for', url, ':', response);
                return response;
            });
    }
};
