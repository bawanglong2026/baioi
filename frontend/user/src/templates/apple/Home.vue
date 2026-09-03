<template>
  <div class="min-h-screen">
    <section v-if="route.path === '/'" class="apple-hero overflow-hidden px-5 py-20 text-center sm:py-28">
      <div class="mx-auto max-w-4xl">
        <p class="mb-3 text-sm font-semibold text-[#0071e3]">{{ t('products.allCategories') }}</p>
        <h1 class="apple-hero-title font-semibold">{{ brandName }}</h1>
        <p class="mx-auto mt-6 max-w-2xl text-lg leading-relaxed text-muted-foreground sm:text-xl">{{ brandDescription }}</p>
        <a href="#apple-products" class="mt-8 inline-flex rounded-full bg-[#0071e3] px-6 py-3 text-sm font-semibold text-white transition hover:bg-[#0077ed]">{{ t('common.viewDetails') }}</a>
      </div>
    </section>

    <section id="apple-products" class="mx-auto max-w-[1120px] px-5 py-14 sm:px-7 sm:py-20">
      <div class="mb-9 flex flex-col gap-5 sm:flex-row sm:items-end sm:justify-between">
        <div class="sm:shrink-0">
          <p class="text-sm font-semibold text-[#0071e3]">Store</p>
          <h2 class="mt-1 text-3xl font-semibold tracking-[-.04em] sm:text-5xl">{{ t('nav.products') }}</h2>
        </div>
        <div class="apple-product-search relative sm:ml-auto sm:shrink-0">
          <Search class="pointer-events-none absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <input v-model="searchQuery" class="apple-panel h-11 w-full rounded-full bg-transparent pl-11 pr-4 text-sm outline-none transition focus:ring-2 focus:ring-[#0071e3]/30" :placeholder="t('products.searchPlaceholder')" />
        </div>
      </div>

      <div class="mb-8 flex gap-2 overflow-x-auto pb-2">
        <button class="shrink-0 rounded-full px-4 py-2 text-sm font-medium transition" :class="selectedCategory === null ? 'bg-[#1d1d1f] text-white dark:bg-white dark:text-black' : 'apple-panel'" @click="selectCategory(null)">{{ t('products.allCategories') }}</button>
        <button v-for="group in categoryGroups" :key="group.id" class="apple-panel shrink-0 rounded-full px-4 py-2 text-sm font-medium transition hover:opacity-75" :class="selectedCategory === group.id ? '!bg-[#0071e3] !text-white' : ''" @click="selectCategory(group.id)">{{ getLocalizedText(group.name) }}</button>
      </div>

      <div v-if="loading" class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <div v-for="i in 6" :key="i" class="apple-panel h-80 animate-pulse rounded-[28px]" />
      </div>
      <div v-else-if="products.length" :class="isListMode ? 'space-y-3' : 'grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3'">
        <ProductCard v-for="(product, index) in products" :key="product.id" :product="product" :index="index" :layout="isListMode ? 'list' : 'card'" @click="goToProduct" @quick-buy="openQuickBuy" />
      </div>
      <div v-else class="apple-panel rounded-[28px] py-20 text-center text-muted-foreground">{{ t('products.empty') }}</div>

      <PaginationNav class="mt-10" :current-page="currentPage" :total-pages="totalPages" :loading="loading" @change-page="changePage" />
    </section>

    <ProductQuickBuy v-if="quickBuyProduct" :product="quickBuyProduct" :visible="quickBuyVisible" @update:visible="quickBuyVisible = $event" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Search } from 'lucide-vue-next'
import { useAppStore } from '../../stores/app'
import { useProductList } from '../../composables/useProductList'
import { useLocalized } from '../../composables/useProduct'
import ProductCard from '../../components/ProductCard.vue'
import ProductQuickBuy from '../../components/ProductQuickBuy.vue'
import PaginationNav from '../../components/PaginationNav.vue'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const { getLocalizedText } = useLocalized()
const { loading, products, selectedCategory, searchQuery, currentPage, totalPages, categoryGroups, selectCategory, changePage, initialize, cleanup } = useProductList({ pageSize: 12, homeRouteName: route.path === '/' ? 'home' : 'products' })
const isListMode = computed(() => appStore.config?.template_mode === 'list')
const brandName = computed(() => String(appStore.config?.brand?.site_name || '').trim() || 'Digital Store')
const brandDescription = computed(() => {
  const value = appStore.config?.brand?.site_description
  if (value && typeof value === 'object') return getLocalizedText(value)
  return t('home.hero.subtitle')
})
const quickBuyProduct = ref<any>(null)
const quickBuyVisible = ref(false)
const openQuickBuy = (product: any) => { quickBuyProduct.value = product; quickBuyVisible.value = true }
const goToProduct = (slug: string) => router.push(`/products/${slug}`)
onMounted(initialize)
onUnmounted(cleanup)
</script>
