<template>
  <Card
    class="product-card group relative flex flex-col overflow-hidden rounded-2xl transition-all theme-slide-up"
    :class="[
      layout === 'card' ? 'h-full' : 'product-card--list',
      isSoldOut(product)
        ? 'cursor-default opacity-85 grayscale-[0.25] saturate-50 border-destructive/30'
        : 'cursor-pointer hover:-translate-y-1 hover:border-primary/30 hover:shadow-lg',
    ]"
    :style="{ animationDelay: `${index * animationStep}ms` }"
    @click="$emit('click', product.slug)">
    <!-- Image Area -->
    <div class="product-card-image relative aspect-[4/3] shrink-0 overflow-hidden bg-muted">
      <div
        class="absolute inset-0 z-10 transition-colors duration-300"
        :class="isSoldOut(product) ? 'bg-black/15' : 'bg-black/15 group-hover:bg-black/5'"
      ></div>
      <img v-if="displayImageSrc && !imageErrored" :src="displayImageSrc"
        :alt="getLocalizedText(product.title)" loading="lazy" decoding="async"
        class="w-full h-full object-cover transform transition-transform duration-700 ease-out"
        :class="[
          isSoldOut(product) ? 'grayscale brightness-75' : 'group-hover:scale-105',
        ]"
        @error="handleImageError" />
      <div v-else class="w-full h-full flex items-center justify-center text-muted-foreground" role="img"
        :aria-label="getLocalizedText(product.title)">
        <ImageIcon class="w-8 h-8 md:w-12 md:h-12" :stroke-width="1.5" aria-hidden="true" />
      </div>

      <div v-if="isSoldOut(product)" class="absolute inset-0 z-20 flex items-center justify-center bg-black/45">
        <span v-if="layout === 'list'" class="text-[10px] font-bold text-white">{{ t('products.stockStatus.outOfStock') }}</span>
      </div>
      <Badge
        v-if="isSoldOut(product) && layout === 'card'"
        variant="destructive"
        size="xs"
        class="absolute left-2 top-2 md:left-4 md:top-4 z-30 tracking-wider shadow-sm"
      >
        {{ t('products.stockStatus.outOfStock') }}
      </Badge>

      <!-- Tags -->
      <div v-if="!isSoldOut(product) && product.tags && product.tags.length > 0"
        class="absolute top-2 right-2 md:top-4 md:right-4 z-20 flex flex-wrap gap-1 md:gap-2 justify-end">
        <span v-for="(tag, tagIndex) in product.tags.slice(0, maxTags)" :key="tagIndex"
          class="inline-flex items-center rounded-md border border-white/25 bg-black/55 px-2 md:px-3 py-0.5 md:py-1 text-xs font-medium text-white backdrop-blur-sm">
          {{ tag }}
        </span>
      </div>
    </div>

    <!-- Content Area -->
    <div class="product-card-body relative z-20 flex flex-1 flex-col p-3 md:p-4">
      <!-- Title and badges stay together so the compact desktop list is vertically centered. -->
      <div class="product-card-info flex flex-col justify-center">
        <div v-if="layout === 'card' && product.category?.name" class="product-card-category-card mb-1 truncate text-xs uppercase tracking-wider text-muted-foreground md:mb-2">
          {{ t('products.categoryLabel') }} · {{ getLocalizedText(product.category.name) }}
        </div>

        <div class="product-card-title-row flex min-w-0 items-center gap-1.5">
          <h3
            class="product-card-title min-w-0 flex-1 truncate text-foreground transition-colors"
            :class="layout === 'list' ? 'text-xs font-semibold sm:text-sm md:text-sm' : 'text-sm font-bold md:text-lg'"
          >
            {{ getLocalizedText(product.title) }}
          </h3>
        </div>

        <!-- Badges -->
        <div
          class="product-card-badges flex flex-wrap items-center gap-1 md:gap-2"
          :class="layout === 'list' ? 'mb-0' : 'mb-2 md:mb-3'"
        >
        <!-- Mobile: show only fulfillment type badge -->
        <Badge
          class="sm:hidden"
          size="xs"
          :variant="product.fulfillment_type === 'auto' ? 'info' : 'neutral'"
        >
          {{ getFulfillmentTypeLabel(product.fulfillment_type) }}
        </Badge>
        <Badge v-if="layout === 'list' && isSoldOut(product)" class="sm:hidden" size="xs" variant="danger">
          {{ getStockStatusLabel(product) }}
        </Badge>

        <!-- Desktop: show all badges -->
        <Badge
          class="hidden sm:inline-flex"
          size="xs"
          :variant="product.purchase_type === 'guest' ? 'warning' : 'success'"
        >
          <UserPlus v-if="product.purchase_type === 'guest'" class="h-3 w-3" />
          <Lock v-else class="h-3 w-3" />
          {{ getPurchaseTypeLabel(product.purchase_type) }}
        </Badge>

        <Badge
          class="hidden sm:inline-flex"
          size="xs"
          :variant="product.fulfillment_type === 'auto' ? 'info' : 'neutral'"
        >
          <Zap v-if="product.fulfillment_type === 'auto'" class="h-3 w-3" />
          <Pencil v-else class="h-3 w-3" />
          {{ getFulfillmentTypeLabel(product.fulfillment_type) }}
        </Badge>

        <Badge class="hidden sm:inline-flex" size="xs" :variant="getStockBadgeVariant(product.stock_status)">
          {{ getStockStatusLabel(product) }}
        </Badge>
        </div>

        <p :class="layout === 'list' ? 'hidden' : 'hidden md:mb-6 md:block'" class="product-card-description text-sm text-muted-foreground line-clamp-2">
          {{ getLocalizedText(product.description) }}
        </p>
      </div>

      <div class="product-card-price-row mt-auto flex items-center justify-between border-t pt-2 md:pt-4">
        <div class="product-card-price flex flex-col">
          <span class="price-label hidden text-xs uppercase tracking-wider text-muted-foreground md:block">{{ t('products.price') }}</span>
          <span
            v-if="hasPromotionPrice(product)"
            class="theme-price-sm theme-price-promotion"
            :aria-label="t('products.promotionPriceAria', { price: formatPrice(getPromotionPriceAmount(product), siteCurrency) })"
          >
            {{ formatPrice(getPromotionPriceAmount(product), siteCurrency) }}
          </span>
          <span
            v-else
            class="theme-price-sm"
            :aria-label="t('products.priceAria', { price: formatPrice(product.price_amount, siteCurrency) })"
          >
            {{ formatPrice(product.price_amount, siteCurrency) }}
          </span>
          <div v-if="hasPromotionPrice(product)" class="mt-0.5 flex flex-wrap items-center gap-1.5">
            <span
              class="hidden md:inline text-xs text-muted-foreground opacity-80 line-through"
              :aria-label="t('products.originalPriceAria', { price: formatPrice(product.price_amount, siteCurrency) })"
            >{{ formatPrice(product.price_amount, siteCurrency) }}</span>
            <Badge variant="danger" size="xs">
              {{ t('products.promotionTag') }}
            </Badge>
          </div>
          <div v-else-if="hasWholesalePrices(product)" class="mt-0.5 flex flex-wrap items-center gap-1.5">
            <Badge variant="success" size="xs">
              {{ t('products.wholesaleTag') }}
            </Badge>
          </div>
          <div v-else-if="hasPromotionRules(product)" class="mt-0.5 flex flex-wrap items-center gap-1.5">
            <Badge variant="warning" size="xs">
              {{ t('products.promotionBadge') }}
            </Badge>
          </div>
        </div>

        <div class="product-card-actions flex items-center gap-2">
          <!-- Quick buy cart button -->
          <Button
            type="button"
            variant="outline"
            size="icon"
            class="h-8 w-8 md:h-9 md:w-9"
            :aria-label="t('products.quickBuyAria')"
            :disabled="isSoldOut(product)"
            @click.stop="$emit('quickBuy', product)"
          >
            <ShoppingCart class="h-4 w-4" />
          </Button>
          <!-- Desktop: view details -->
          <span
            class="hidden md:flex text-xs uppercase font-bold transition-colors items-center gap-1"
            :class="isSoldOut(product)
              ? 'text-destructive/90'
              : 'text-muted-foreground group-hover:text-foreground'">
            <ArrowRight class="w-4 h-4 transition-transform" :class="isSoldOut(product) ? '' : 'group-hover:translate-x-1'" />
          </span>
          <!-- Mobile: arrow only -->
          <ChevronRight v-if="layout === 'card'" class="h-4 w-4 text-muted-foreground md:hidden" />
        </div>
      </div>
    </div>
  </Card>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { computed, ref, watch } from 'vue'
import { ArrowRight, ChevronRight, Image as ImageIcon, Lock, Pencil, ShoppingCart, UserPlus, Zap } from 'lucide-vue-next'
import { getFirstImageUrl, getImageUrl } from '../utils/image'
import { useLocalized, useProductLabels } from '../composables/useProduct'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'

const props = withDefaults(defineProps<{
  product: any
  index?: number
  maxTags?: number
  animationStep?: number
  layout?: 'card' | 'list'
}>(), {
  index: 0,
  maxTags: 2,
  animationStep: 50,
  layout: 'card',
})

defineEmits<{
  click: [slug: string]
  quickBuy: [product: any]
}>()

const { t } = useI18n()
const { getLocalizedText, siteCurrency, formatPrice } = useLocalized()
const { getPurchaseTypeLabel, getFulfillmentTypeLabel, getStockBadgeVariant, getStockStatusLabel, isSoldOut, hasPromotionPrice, getPromotionPriceAmount, hasPromotionRules, hasWholesalePrices } = useProductLabels()
const layout = computed(() => props.layout)

const imageErrored = ref(false)
const attemptIdx = ref(0)

const imageCandidates = computed<string[]>(() => {
  const arr: string[] = []
  const primary = getFirstImageUrl(props.product?.images)
  if (primary) arr.push(primary)
  const categoryIcon = props.product?.category?.icon
  if (categoryIcon) {
    const resolved = getImageUrl(categoryIcon)
    if (resolved && resolved !== primary) arr.push(resolved)
  }
  return arr
})

const displayImageSrc = computed(() => imageCandidates.value[attemptIdx.value] ?? '')

watch(imageCandidates, () => {
  attemptIdx.value = 0
  imageErrored.value = false
}, { deep: true })

const handleImageError = () => {
  if (attemptIdx.value < imageCandidates.value.length - 1) {
    attemptIdx.value++
  } else {
    imageErrored.value = true
  }
}
</script>

<style scoped>
/* The products view uses the same responsive horizontal structure at every width. */
.product-card--list {
  flex-direction: row;
  align-items: center;
  min-height: 58px;
  border-radius: 0.75rem;
}

.product-card--list .product-card-image {
  width: 44px;
  height: 44px;
  flex: 0 0 44px;
  margin: 6px;
  aspect-ratio: auto;
  border-radius: 0.5rem;
}

.product-card--list .product-card-body {
  min-width: 0;
  flex: 1 1 auto;
  flex-direction: row;
  align-items: center;
  padding: 6px 2px 6px 0;
}

.product-card--list .product-card-info {
  min-width: 0;
  flex: 1 1 0%;
  justify-content: center;
  gap: 2px;
}

.product-card--list .product-card-title-row,
.product-card--list .product-card-badges {
  gap: 4px;
}

.product-card--list .product-card-description,
.product-card--list .price-label {
  display: none;
}

.product-card--list .product-card-price-row {
  flex: 0 0 auto;
  align-items: center;
  justify-content: flex-end;
  gap: 4px;
  margin: 0;
  padding: 0 6px 0 0;
  border-top: 0;
}

.product-card--list .product-card-price {
  align-items: flex-end;
  white-space: nowrap;
}

.product-card--list .theme-price-sm {
  font-size: 0.75rem;
  line-height: 1rem;
}

.product-card--list .product-card-actions {
  gap: 4px;
}

.product-card--list .product-card-actions > button {
  width: 28px;
  height: 28px;
  padding: 0;
}

@media (min-width: 640px) {
  .product-card--list {
    height: 86px;
    min-height: 86px;
  }

  .product-card--list .product-card-image {
    width: 64px;
    height: 64px;
    flex: 0 0 64px;
    margin: 10px;
    aspect-ratio: auto;
    border-radius: 0.5rem;
  }

  .product-card--list .product-card-body {
    min-width: 0;
    flex: 1 1 auto;
    flex-direction: row;
    align-items: center;
    padding: 8px 4px 8px 0;
  }

  .product-card--list .product-card-info {
    min-width: 0;
    flex: 1 1 auto;
    justify-content: center;
    gap: 2px;
  }

  .product-card--list .product-card-title-row {
    gap: 6px;
  }

  .product-card--list .product-card-badges {
    gap: 4px;
  }

  .product-card--list .product-card-description,
  .product-card--list .price-label {
    display: none;
  }

  .product-card--list .product-card-price-row {
    flex: 0 0 auto;
    align-items: center;
    justify-content: flex-end;
    gap: 12px;
    margin: 0;
    padding: 0 16px 0 0;
    border-top: 0;
  }

  .product-card--list .product-card-price {
    align-items: flex-end;
    white-space: nowrap;
  }

  .product-card--list .theme-price-sm {
    font-size: 0.875rem;
    line-height: 1.25rem;
  }

  .product-card--list .product-card-actions {
    gap: 12px;
  }

  .product-card--list .product-card-actions > button {
    width: 32px;
    height: 32px;
    padding: 0;
  }
}
</style>
