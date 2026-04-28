<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { likeComment, replyToComment, unlikeComment, type ApiComment } from '../api/client'
import { useAuthStore } from '../stores/auth'
import MarkdownContent from './MarkdownContent.vue'

const props = defineProps<{
  comment: ApiComment
}>()

const emit = defineEmits<{
  changed: []
}>()

const authStore = useAuthStore()
const liked = ref(Boolean(props.comment.liked_by_viewer))
const localLikeCount = ref(props.comment.like_count ?? 0)
const isLiking = ref(false)
const showReplyBox = ref(false)
const replyContent = ref('')
const isSubmitting = ref(false)
const errorMessage = ref('')

const canLike = computed(() => Boolean(authStore.token))
const canReply = computed(() => Boolean(authStore.token))

watch(
  () => props.comment,
  (comment) => {
    liked.value = Boolean(comment.liked_by_viewer)
    localLikeCount.value = comment.like_count ?? 0
  },
)

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
      await unlikeComment(authStore.token, props.comment.id)
      liked.value = false
      localLikeCount.value = Math.max(0, localLikeCount.value - 1)
    } else {
      await likeComment(authStore.token, props.comment.id)
      liked.value = true
      localLikeCount.value += 1
    }
  } finally {
    isLiking.value = false
  }
}

async function submitReply() {
  if (!authStore.token || !replyContent.value.trim()) {
    return
  }

  isSubmitting.value = true
  errorMessage.value = ''
  try {
    await replyToComment(authStore.token, props.comment.id, replyContent.value.trim())
    replyContent.value = ''
    showReplyBox.value = false
    emit('changed')
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Failed to reply'
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <article class="comment-thread">
    <header class="comment-thread__header">
      <div class="comment-thread__avatar">{{ comment.author_id.slice(0, 2).toUpperCase() }}</div>
      <div>
        <strong class="post-card__author">
          <RouterLink
            class="post-card__author-link"
            :to="comment.author_username ? `/u/${comment.author_username}` : '#'"
          >
            {{ comment.author_username ? `@${comment.author_username}` : `User ${comment.author_id.slice(0, 6)}` }}
          </RouterLink>
        </strong>
        <p class="post-card__time">{{ formatTime(comment.created_at) }}</p>
      </div>
    </header>

    <MarkdownContent class="comment-thread__body" :content="comment.content_markdown" />

    <div class="comment-thread__actions">
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
      <button
        v-if="canReply"
        class="post-action post-action--button"
        type="button"
        :aria-label="showReplyBox ? '取消回复' : '回复'"
        @click="showReplyBox = !showReplyBox"
      >
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path d="M21 15a4 4 0 0 1-4 4H8l-5 3V7a4 4 0 0 1 4-4h10a4 4 0 0 1 4 4z" />
        </svg>
        <span>{{ showReplyBox ? 'Cancel' : 'Reply' }}</span>
      </button>
    </div>

    <div v-if="showReplyBox" class="comment-thread__composer">
      <textarea
        v-model="replyContent"
        class="composer__textarea comment-thread__textarea"
        rows="4"
        placeholder="Write a reply..."
      />
      <div class="comment-thread__submit-row">
        <button class="solid-button" :disabled="isSubmitting || !replyContent.trim()" @click="submitReply">
          {{ isSubmitting ? 'Replying...' : 'Publish reply' }}
        </button>
      </div>
      <p v-if="errorMessage" class="composer__hint composer__hint--error">{{ errorMessage }}</p>
    </div>

    <div v-if="comment.replies.length" class="comment-thread__replies">
      <CommentThread
        v-for="reply in comment.replies"
        :key="reply.id"
        :comment="reply"
        @changed="emit('changed')"
      />
    </div>
  </article>
</template>
