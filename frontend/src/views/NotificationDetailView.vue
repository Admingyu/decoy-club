<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { fetchNotificationDetail, type NotificationListItem } from '../api/client'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const authStore = useAuthStore()
const notification = ref<NotificationListItem | null>(null)
const isLoading = ref(true)
const errorMessage = ref('')

function titleFor(item: NotificationListItem) {
  if (item.type === 'post_liked') {
    return `${item.actor?.username ?? 'Someone'} liked your post`
  }
  if (item.type === 'post_commented') {
    return `${item.actor?.username ?? 'Someone'} commented on your post`
  }
  return `${item.actor?.username ?? 'Someone'} replied to your comment`
}

async function loadNotification() {
  if (!authStore.token) {
    errorMessage.value = 'Sign in required'
    isLoading.value = false
    return
  }

  isLoading.value = true
  errorMessage.value = ''
  try {
    const response = await fetchNotificationDetail(authStore.token, String(route.params.notificationId ?? ''))
    notification.value = response.notification
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Failed to load notification detail'
  } finally {
    isLoading.value = false
  }
}

onMounted(loadNotification)
</script>

<template>
  <section class="notifications-page">
    <div class="timeline">
      <p v-if="isLoading" class="timeline__state">Loading notification detail...</p>
      <p v-else-if="errorMessage" class="timeline__state timeline__state--error">{{ errorMessage }}</p>

      <template v-else-if="notification">
        <header class="timeline__header">
          <div>
            <p class="timeline__eyebrow">Notification detail</p>
            <h2>{{ titleFor(notification) }}</h2>
          </div>
          <RouterLink v-if="notification.post?.id" class="ghost-button" :to="`/posts/${notification.post.id}`">
            Open post
          </RouterLink>
        </header>

        <article class="post-card">
          <div class="notification-block">
            <p v-if="notification.post?.content_markdown"><strong>Post markdown</strong></p>
            <pre v-if="notification.post?.content_markdown" class="notification-detail__code">{{ notification.post.content_markdown }}</pre>

            <p v-if="notification.comment?.content_markdown"><strong>Comment markdown</strong></p>
            <pre v-if="notification.comment?.content_markdown" class="notification-detail__code">{{ notification.comment.content_markdown }}</pre>

            <p v-if="notification.parent_comment?.content_markdown"><strong>Parent comment markdown</strong></p>
            <pre v-if="notification.parent_comment?.content_markdown" class="notification-detail__code">{{ notification.parent_comment.content_markdown }}</pre>
          </div>
        </article>
      </template>
    </div>
  </section>
</template>
