<template>
  <div class="apple-scope min-h-screen">
    <header class="apple-nav sticky top-0 z-50 border-b border-black/5 bg-white/80 backdrop-blur-2xl dark:border-white/10 dark:bg-black/75">
      <div class="mx-auto flex h-12 max-w-[1120px] items-center px-5 sm:px-7">
        <RouterLink to="/" class="flex min-w-0 items-center gap-2 font-semibold tracking-tight" :title="brandName">
          <img v-if="brandLogo" :src="brandLogo" :alt="brandName" class="h-7 w-7 rounded-lg object-contain" />
          <span class="truncate text-sm">{{ brandName }}</span>
        </RouterLink>
        <nav class="mx-auto hidden items-center gap-7 text-xs text-foreground/70 md:flex">
          <RouterLink v-for="item in menuItems" :key="item.key" :to="item.path" class="transition-colors hover:text-foreground">{{ item.label }}</RouterLink>
        </nav>
        <div class="ml-auto flex items-center gap-1">
          <RouterLink to="/guest/orders" class="apple-icon-button" :aria-label="t('navbar.guestOrders')"><ClipboardList /></RouterLink>
          <button class="apple-icon-button" type="button" @click="toggleTheme"><Sun v-if="theme === 'dark'" /><Moon v-else /></button>
          <RouterLink to="/cart" class="apple-icon-button relative" :aria-label="t('navbar.cart')">
            <ShoppingBag />
            <span v-if="cartCount" class="absolute -right-0.5 -top-0.5 grid h-4 min-w-4 place-items-center rounded-full bg-[#0071e3] px-1 text-[9px] font-bold text-white">{{ cartCount }}</span>
          </RouterLink>
          <RouterLink :to="userAuthStore.isAuthenticated ? '/me' : '/auth/login'" class="apple-icon-button"><User /></RouterLink>
        </div>
      </div>
    </header>

    <main><slot /></main>

    <footer class="border-t border-black/5 bg-[#f5f5f7] dark:border-white/10 dark:bg-[#161617]">
      <div class="mx-auto max-w-[1120px] px-5 py-10 text-xs text-muted-foreground sm:px-7">
        <div class="flex flex-col justify-between gap-5 sm:flex-row sm:items-center">
          <span>© {{ year }} {{ brandName }}. {{ t('footer.rights') }}</span>
          <div class="flex gap-5">
            <RouterLink to="/privacy" class="hover:text-foreground">{{ t('footer.privacy') }}</RouterLink>
            <RouterLink to="/terms" class="hover:text-foreground">{{ t('footer.terms') }}</RouterLink>
          </div>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ClipboardList, Moon, ShoppingBag, Sun, User } from 'lucide-vue-next'
import { useAppStore } from '../../../stores/app'
import { useCartStore } from '../../../stores/cart'
import { useUserAuthStore } from '../../../stores/userAuth'
import { useNavConfig } from '../../../composables/useNavConfig'
import { useTheme } from '../../../utils/theme'
import { getImageUrl } from '../../../utils/image'
import '../styles/apple.css'

const { t } = useI18n()
const appStore = useAppStore()
const cartStore = useCartStore()
const userAuthStore = useUserAuthStore()
const { toggleTheme, theme } = useTheme()
const { secondaryNavItems } = useNavConfig()
const menuItems = secondaryNavItems
const cartCount = computed(() => cartStore.totalItems)
const brandName = computed(() => String(appStore.config?.brand?.site_name || '').trim() || 'Digital Store')
const brandLogo = computed(() => {
  const raw = String(appStore.config?.brand?.site_logo || '').trim()
  return raw ? getImageUrl(raw) : '/brand-logo.png'
})
const year = new Date().getFullYear()
</script>
