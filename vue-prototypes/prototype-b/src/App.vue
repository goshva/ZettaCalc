<script setup lang="ts">
import { onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import AppHeader from '@/components/AppHeader.vue'
import AppFooter from '@/components/AppFooter.vue'

const auth = useAuthStore()
onMounted(() => { if (auth.token) auth.fetchMe() })
</script>
<template>
  <div class="app">
    <AppHeader />
    <main><RouterView /></main>
    <AppFooter />
  </div>
</template>
<style scoped>
.app { display: flex; flex-direction: column; min-height: 100vh; }
main { flex: 1; }
</style>
