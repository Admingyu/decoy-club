<script setup lang="ts">
import { computed, ref } from 'vue'
import { likePost, unlikePost, type ApiPost } from '../api/client'
import { useAuthStore } from '../stores/auth'

const props = defineProps<{
  post: ApiPost
}>()

const authStore = useAuthStore()
const liked = ref(Boolean(props.post.liked_by_viewer))
const localLikeCount = ref(props.post.like_count)
const isLiking = ref(false)
const canLike = computed(() => Boolean(authStore.token))
const embeddedImages = computed(() => Array.isArray(props.post.embedded_images) ? props.post.embedded_images : [])
const authorName = computed(() => props.post.author_username ?? `User ${props.post.author_id.slice(0, 6)}`)
const authorHandle = computed(() => props.post.author_username ? `@${props.post.author_username}` : props.post.author_id.slice(0, 8))
const topics = computed(() => Array.isArray(props.post.topics) ? props.post.topics : [])

function formatTime(value: string) {
  return new Intl.DateTimeFormat('zh-CN', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(value))
}

async function toggleLike() {
  if (!authStore.token || isLiking.value) {
    return
  }

  isLiking.value = true
  try {
    if (liked.value) {
      await unlikePost(authStore.token, props.post.id)
      liked.value = false
      localLikeCount.value = Math.max(0, localLikeCount.value - 1)
    } else {
      await likePost(authStore.token, props.post.id)
      liked.value = true
      localLikeCount.value += 1
    }
  } finally {
    isLiking.value = false
  }
}
</script>

<template>
  <article class="post-card post-card--clickable">
    <div class="post-card__row">
      <div class="post-card__avatar">{{ post.author_id.slice(0, 2).toUpperCase() }}</div>
      <div class="post-card__main">
        <header class="post-card__header">
          <div class="post-card__identity">
            <strong class="post-card__author">
              <RouterLink
                v-if="post.author_username"
                class="post-card__author-link"
                :to="`/u/${post.author_username}`"
              >
                {{ authorName }}
              </RouterLink>
              <span v-else>{{ authorName }}</span>
            </strong>
            <span class="post-card__handle">{{ authorHandle }}</span>
          </div>
          <p class="post-card__time">{{ formatTime(post.created_at) }}</p>
        </header>

        <div class="post-card__content">
          <div class="post-card__body" v-html="post.content_html" />

          <div v-if="embeddedImages.length" class="post-card__images">
            <img
              v-for="image in embeddedImages"
              :key="image"
              class="post-card__image"
              :src="image"
              alt="embedded image"
            />
          </div>

          <div v-if="topics.length" class="post-card__topics">
            <RouterLink v-for="topic in topics" :key="topic" class="post-card__topic" to="/">
              # {{ topic }}
            </RouterLink>
          </div>
        </div>

        <footer class="post-card__meta">
          <button
            v-if="canLike"
            class="post-action post-action--button"
            :class="{ 'post-action--liked': liked }"
            type="button"
            :disabled="isLiking"
            :aria-label="liked ? '取消点赞' : '点赞'"
            @click="toggleLike"
          >
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M20.8 4.6a5.4 5.4 0 0 0-7.6 0L12 5.8l-1.2-1.2a5.4 5.4 0 0 0-7.6 7.6L12 21l8.8-8.8a5.4 5.4 0 0 0 0-7.6z" />
            </svg>
            <span>{{ localLikeCount }}</span>
          </button>
          <RouterLink v-else class="post-action" to="/login" aria-label="登录后点赞">
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M20.8 4.6a5.4 5.4 0 0 0-7.6 0L12 5.8l-1.2-1.2a5.4 5.4 0 0 0-7.6 7.6L12 21l8.8-8.8a5.4 5.4 0 0 0 0-7.6z" />
            </svg>
            <span>{{ localLikeCount }}</span>
          </RouterLink>
          <RouterLink class="post-action" :to="`/posts/${post.id}`">
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M21 15a4 4 0 0 1-4 4H8l-5 3V7a4 4 0 0 1 4-4h10a4 4 0 0 1 4 4z" />
            </svg>
            <span>{{ post.comment_count }}</span>
          </RouterLink>
        </footer>
      </div>
    </div>
  </article>
</template>
