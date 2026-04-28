<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { fetchUnreadCount } from './api/client'
import { useAuthStore } from './stores/auth'

const authStore = useAuthStore()
const usernameLabel = computed(() => String(authStore.user?.username ?? 'operator'))
const unreadCount = ref(0)
const theme = ref<'light' | 'dark'>('light')
const isDarkMode = computed(() => theme.value === 'dark')

function applyTheme(nextTheme: 'light' | 'dark') {
  theme.value = nextTheme
  document.documentElement.dataset.theme = nextTheme
  window.localStorage.setItem('decoy-theme', nextTheme)
}

function toggleTheme() {
  applyTheme(isDarkMode.value ? 'light' : 'dark')
}

async function refreshUnreadCount() {
  if (!authStore.token) {
    unreadCount.value = 0
    return
  }
  try {
    const response = await fetchUnreadCount(authStore.token)
    unreadCount.value = response.total
  } catch {
    unreadCount.value = 0
  }
}

watch(
  () => authStore.token,
  () => {
    void refreshUnreadCount()
  },
  { immediate: true },
)

onMounted(() => {
  const storedTheme = window.localStorage.getItem('decoy-theme')
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
  applyTheme(storedTheme === 'dark' || storedTheme === 'light' ? storedTheme : prefersDark ? 'dark' : 'light')
  void refreshUnreadCount()
})
</script>

<template>
  <div class="app-frame">
    <aside class="app-sidebar">
      <RouterLink class="app-sidebar__brand" to="/">
        <svg class="app-sidebar__logo" viewBox="0 0 168 34" role="img" aria-label="Decoy Club">
          <path class="app-sidebar__logo-mark" d="M17 4 29 11v14l-12 7-12-7V11L17 4z" />
          <path class="app-sidebar__logo-line" d="M17 9v18" />
          <path class="app-sidebar__logo-line" d="M9.5 13.5 17 18l7.5-4.5" />
          <circle class="app-sidebar__logo-dot" cx="17" cy="18" r="3.4" />
          <text x="42" y="17" class="app-sidebar__logo-main">DECOY</text>
          <text x="42" y="29" class="app-sidebar__logo-sub">CLUB</text>
        </svg>
      </RouterLink>

      <nav class="app-sidebar__nav" aria-label="Primary">
        <RouterLink class="app-nav-link" to="/">
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <path d="M3 10.8 12 3l9 7.8" />
            <path d="M5.5 9.5V21h13V9.5" />
            <path d="M9.5 21v-6h5v6" />
          </svg>
          <span>广场</span>
        </RouterLink>

        <RouterLink class="app-nav-link" to="/search">
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <circle cx="11" cy="11" r="7" />
            <path d="m16 16 4 4" />
          </svg>
          <span>搜索</span>
        </RouterLink>

        <template v-if="authStore.token">
          <RouterLink class="app-nav-link" to="/following">
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M5 12h14" />
              <path d="M12 5v14" />
              <path d="M7 5h10l-2 14H5z" />
            </svg>
            <span>关注</span>
          </RouterLink>

          <RouterLink class="app-nav-link app-nav-link--badge" to="/notifications">
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M18 8a6 6 0 0 0-12 0c0 7-3 7-3 9h18c0-2-3-2-3-9" />
              <path d="M10 21h4" />
            </svg>
            <span>消息</span>
            <span v-if="unreadCount" class="app-nav-link__badge">{{ unreadCount }}</span>
          </RouterLink>

          <RouterLink
            v-if="authStore.user?.username"
            class="app-nav-link"
            :to="`/u/${String(authStore.user.username)}`"
          >
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M20 21a8 8 0 0 0-16 0" />
              <circle cx="12" cy="8" r="4" />
            </svg>
            <span>主页</span>
          </RouterLink>
        </template>

        <template v-else>
          <RouterLink class="app-nav-link" to="/login">
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4" />
              <path d="M10 17l5-5-5-5" />
              <path d="M15 12H3" />
            </svg>
            <span>登录</span>
          </RouterLink>
          <RouterLink class="app-nav-link" to="/register">
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
              <circle cx="9" cy="7" r="4" />
              <path d="M19 8v6" />
              <path d="M22 11h-6" />
            </svg>
            <span>注册</span>
          </RouterLink>
        </template>
      </nav>

      <button
        class="app-sidebar__theme"
        :class="{ 'app-sidebar__theme--dark': isDarkMode }"
        type="button"
        role="switch"
        :aria-checked="isDarkMode"
        :aria-label="isDarkMode ? '切换到浅色模式' : '切换到暗色模式'"
        @click="toggleTheme"
      >
        <svg v-if="isDarkMode" class="app-sidebar__theme-icon" viewBox="0 0 24 24" aria-hidden="true">
          <circle cx="12" cy="12" r="4" />
          <path d="M12 2v2" />
          <path d="M12 20v2" />
          <path d="m4.93 4.93 1.41 1.41" />
          <path d="m17.66 17.66 1.41 1.41" />
          <path d="M2 12h2" />
          <path d="M20 12h2" />
          <path d="m6.34 17.66-1.41 1.41" />
          <path d="m19.07 4.93-1.41 1.41" />
        </svg>
        <svg v-else class="app-sidebar__theme-icon" viewBox="0 0 24 24" aria-hidden="true">
          <path d="M21 12.8A8.6 8.6 0 1 1 11.2 3 6.8 6.8 0 0 0 21 12.8z" />
        </svg>
        <span class="app-sidebar__theme-track" aria-hidden="true">
          <span class="app-sidebar__theme-thumb" />
        </span>
      </button>

      <div v-if="authStore.token" class="app-sidebar__account">
        <span class="app-sidebar__user">{{ usernameLabel }}</span>
        <button class="app-sidebar__logout" @click="authStore.logout()">退出</button>
      </div>
    </aside>

    <main class="app-main">
      <router-view />
    </main>
  </div>
</template>
