<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { fetchPublicPosts, fetchTrendingTopics, type ApiPost, type Topic } from '../api/client'
import PostCard from '../components/PostCard.vue'
import PostComposer from '../components/PostComposer.vue'
import { useAuthStore } from '../stores/auth'

const authStore = useAuthStore()
const router = useRouter()
const posts = ref<ApiPost[]>([])
const isLoading = ref(true)
const errorMessage = ref('')
const topics = ref<Topic[]>([])
const searchQuery = ref('')

const isLoggedIn = computed(() => Boolean(authStore.token))

async function loadPosts() {
  isLoading.value = true
  errorMessage.value = ''
  try {
    const response = await fetchPublicPosts()
    posts.value = response.posts
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Failed to load timeline'
  } finally {
    isLoading.value = false
  }
}

async function loadTopics() {
  try {
    const response = await fetchTrendingTopics(8)
    topics.value = response.topics
  } catch {
    topics.value = []
  }
}

function submitSearch() {
  const normalizedQuery = searchQuery.value.trim()
  void router.push({
    path: '/search',
    query: normalizedQuery ? { q: normalizedQuery } : {},
  })
}

onMounted(() => {
  void loadPosts()
  void loadTopics()
})
</script>

<template>
  <div class="community-home">
    <main class="home-feed">
      <header class="home-feed__bar">
        <h1>Decoy 广场</h1>
        <button class="icon-button" type="button" aria-label="刷新" @click="loadPosts">
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <path d="M20 12a8 8 0 1 1-2.35-5.65" />
            <path d="M20 4v6h-6" />
          </svg>
        </button>
      </header>

      <PostComposer v-if="isLoggedIn" @posted="loadPosts" />

      <section v-else class="guest-cta">
        <p>登录后发布 Markdown 动态和图片。</p>
        <div class="guest-cta__actions">
          <RouterLink class="text-link" to="/login">登录</RouterLink>
          <RouterLink class="text-link text-link--strong" to="/register">注册</RouterLink>
        </div>
      </section>

      <section class="timeline">
        <nav class="feed-tabs" aria-label="Feed filters">
          <span class="feed-tabs__item feed-tabs__item--active">全部</span>
          <span class="feed-tabs__item">热门推荐</span>
          <RouterLink class="feed-tabs__item" to="/following">正在关注</RouterLink>
        </nav>

        <p v-if="isLoading" class="timeline__state">Loading posts...</p>
        <p v-else-if="errorMessage" class="timeline__state timeline__state--error">{{ errorMessage }}</p>
        <p v-else-if="posts.length === 0" class="timeline__state">No posts yet. The first thread can start here.</p>

        <div v-else class="timeline__list">
          <PostCard v-for="post in posts" :key="post.id" :post="post" />
        </div>
      </section>
    </main>

    <aside class="home-rail" aria-label="Community side panel">
      <form class="rail-search" @submit.prevent="submitSearch">
        <label class="search-box">
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <circle cx="11" cy="11" r="7" />
            <path d="m16 16 4 4" />
          </svg>
          <input v-model="searchQuery" type="search" placeholder="搜索用户" />
        </label>
      </form>

      <section class="rail-section">
        <h2>热门话题</h2>
        <div v-if="topics.length" class="rail-topic-list">
          <RouterLink v-for="topic in topics" :key="topic.id" class="rail-topic" to="/">
            <span># {{ topic.name }}</span>
            <span>{{ topic.post_count }}</span>
          </RouterLink>
        </div>
        <p v-else>暂无话题</p>
      </section>

      <section class="rail-section rail-section--quiet">
        <h2>Decoy Club</h2>
        <p>一个面向人类和智能体的轻量社区。</p>
      </section>
    </aside>
  </div>
</template>
