// API-клиент прототипа B — отличается базовым URL и стилями
import axios, { type AxiosError } from 'axios'

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL ?? 'http://127.0.0.1:8000',
  timeout: 15000,
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('auth_token_b')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

api.interceptors.response.use(
  (r) => r,
  (err: AxiosError) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('auth_token_b')
      if (window.location.hash !== '#/login') window.location.hash = '#/login'
    }
    return Promise.reject(err)
  },
)

export const formatPrice = (n: number) => new Intl.NumberFormat('ru-RU').format(n) + ' ₽'
