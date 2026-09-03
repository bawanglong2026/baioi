<template>
  <div class="products-page min-h-screen bg-background pb-6 pt-24 text-foreground">
    <div class="container mx-auto px-4">
      <div class="flex flex-col gap-6 lg:flex-row">
        <CategorySidebar
          :categories="categoryGroups"
          :selected-category="selectedCategory"
          :expanded-parent-ids="expandedParentIds"
          :show-drawer="showFilterDrawer"
          :show-search="false"
          :compact="true"
          :search-query="searchQuery"
          @select-category="selectCategory"
          @toggle-parent="toggleParentCategory"
          @update:show-drawer="showFilterDrawer = $event"
          @update:search-query="searchQuery = $event"
          @clear-search="clearSearch"
        />

        <!-- Main Content - Products Grid -->
        <main class="min-w-0 flex-1">
          <div class="relative mb-4">
            <div class="pointer-events-none absolute inset-y-0 left-3.5 flex items-center">
              <Search class="h-4 w-4 text-muted-foreground" />
            </div>
            <Input v-model="searchQuery" class="h-10 rounded-xl pl-10 pr-10 shadow-none" :placeholder="t('products.searchPlaceholder')" />
          </div>
          <!-- Loading Skeleton -->
          <div v-if="loading" class="grid grid-cols-1 gap-3 md:gap-3">
            <div v-for="i in 6" :key="i"
              class="rounded-2xl border bg-card overflow-hidden flex flex-col">
              <div class="h-36 md:h-56 theme-skeleton"></div>
              <div class="p-3 md:p-5 space-y-3">
                <div class="h-3 w-16 rounded theme-skeleton"></div>
                <div class="h-5 w-3/4 rounded theme-skeleton"></div>
                <div class="flex gap-2">
                  <div class="h-5 w-14 rounded-full theme-skeleton"></div>
                  <div class="h-5 w-14 rounded-full theme-skeleton"></div>
                </div>
                <div class="h-3 w-full rounded theme-skeleton"></div>
                <div class="h-3 w-2/3 rounded theme-skeleton"></div>
                <div class="border-t pt-3 flex justify-between items-center">
                  <div class="h-6 w-20 rounded theme-skeleton"></div>
                  <div class="h-4 w-16 rounded theme-skeleton"></div>
                </div>
              </div>
            </div>
          </div>

          <!-- Products Grid -->
          <div v-else-if="products.length > 0" class="space-y-6">
            <section v-for="group in visibleProductGroups" :key="group.key">
              <div class="mb-3 flex items-center gap-2 px-0.5">
                <span class="h-5 w-1 rounded-full bg-primary"></span>
                <img v-if="group.icon" :src="getImageUrl(group.icon)" :alt="group.name" class="h-5 w-5 rounded object-cover" />
                <h2 class="truncate text-base font-bold leading-5 text-foreground">{{ group.name }}</h2>
                <span class="text-sm text-muted-foreground">({{ group.products.length }})</span>
              </div>
              <div class="grid grid-cols-1 gap-2">
                <ProductCard
                  v-for="(product, idx) in group.products"
                  :key="product.id"
                  :product="product"
                  :index="idx"
                  layout="list"
                  :max-tags="isMobileGrid ? 1 : 2"
                  :animation-step="30"
                  @click="goToProduct"
                  @quick-buy="openQuickBuy"
                />
              </div>
            </section>

            <PaginationNav
              :current-page="currentPage"
              :total-pages="totalPages"
              :loading="loading"
              @change-page="changePage"
            />
          </div>

          <!-- Empty State -->
          <EmptyState
            v-else
            variant="soft"
            size="lg"
            :icon="(searchQuery || selectedCategory) ? 'search' : 'package'"
            :title="(searchQuery || selectedCategory) ? t('products.emptyFiltered') : t('products.empty')"
          >
            <template v-if="searchQuery || selectedCategory" #action>
              <Button variant="secondary" @click="clearSearch(); selectCategory(null)">
                {{ t('products.clearFilters') }}
              </Button>
            </template>
          </EmptyState>
        </main>
      </div>
    </div>

    <ProductQuickBuy
      v-if="quickBuyProduct"
      :product="quickBuyProduct"
      :visible="quickBuyVisible"
      @update:visible="quickBuyVisible = $event"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useProductList } from '../composables/useProductList'
import { usePageSeo } from '../composables/usePageSeo'
import { useLocalized } from '../composables/useProduct'
import { getImageUrl } from '../utils/image'
import ProductCard from '../components/ProductCard.vue'
import ProductQuickBuy from '../components/ProductQuickBuy.vue'
import CategorySidebar from '../components/CategorySidebar.vue'
import PaginationNav from '../components/PaginationNav.vue'
import EmptyState from '../components/EmptyState.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Search } from 'lucide-vue-next'

const router = useRouter()
const { t } = useI18n()

const {
  loading,
  products,
  selectedCategory,
  searchQuery,
  currentPage,
  totalPages,
  showFilterDrawer,
  expandedParentIds,
  categoryGroups,
  categoryMap,
  selectCategory,
  toggleParentCategory,
  changePage,
  clearSearch,
  initialize,
  cleanup,
} = useProductList({ pageSize: 12, homeRouteName: 'products' })

// ==================== SEO ====================
const route = useRoute()
const { getLocalizedText } = useLocalized()
const seoCategoryName = computed(() => {
  if (!selectedCategory.value) return ''
  const cat = categoryMap.value.get(selectedCategory.value)
  return cat ? getLocalizedText(cat.name) : ''
})
const visibleProductGroups = computed(() => {
  const groups = new Map<string, { key: string; name: string; icon?: string; products: any[] }>()
  products.value.forEach((product) => {
    const category = product.category
    const key = String(category?.id ?? 'uncategorized')
    if (!groups.has(key)) {
      groups.set(key, {
        key,
        name: category?.name ? getLocalizedText(category.name) : t('products.allCategories'),
        icon: category?.icon,
        products: [],
      })
    }
    groups.get(key)!.products.push(product)
  })
  return Array.from(groups.values())
})
usePageSeo({
  canonicalPath: () => route.path,
  title: () => {
    if (route.name === 'category-products') {
      return seoCategoryName.value || t('nav.products')
    }
    return t('nav.products')
  },
})

const quickBuyProduct = ref<any>(null)
const quickBuyVisible = ref(false)

const openQuickBuy = (product: any) => {
  quickBuyProduct.value = product
  quickBuyVisible.value = true
}

// Detect mobile 2-col grid (< md breakpoint)
const isMobileGrid = ref(window.innerWidth < 768)
const handleResize = () => {
  isMobileGrid.value = window.innerWidth < 768
}

const goToProduct = (slug: string) => {
  router.push(`/products/${slug}`)
}

onMounted(async () => {
  window.addEventListener('resize', handleResize, { passive: true })
  await initialize()
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  cleanup()
})
</script>
