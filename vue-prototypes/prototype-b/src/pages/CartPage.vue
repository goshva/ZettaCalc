<script setup lang="ts">
import { onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import { useCartStore } from '@/stores'
import { formatPrice } from '@/api/client'
import { imgUrl } from '@/composables/useImg'

const cart = useCartStore()
const { items, total_rub } = storeToRefs(cart)
onMounted(() => cart.fetch())
</script>
<template>
  <section class="section">
    <div class="wrap">
      <h2>Корзина</h2>
      <p v-if="cart.loading" class="muted">Загрузка…</p>
      <p v-else-if="items.length === 0" class="muted" style="padding: 40px 0; text-align: center;">Корзина пуста</p>
      <div v-else>
        <div class="items">
          <div v-for="item in items" :key="item.sku" class="item">
            <div class="item-img">
              <img v-if="item.image" :src="imgUrl(item.image)" :alt="item.title" />
            </div>
            <div class="info"><strong>{{ item.title }}</strong><p class="muted">{{ item.brand }} · {{ item.sku }}</p></div>
            <div class="qty">×{{ item.qty }}</div>
            <div class="price"><b>{{ formatPrice(item.price * item.qty) }}</b></div>
          </div>
        </div>
        <div class="total-bar"><span>Итого:</span><b>{{ formatPrice(total_rub) }}</b></div>
        <RouterLink to="/cpr" class="btn gold">Оформить заказ</RouterLink>
      </div>
    </div>
  </section>
</template>
<style scoped>
.items { background: #fff; border-radius: var(--radius); overflow: hidden; }
.item { display: flex; align-items: center; gap: 16px; padding: 16px 20px; border-bottom: 1px solid var(--line); }
.item-img { width: 64px; height: 48px; border-radius: 8px; overflow: hidden; flex-shrink: 0; background: #f5f0eb; }
.item-img img { width: 100%; height: 100%; object-fit: cover; }
.info { flex: 1; }
.info strong { display: block; margin-bottom: 2px; }
.qty { font-size: 14px; color: #5f584f; }
.price { font-size: 16px; font-weight: 600; }
.total-bar { display: flex; justify-content: space-between; padding: 20px 0; font-size: 18px; font-weight: 600; }
</style>
