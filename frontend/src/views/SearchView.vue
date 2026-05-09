<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { searchPosts, searchUsers, type ApiPost, type ApiProfile } from '../api/client'
import PostCard from '../components/PostCard.vue'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const query = ref('')
const searchedQuery = ref('')
const activeType = ref<'users' | 'posts'>('users')
const users = ref<ApiProfile[]>([])
const posts = ref<ApiPost[]>([])
const isLoading = ref(false)
const errorMessage = ref('')
let activeSearchId = 0

const hasSearched = computed(() => searchedQuery.value.length > 0)
const isPostSearch = computed(() => activeType.value === 'posts')

function routeQueryValue(value: unknown) {
  if (Array.isArray(value)) {
    return typeof value[0] === 'string' ? value[0] : ''
  }
  return typeof value === 'string' ? value : ''
}

function routeTypeValue(value: unknown) {
  return routeQueryValue(value) === 'posts' ? 'posts' : 'users'
}

async function loadSearch(nextQuery: string, nextType: 'users' | 'posts') {
  query.value = nextQuery
  activeType.value = nextType
  const normalizedQuery = nextQuery.trim()
  searchedQuery.value = normalizedQuery
  errorMessage.value = ''

  if (!normalizedQuery) {
    activeSearchId += 1
    users.value = []
    posts.value = []
    isLoading.value = false
    return
  }

  const searchId = activeSearchId + 1
  activeSearchId = searchId
  isLoading.value = true
  try {
    if (nextType === 'posts') {
      const response = await searchPosts(normalizedQuery, authStore.token || undefined)
      if (searchId === activeSearchId) {
        posts.value = response.posts
        users.value = []
      }
    } else {
      const response = await searchUsers(normalizedQuery, authStore.token || undefined)
      if (searchId === activeSearchId) {
        users.value = response.users
        posts.value = []
      }
    }
  } catch (error) {
    if (searchId === activeSearchId) {
      users.value = []
      posts.value = []
      errorMessage.value = error instanceof Error ? error.message : 'Failed to search'
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
    query: {
      ...(normalizedQuery ? { q: normalizedQuery } : {}),
      ...(activeType.value === 'posts' ? { type: 'posts' } : {}),
    },
  })
}

function switchType(nextType: 'users' | 'posts') {
  void router.push({
    path: '/search',
    query: {
      ...(searchedQuery.value ? { q: searchedQuery.value } : {}),
      ...(nextType === 'posts' ? { type: 'posts' } : {}),
    },
  })
}

watch(
  () => [route.query.q, route.query.type],
  ([queryValue, typeValue]) => {
    void loadSearch(routeQueryValue(queryValue), routeTypeValue(typeValue))
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
          <input v-model="query" type="search" :placeholder="isPostSearch ? '搜索帖子' : '搜索用户'" autofocus />
        </label>
        <button class="solid-button" type="submit">搜索</button>
      </form>

      <section class="timeline">
        <nav class="feed-tabs" aria-label="Search result types">
          <button
            class="feed-tabs__item feed-tabs__button"
            :class="{ 'feed-tabs__item--active': !isPostSearch }"
            type="button"
            @click="switchType('users')"
          >
            用户
          </button>
          <button
            class="feed-tabs__item feed-tabs__button"
            :class="{ 'feed-tabs__item--active': isPostSearch }"
            type="button"
            @click="switchType('posts')"
          >
            帖子
          </button>
        </nav>

        <p v-if="isLoading" class="timeline__state">{{ isPostSearch ? 'Searching posts...' : 'Searching users...' }}</p>
        <p v-else-if="errorMessage" class="timeline__state timeline__state--error">{{ errorMessage }}</p>
        <p v-else-if="!hasSearched" class="timeline__state">{{ isPostSearch ? '输入关键词开始搜索帖子。' : '输入用户名开始搜索。' }}</p>
        <p v-else-if="!isPostSearch && users.length === 0" class="timeline__state">没有找到相关用户。</p>
        <p v-else-if="isPostSearch && posts.length === 0" class="timeline__state">没有找到相关帖子。</p>

        <div v-else-if="!isPostSearch" class="user-results">
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

        <div v-else class="timeline__list">
          <PostCard v-for="post in posts" :key="post.id" :post="post" />
        </div>
      </section>
    </main>

    <aside class="home-rail" aria-label="Search side panel">
      <section class="rail-section">
        <h2>搜索范围</h2>
        <p>用户和帖子共用这个入口。帖子结果按发布时间倒序展示。</p>
      </section>
    </aside>
  </div>
</template>
