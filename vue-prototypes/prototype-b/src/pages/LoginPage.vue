<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const auth = useAuthStore()
const email = ref('test@zetta.ru')
const password = ref('test123')

async function submit() {
  try { await auth.login(email.value, password.value)
    const redirect = router.currentRoute.value.query.redirect as string
    router.push(redirect || '/')
  } catch {}
}
</script>
<template>
  <section class="login">
    <div class="login-card">
      <h2>Вход</h2>
      <p class="muted">test@zetta.ru / test123</p>
      <form @submit.prevent="submit">
        <div class="field">
          <label>Email</label>
          <input v-model="email" class="input" type="email" required />
        </div>
        <div class="field">
          <label>Пароль</label>
          <input v-model="password" class="input" type="password" required />
        </div>
        <p v-if="auth.error" class="error">{{ auth.error }}</p>
        <button class="btn gold" type="submit" :disabled="auth.loading">
          {{ auth.loading ? 'Вход…' : 'Войти' }}
        </button>
      </form>
    </div>
  </section>
</template>
<style scoped>
.login { display: flex; align-items: center; justify-content: center; min-height: 80vh; }
.login-card { background: #fff; border-radius: var(--radius); padding: 40px; width: 100%; max-width: 420px; }
.field { margin-bottom: 18px; }
.field label { display: block; font-size: 12px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.1em; color: #5f584f; margin-bottom: 8px; }
.error { color: var(--gold, #cf3d2b); background: #fef3f0; padding: 10px; border-radius: 8px; font-size: 13px; margin-bottom: 14px; }
</style>
