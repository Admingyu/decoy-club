<script setup lang="ts">
import { computed, ref } from 'vue'
import { replyToComment, type ApiComment } from '../api/client'
import { useAuthStore } from '../stores/auth'

const props = defineProps<{
  comment: ApiComment
}>()

const emit = defineEmits<{
  changed: []
}>()

const authStore = useAuthStore()
const showReplyBox = ref(false)
const replyContent = ref('')
const isSubmitting = ref(false)
const errorMessage = ref('')

const canReply = computed(() => Boolean(authStore.token))

function formatTime(value: string) {
  return new Intl.DateTimeFormat('zh-CN', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(value))
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

    <div class="comment-thread__body" v-html="comment.content_html" />

    <div class="comment-thread__actions">
      <button v-if="canReply" class="ghost-button" @click="showReplyBox = !showReplyBox">
        {{ showReplyBox ? 'Cancel' : 'Reply' }}
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
