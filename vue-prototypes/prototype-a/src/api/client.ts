import axios, { type AxiosError } from 'axios'

export const API_BASE = import.meta.env.VITE_API_URL ?? ''

export const api = axios.create({
  baseURL: API_BASE,
  timeout: 15000,
})

// JWT-интерцептор: добавляем Bearer токен
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('auth_token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// Глобальная обработка 401 — разлогин
api.interceptors.response.use(
  (r) => r,
  (err: AxiosError) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('auth_token')
      if (window.location.hash !== '#/login') window.location.hash = '#/login'
    }
    return Promise.reject(err)
  },
)

export const formatPrice = (n: number) => new Intl.NumberFormat('ru-RU').format(n) + ' ₽'
