<script setup lang="ts">
import { onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import { useCartStore } from '@/stores/cart'
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

      <div v-if="cart.loading" class="muted">Загрузка…</div>
      <div v-else-if="items.length === 0" class="empty">Корзина пуста</div>
      <div v-else>
        <div class="items">
          <div v-for="item in items" :key="item.sku" class="item">
            <div class="item-img">
              <img v-if="item.image" :src="imgUrl(item.image)" :alt="item.title" />
            </div>
            <div class="item-info">
              <strong>{{ item.title }}</strong>
              <p class="muted">{{ item.brand }} · {{ item.sku }}</p>
            </div>
            <div class="item-qty">
              <span>×{{ item.qty }}</span>
            </div>
            <div class="item-price">
              <b>{{ formatPrice(item.price * item.qty) }}</b>
            </div>
          </div>
        </div>

        <div class="total-bar">
          <span>Итого:</span>
          <b>{{ formatPrice(total_rub) }}</b>
        </div>

        <div class="actions">
          <RouterLink to="/cpr" class="btn red">Оформить заказ</RouterLink>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.empty { padding: 40px 0; text-align: center; }
.items { background: #fff; border: 1px solid var(--line); border-radius: var(--radius); overflow: hidden; }
.item { display: flex; align-items: center; gap: 20px; padding: 18px 24px; border-bottom: 1px solid var(--line); }
.item-img { width: 72px; height: 54px; border-radius: 8px; overflow: hidden; flex-shrink: 0; background: #f2f0ec; }
.item-img img { width: 100%; height: 100%; object-fit: cover; }
.item-info { flex: 1; }
.item-info strong { display: block; margin-bottom: 4px; }
.item-qty { font-size: 14px; color: var(--muted); }
.item-price { font-size: 18px; font-weight: 600; }
.total-bar { display: flex; justify-content: space-between; padding: 20px 0; font-size: 20px; font-weight: 600; }
.actions { margin-top: 20px; }
</style>
