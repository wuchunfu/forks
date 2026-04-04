import axios from 'axios';
import qs from 'qs';

function getApiBaseUrl() {
    // 1. 优先使用显式配置的环境变量
    if (import.meta.env.VITE_API_URL) {
      return import.meta.env.VITE_API_URL
    }
    
    // 开发环境检测
    if (import.meta.env.DEV) {
      return 'http://localhost:8080'
    }
    
    // 生产环境使用当前域名
    const { protocol, hostname, port } = window.location
    let basePort = port ? `:${port}` : ''
    
    // 处理非标准端口
    if ((protocol === 'http:' && port === '80') || 
        (protocol === 'https:' && port === '443')) {
      basePort = ''
    }
    
    return `${protocol}//${hostname}${basePort}`
  }



// 配置新建一个 axios 实例
const service = axios.create({
	baseURL: getApiBaseUrl(),  // 修改为根路径，避免重复的/api
	timeout: 50000,
	headers: { 'Content-Type': 'application/json' },
	paramsSerializer: {
		serialize(params) {
			return qs.stringify(params, { allowDots: true, arrayFormat: 'brackets' });
		},
	},
});

// 添加请求拦截器
service.interceptors.request.use(
	(config) => {
		// 添加token到请求头
		const token = localStorage.getItem('token');
		if (token) {
			config.headers.Authorization = `Bearer ${token}`;
		}
		return config;
	},
	(error) => {
		return Promise.reject(error);
	}
);

// 添加响应拦截器
service.interceptors.response.use(
	(response) => {
		// 如果是下载文件，直接返回response
		if (response.config.responseType === 'blob') {
			return response;
		}

		const res = response.data;
		// 只有当API明确返回错误时才reject
		if (res && res.code && res.code !== 0) {
			console.error('API错误:', res.message);
			return Promise.reject(new Error(res.message || '未知错误'));
		}

		// 成功响应直接返回
		return response;
	},
	(error) => {
		// 处理401未授权错误
		if (error.response && error.response.status === 401) {
			// 清除本地存储的认证信息
			localStorage.removeItem('token');
			localStorage.removeItem('userInfo');

			// 跳转到登录页面
			window.location.href = '/login';

			return Promise.reject(new Error('认证失败，请重新登录'));
		}

		// 对于其他HTTP错误，尝试提取后端返回的错误信息
		if (error.response && error.response.data) {
			const errorData = error.response.data
			const errorMessage = errorData.message || errorData.error || `请求失败 (${error.response.status})`
			console.error('请求错误:', errorMessage)
			return Promise.reject(new Error(errorMessage))
		}

		console.error('请求错误:', error.message || '网络错误')
		return Promise.reject(error);
	}
);

// 导出 axios 实例
export default service;
