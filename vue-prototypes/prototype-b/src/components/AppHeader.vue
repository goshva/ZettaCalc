<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useCartStore } from '@/stores'
import { storeToRefs } from 'pinia'
const router = useRouter()
const auth = useAuthStore()
const cart = useCartStore()
const { count } = storeToRefs(cart)
function logout() { auth.logout(); router.push('/') }
</script>
<template>
  <header class="hdr">
    <div class="wrap inner">
      <RouterLink to="/" class="brand"><b>ZETTA</b><em>home</em></RouterLink>
      <nav class="nav">
        <RouterLink to="/">Главная</RouterLink>
        <RouterLink to="/gallery">Галерея</RouterLink>
        <RouterLink to="/calculator">Калькулятор</RouterLink>
        <RouterLink to="/flow">Путь клиента</RouterLink>
        <RouterLink to="/cpr">ЦПР</RouterLink>
      </nav>
      <div class="tools">
        <RouterLink to="/cart" class="cart-link">
          Корзина<span v-if="count > 0" class="badge">{{ count }}</span>
        </RouterLink>
        <template v-if="auth.token">
          <span class="user">{{ auth.user?.name ?? auth.user?.email }}</span>
          <button class="btn ghost sm" @click="logout">Выйти</button>
        </template>
        <RouterLink v-else to="/login" class="btn gold sm">Войти</RouterLink>
      </div>
    </div>
  </header>
</template>
<style scoped>
.hdr { position: sticky; top: 0; z-index: 20; background: #f8f6f1f2; backdrop-filter: blur(12px); border-bottom: 1px solid var(--line); }
.inner { display: flex; align-items: center; justify-content: space-between; height: 72px; gap: 28px; }
.brand { font-size: 22px; font-weight: 800; letter-spacing: 0.08em; }
.brand em { font-style: italic; font-weight: 400; color: var(--gold); margin-left: 4px; }
.nav { display: flex; gap: 28px; font-size: 14px; font-weight: 500; }
.nav a:hover, .nav a.router-link-active { color: var(--gold); }
.tools { display: flex; align-items: center; gap: 14px; }
.cart-link { position: relative; font-size: 14px; }
.badge { position: absolute; top: -10px; right: -14px; min-width: 20px; height: 20px; border-radius: 10px; background: var(--gold); color: var(--ink); font-size: 11px; font-weight: 700; display: inline-flex; align-items: center; justify-content: center; padding: 0 5px; }
.user { font-size: 14px; color: #5f584f; }
.btn.sm { min-height: 40px; padding: 8px 16px; }
</style>
