<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { createComment, fetchPost, fetchPostComments, likePost, unlikePost, type ApiComment, type ApiPost } from '../api/client'
import CommentThread from '../components/CommentThread.vue'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const authStore = useAuthStore()

const post = ref<ApiPost | null>(null)
const comments = ref<ApiComment[]>([])
const isLoading = ref(true)
const isCommentSubmitting = ref(false)
const isLikeSubmitting = ref(false)
const errorMessage = ref('')
const commentContent = ref('')

const postId = computed(() => String(route.params.postId ?? ''))
const isLoggedIn = computed(() => Boolean(authStore.token))
const liked = computed(() => Boolean(post.value?.liked_by_viewer))

function formatTime(value: string) {
  return new Intl.DateTimeFormat('zh-CN', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(value))
}

async function loadPostDetail() {
  if (!postId.value) {
    return
  }

  isLoading.value = true
  errorMessage.value = ''
  try {
    const [postResponse, commentsResponse] = await Promise.all([
      fetchPost(postId.value),
      fetchPostComments(postId.value),
    ])
    post.value = postResponse.post
    comments.value = commentsResponse.comments
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Failed to load post detail'
  } finally {
    isLoading.value = false
  }
}

async function submitComment() {
  if (!authStore.token || !commentContent.value.trim() || !postId.value) {
    return
  }

  isCommentSubmitting.value = true
  errorMessage.value = ''
  try {
    await createComment(authStore.token, postId.value, commentContent.value.trim())
    commentContent.value = ''
    await loadPostDetail()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Failed to comment'
  } finally {
    isCommentSubmitting.value = false
  }
}

async function toggleLike() {
  if (!authStore.token || !post.value || isLikeSubmitting.value) {
    return
  }

  isLikeSubmitting.value = true
  try {
    if (post.value.liked_by_viewer) {
      await unlikePost(authStore.token, post.value.id)
      post.value = {
        ...post.value,
        liked_by_viewer: false,
        like_count: Math.max(0, post.value.like_count - 1),
      }
    } else {
      await likePost(authStore.token, post.value.id)
      post.value = {
        ...post.value,
        liked_by_viewer: true,
        like_count: post.value.like_count + 1,
      }
    }
  } finally {
    isLikeSubmitting.value = false
  }
}

onMounted(loadPostDetail)
</script>

<template>
  <section class="detail-page">
    <div class="timeline">
      <p v-if="isLoading" class="timeline__state">Loading post...</p>
      <p v-else-if="errorMessage" class="timeline__state timeline__state--error">{{ errorMessage }}</p>

      <template v-else-if="post">
        <article class="post-card post-card--detail">
          <header class="post-card__header">
            <div class="post-card__avatar">{{ post.author_id.slice(0, 2).toUpperCase() }}</div>
            <div>
              <strong class="post-card__author">
                <RouterLink
                  class="post-card__author-link"
                  :to="post.author_username ? `/u/${post.author_username}` : '#'"
                >
                  {{ post.author_username ? `@${post.author_username}` : `User ${post.author_id.slice(0, 6)}` }}
                </RouterLink>
              </strong>
              <p class="post-card__time">{{ formatTime(post.created_at) }}</p>
            </div>
          </header>

          <div class="post-card__body" v-html="post.content_html" />

          <div v-if="post.embedded_images.length" class="post-card__images">
            <img v-for="image in post.embedded_images" :key="image" class="post-card__image" :src="image" alt="embedded image" />
          </div>

          <footer class="post-card__meta">
            <button v-if="isLoggedIn" class="post-card__like-button" :disabled="isLikeSubmitting" @click="toggleLike">
              {{ liked ? 'Unlike' : 'Like' }}
            </button>
            <span>{{ post.like_count }} likes</span>
            <span>{{ post.comment_count }} comments</span>
          </footer>
        </article>

        <section class="timeline">
          <header class="timeline__header">
            <div>
              <p class="timeline__eyebrow">Discussion</p>
              <h2>Comments and replies</h2>
            </div>
          </header>

          <div v-if="isLoggedIn" class="composer composer--compact">
            <textarea
              v-model="commentContent"
              class="composer__textarea composer__textarea--compact"
              rows="5"
              placeholder="Join the discussion..."
            />
            <div class="composer__actions">
              <button class="solid-button" :disabled="isCommentSubmitting || !commentContent.trim()" @click="submitComment">
                {{ isCommentSubmitting ? 'Posting...' : 'Post comment' }}
              </button>
            </div>
          </div>

          <p v-else class="timeline__state">Sign in to reply and participate in the thread.</p>
          <p v-if="!comments.length" class="timeline__state">No comments yet. Start the conversation.</p>

          <div v-else class="comment-list">
            <CommentThread
              v-for="comment in comments"
              :key="comment.id"
              :comment="comment"
              @changed="loadPostDetail"
            />
          </div>
        </section>
      </template>
    </div>
  </section>
</template>
