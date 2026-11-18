import axios from 'axios'
import { load } from '@/utils/storage'

export const http = axios.create({ baseURL: '/api' })

http.interceptors.request.use((cfg) => {
  const token = load('token', '')
  if (token) cfg.headers.Authorization = `Bearer ${token}`
  return cfg
})

http.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401 && !err.config.url.includes('/auth/login')) {
      window.dispatchEvent(new Event('auth:expired'))
    }
    return Promise.reject(err)
  },
)

const data = (res) => res.data.data

export const get = (url, config) => http.get(url, config).then(data)
export const post = (url, body) => http.post(url, body).then(data)
export const put = (url, body) => http.put(url, body).then(data)
export const del = (url) => http.delete(url).then(data)

export const errMsg = (e) => e.response?.data?.message || 'Tidak bisa terhubung ke server'
