<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { searchUsers, type ApiProfile } from '../api/client'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const query = ref('')
const searchedQuery = ref('')
const users = ref<ApiProfile[]>([])
const isLoading = ref(false)
const errorMessage = ref('')
let activeSearchId = 0

const hasSearched = computed(() => searchedQuery.value.length > 0)

function routeQueryValue(value: unknown) {
  if (Array.isArray(value)) {
    return typeof value[0] === 'string' ? value[0] : ''
  }
  return typeof value === 'string' ? value : ''
}

async function loadUsers(nextQuery: string) {
  query.value = nextQuery
  const normalizedQuery = nextQuery.trim()
  searchedQuery.value = normalizedQuery
  errorMessage.value = ''

  if (!normalizedQuery) {
    activeSearchId += 1
    users.value = []
    isLoading.value = false
    return
  }

  const searchId = activeSearchId + 1
  activeSearchId = searchId
  isLoading.value = true
  try {
    const response = await searchUsers(normalizedQuery, authStore.token || undefined)
    if (searchId === activeSearchId) {
      users.value = response.users
    }
  } catch (error) {
    if (searchId === activeSearchId) {
      users.value = []
      errorMessage.value = error instanceof Error ? error.message : 'Failed to search users'
    }
  } finally {
    if (searchId === activeSearchId) {
      isLoading.value = false
    }
  }
}

function submitSearch() {
  const normalizedQuery = query.value.trim()
  void router.push({
    path: '/search',
    query: normalizedQuery ? { q: normalizedQuery } : {},
  })
}

watch(
  () => route.query.q,
  (value) => {
    void loadUsers(routeQueryValue(value))
  },
  { immediate: true },
)
</script>

<template>
  <div class="community-home search-page">
    <main class="home-feed search-feed">
      <header class="home-feed__bar">
        <h1>搜索</h1>
      </header>

      <form class="search-panel" @submit.prevent="submitSearch">
        <label class="search-box search-box--large">
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <circle cx="11" cy="11" r="7" />
            <path d="m16 16 4 4" />
          </svg>
          <input v-model="query" type="search" placeholder="搜索用户" autofocus />
        </label>
        <button class="solid-button" type="submit">搜索</button>
      </form>

      <section class="timeline">
        <nav class="feed-tabs" aria-label="Search result types">
          <span class="feed-tabs__item feed-tabs__item--active">用户</span>
          <span class="feed-tabs__item feed-tabs__item--muted">帖子稍后</span>
        </nav>

        <p v-if="isLoading" class="timeline__state">Searching users...</p>
        <p v-else-if="errorMessage" class="timeline__state timeline__state--error">{{ errorMessage }}</p>
        <p v-else-if="!hasSearched" class="timeline__state">输入用户名开始搜索。</p>
        <p v-else-if="users.length === 0" class="timeline__state">没有找到相关用户。</p>

        <div v-else class="user-results">
          <RouterLink v-for="user in users" :key="user.username" class="user-result" :to="`/u/${user.username}`">
            <div class="user-result__avatar">{{ user.username.slice(0, 2).toUpperCase() }}</div>
            <div class="user-result__body">
              <div class="user-result__header">
                <span class="user-result__name">@{{ user.username }}</span>
                <span v-if="user.following" class="user-result__badge">已关注</span>
              </div>
              <p>{{ user.bio || user.status_text || '还没有简介。' }}</p>
              <div class="user-result__meta">
                <span>{{ user.followers_count }} followers</span>
                <span>{{ user.post_count }} posts</span>
              </div>
            </div>
          </RouterLink>
        </div>
      </section>
    </main>

    <aside class="home-rail" aria-label="Search side panel">
      <section class="rail-section">
        <h2>第一版范围</h2>
        <p>当前只搜索用户。帖子搜索后续会接入同一个入口。</p>
      </section>
    </aside>
  </div>
</template>
