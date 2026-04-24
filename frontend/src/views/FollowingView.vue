<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { fetchFollowingPosts, type ApiPost } from '../api/client'
import PostCard from '../components/PostCard.vue'
import { useAuthStore } from '../stores/auth'

const authStore = useAuthStore()
const posts = ref<ApiPost[]>([])
const isLoading = ref(true)
const errorMessage = ref('')

const hasSession = computed(() => Boolean(authStore.token))

async function loadFollowingPosts() {
  if (!authStore.token) {
    posts.value = []
    isLoading.value = false
    return
  }

  isLoading.value = true
  errorMessage.value = ''
  try {
    const response = await fetchFollowingPosts(authStore.token)
    posts.value = response.posts
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Failed to load following timeline'
  } finally {
    isLoading.value = false
  }
}

onMounted(loadFollowingPosts)
</script>

<template>
  <section class="notifications-page">
    <div class="timeline">
      <header class="timeline__header">
        <div>
          <p class="timeline__eyebrow">Following feed</p>
          <h2>Posts from followed accounts</h2>
        </div>
        <button class="ghost-button" @click="loadFollowingPosts">Refresh</button>
      </header>

      <p v-if="!hasSession" class="timeline__state">Sign in to see posts from accounts you follow.</p>
      <p v-else-if="isLoading" class="timeline__state">Loading following timeline...</p>
      <p v-else-if="errorMessage" class="timeline__state timeline__state--error">{{ errorMessage }}</p>
      <p v-else-if="posts.length === 0" class="timeline__state">You are not following anyone with posts yet.</p>

      <div v-else class="timeline__list">
        <PostCard v-for="post in posts" :key="post.id" :post="post" />
      </div>
    </div>
  </section>
</template>
