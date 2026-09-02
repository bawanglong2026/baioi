<template>
  <div class="products-page min-h-screen bg-background text-foreground pt-20 pb-16">
    <div class="container mx-auto px-4">
      <div class="flex flex-col lg:flex-row gap-3 pt-8">
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
        <main class="flex-1">
          <div class="mb-3 flex h-10 items-center rounded-xl border bg-card px-3">
            <Search class="mr-2 h-4 w-4 shrink-0 text-muted-foreground" />
            <Input v-model="searchQuery" class="h-10 border-0 bg-transparent p-0 shadow-none focus-visible:ring-0" :placeholder="t('products.searchPlaceholder')" />
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
            <section v-for="group in visibleProductGroups" :key="group.key" class="space-y-2">
              <div class="flex items-center gap-2 px-1">
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

<style scoped>
@media (min-width: 768px) {
  .products-page :deep(.group) {
    flex-direction: row;
    height: 86px;
    min-height: 86px;
    align-items: stretch;
    border-radius: 0.75rem;
  }
  .products-page :deep(.group > div:first-child) {
    width: 64px;
    height: 64px;
    aspect-ratio: auto;
    flex: 0 0 64px;
    align-self: center;
    margin-left: 0.75rem;
    border-radius: 0.5rem;
  }
  .products-page :deep(.group > div:nth-child(2)) {
    padding: 0.5rem 0.75rem;
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    grid-template-rows: auto auto;
    column-gap: 0.75rem;
    align-items: center;
    min-width: 0;
  }
  .products-page :deep(.group > div:nth-child(2) > div:last-child) {
    grid-column: 2;
    grid-row: 1 / -1;
    margin-top: 0;
    min-width: 160px;
    padding-left: 0.75rem;
    justify-content: flex-end;
  }
  .products-page :deep(.group > div:nth-child(2) > div:last-child > div:first-child) {
    align-items: flex-end;
    white-space: nowrap;
  }
  .products-page :deep(.group > div:nth-child(2) > div:first-child),
  .products-page :deep(.group > div:nth-child(2) .price-label) {
    display: none;
  }
  .products-page :deep(.group > div:nth-child(2) h3) {
    font-size: 0.875rem;
    line-height: 1.25rem;
    margin-bottom: 0.25rem;
  }
  .products-page :deep(.group > div:nth-child(2) .theme-price-sm) {
    font-size: 0.875rem;
    line-height: 1.25rem;
  }
}

.line-clamp-1 {
  overflow: hidden;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 1;
  line-clamp: 1;
}
.line-clamp-2 {
  overflow: hidden;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  line-clamp: 2;
}
</style>
