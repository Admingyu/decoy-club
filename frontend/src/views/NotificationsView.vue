<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { fetchNotifications, markNotificationsRead, type NotificationListItem } from '../api/client'
import { useAuthStore } from '../stores/auth'

const authStore = useAuthStore()
const notifications = ref<NotificationListItem[]>([])
const isLoading = ref(true)
const errorMessage = ref('')

const unreadIds = computed(() => notifications.value.filter((item) => !item.is_read).map((item) => item.id))

function titleFor(item: NotificationListItem) {
  if (item.type === 'post_liked') {
    return `${item.actor?.username ?? 'Someone'} liked your post`
  }
  if (item.type === 'post_commented') {
    return `${item.actor?.username ?? 'Someone'} commented on your post`
  }
  return `${item.actor?.username ?? 'Someone'} replied to your comment`
}

function formatTime(value: string) {
  return new Intl.DateTimeFormat('zh-CN', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(value))
}

async function loadNotifications() {
  if (!authStore.token) {
    notifications.value = []
    isLoading.value = false
    return
  }

  isLoading.value = true
  errorMessage.value = ''
  try {
    const response = await fetchNotifications(authStore.token)
    notifications.value = response.notifications
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Failed to load notifications'
  } finally {
    isLoading.value = false
  }
}

async function markAllRead() {
  if (!authStore.token || unreadIds.value.length === 0) {
    return
  }
  await markNotificationsRead(authStore.token, unreadIds.value)
  await loadNotifications()
}

onMounted(loadNotifications)
</script>

<template>
  <section class="notifications-page">
    <div class="timeline">
      <header class="timeline__header">
        <div>
          <p class="timeline__eyebrow">Unread center</p>
          <h2>Notifications</h2>
        </div>
        <button class="ghost-button" :disabled="unreadIds.length === 0" @click="markAllRead">
          Mark all read
        </button>
      </header>

      <p v-if="isLoading" class="timeline__state">Loading notifications...</p>
      <p v-else-if="errorMessage" class="timeline__state timeline__state--error">{{ errorMessage }}</p>
      <p v-else-if="notifications.length === 0" class="timeline__state">No notifications yet.</p>

      <div v-else class="timeline__list">
        <article
          v-for="item in notifications"
          :key="item.id"
          class="post-card"
          :style="{ borderColor: item.is_read ? 'rgba(29, 27, 25, 0.08)' : 'rgba(203, 92, 43, 0.35)' }"
        >
          <header class="post-card__header">
            <div class="post-card__avatar">{{ item.actor?.username?.slice(0, 2).toUpperCase() ?? 'NT' }}</div>
            <div>
              <strong class="post-card__author">
                <RouterLink v-if="item.actor?.username" class="notification-block__link" :to="`/notifications/${item.id}`">
                  {{ titleFor(item) }}
                </RouterLink>
                <template v-else>{{ titleFor(item) }}</template>
              </strong>
              <p class="post-card__time">{{ formatTime(item.created_at) }}</p>
            </div>
          </header>

          <div class="notification-block">
            <p>
              <strong>Post</strong>:
              <RouterLink v-if="item.post?.id" class="notification-block__link" :to="`/posts/${item.post.id}`">
                {{ item.post?.content_preview ?? 'No post context' }}
              </RouterLink>
              <template v-else>
                {{ item.post?.content_preview ?? 'No post context' }}
              </template>
            </p>
            <p v-if="item.comment"><strong>Comment</strong>: {{ item.comment.content_preview }}</p>
            <p v-if="item.parent_comment"><strong>Parent comment</strong>: {{ item.parent_comment.content_preview }}</p>
          </div>
        </article>
      </div>
    </div>
  </section>
</template>
