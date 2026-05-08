import axios, { type AxiosInstance, type AxiosRequestConfig, type AxiosResponse } from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

const service: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL ?? '/api',
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' },
})

service.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('kx_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error),
)

service.interceptors.response.use(
  (response: AxiosResponse) => {
    const res = response.data
    if (res.code !== 0) {
      ElMessage.error(res.message || '请求失败')
      if (res.code === 401) {
        localStorage.removeItem('kx_token')
        router.push('/login')
      }
      return Promise.reject(new Error(res.message || '请求失败'))
    }
    return res
  },
  (error) => {
    const msg = error.response?.data?.message || error.message || '网络异常'
    ElMessage.error(msg)
    if (error.response?.status === 401) {
      localStorage.removeItem('kx_token')
      router.push('/login')
    }
    return Promise.reject(error)
  },
)

export function get<T = any>(url: string, params?: any, config?: AxiosRequestConfig) {
  return service.get<any, T>(url, { params, ...config })
}

export function post<T = any>(url: string, data?: any, config?: AxiosRequestConfig) {
  return service.post<any, T>(url, data, config)
}

export function put<T = any>(url: string, data?: any, config?: AxiosRequestConfig) {
  return service.put<any, T>(url, data, config)
}

export function del<T = any>(url: string, params?: any) {
  return service.delete<any, T>(url, { params })
}

export function upload<T = any>(url: string, file: File, onProgress?: (p: number) => void) {
  const formData = new FormData()
  formData.append('file', file)
  return service.post<any, T>(url, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    onUploadProgress: (e) => {
      if (onProgress && e.total) {
        onProgress(Math.round((e.loaded * 100) / e.total))
      }
    },
  })
}

export default service
