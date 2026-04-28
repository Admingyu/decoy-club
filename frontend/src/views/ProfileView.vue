<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import {
  fetchActivityCounts,
  fetchProfile,
  fetchUserActivity,
  fetchUserPosts,
  followUser,
  unfollowUser,
  updateMyStatus,
  type ActivityType,
  type ApiActivityCounts,
  type ApiActivityItem,
  type ApiPost,
  type ApiProfile,
} from '../api/client'
import MarkdownContent from '../components/MarkdownContent.vue'
import PostCard from '../components/PostCard.vue'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const authStore = useAuthStore()

const profile = ref<ApiProfile | null>(null)
const isLoading = ref(true)
const isUpdatingFollow = ref(false)
const isSavingStatus = ref(false)
const isActivityLoading = ref(false)
const errorMessage = ref('')
const posts = ref<ApiPost[]>([])
const activityCounts = ref<ApiActivityCounts>({ views: 0, likes: 0, comments: 0 })
const activeActivity = ref<ActivityType | null>(null)
const activities = ref<ApiActivityItem[]>([])
const statusText = ref('')
const statusPreset = ref('')
const statusOptions = ['在线', '忙碌', '开发中', '摸鱼中', '潜水']

const username = computed(() => String(route.params.username ?? authStore.user?.username ?? ''))
const isOwnProfile = computed(() => username.value !== '' && username.value === String(authStore.user?.username ?? ''))
const canViewActivity = computed(() => Boolean(authStore.token && isOwnProfile.value))

async function loadProfile() {
  if (!username.value) {
    profile.value = null
    isLoading.value = false
    return
  }

  isLoading.value = true
  errorMessage.value = ''
  try {
    const [profileResponse, postsResponse, countsResponse] = await Promise.all([
      fetchProfile(username.value, authStore.token || undefined),
      fetchUserPosts(username.value, authStore.token || undefined),
      fetchActivityCounts(username.value, authStore.token || undefined),
    ])
    profile.value = profileResponse.profile
    posts.value = postsResponse.posts
    activityCounts.value = countsResponse.counts
    statusText.value = profileResponse.profile.status_text || ''
    statusPreset.value = profileResponse.profile.status_preset || ''
    activeActivity.value = null
    activities.value = []
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Failed to load profile'
  } finally {
    isLoading.value = false
  }
}

async function toggleFollow() {
  if (!authStore.token || !profile.value || isOwnProfile.value) {
    return
  }

  isUpdatingFollow.value = true
  errorMessage.value = ''
  try {
    const response = profile.value.following
      ? await unfollowUser(authStore.token, username.value)
      : await followUser(authStore.token, username.value)
    profile.value = response.profile
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Failed to update follow state'
  } finally {
    isUpdatingFollow.value = false
  }
}

function chooseStatus(option: string) {
  statusPreset.value = option
  statusText.value = option
}

async function saveStatus() {
  if (!authStore.token || !isOwnProfile.value) {
    return
  }

  isSavingStatus.value = true
  errorMessage.value = ''
  try {
    const response = await updateMyStatus(authStore.token, statusText.value, statusPreset.value)
    profile.value = response.profile
    statusText.value = response.profile.status_text || ''
    statusPreset.value = response.profile.status_preset || ''
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Failed to update status'
  } finally {
    isSavingStatus.value = false
  }
}

async function loadActivity(type: ActivityType) {
  if (!authStore.token || !canViewActivity.value) {
    return
  }

  activeActivity.value = type
  isActivityLoading.value = true
  errorMessage.value = ''
  try {
    const response = await fetchUserActivity(authStore.token, username.value, type)
    activities.value = response.activities
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Failed to load activity'
  } finally {
    isActivityLoading.value = false
  }
}

function activityTitle(type: ActivityType | null) {
  if (type === 'views') return '浏览记录'
  if (type === 'likes') return '点赞记录'
  if (type === 'comments') return '评论记录'
  return '记录'
}

function formatActivityTime(value: string) {
  return new Intl.DateTimeFormat('zh-CN', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(value))
}

watch(() => route.params.username, () => {
  void loadProfile()
}, { immediate: true })
</script>

<template>
  <section class="notifications-page">
    <div class="timeline">
      <p v-if="isLoading" class="timeline__state">Loading profile...</p>
      <p v-else-if="errorMessage" class="timeline__state timeline__state--error">{{ errorMessage }}</p>

      <template v-else-if="profile">
        <header class="profile-hero">
          <div class="profile-hero__identity">
            <div class="profile-hero__avatar">{{ profile.username.slice(0, 2).toUpperCase() }}</div>
            <div>
              <p class="timeline__eyebrow">Profile</p>
              <h2>@{{ profile.username }}</h2>
              <p>{{ profile.bio || 'No bio yet.' }}</p>
              <p v-if="profile.status_text" class="profile-status">
                {{ profile.status_text }}
              </p>
            </div>
          </div>

          <button
            v-if="authStore.token && !isOwnProfile"
            class="solid-button"
            :disabled="isUpdatingFollow"
            @click="toggleFollow"
          >
            {{ isUpdatingFollow ? 'Updating...' : profile.following ? 'Unfollow' : 'Follow' }}
          </button>
        </header>

        <section v-if="isOwnProfile" class="status-editor">
          <div class="status-editor__options">
            <button
              v-for="option in statusOptions"
              :key="option"
              class="ghost-button status-editor__option"
              :class="{ 'status-editor__option--active': statusPreset === option }"
              type="button"
              @click="chooseStatus(option)"
            >
              {{ option }}
            </button>
          </div>
          <div class="status-editor__row">
            <input
              v-model="statusText"
              class="status-editor__input"
              maxlength="80"
              placeholder="设置你的状态"
              @input="statusPreset = ''"
            />
            <button class="solid-button" :disabled="isSavingStatus" @click="saveStatus">
              {{ isSavingStatus ? 'Saving...' : 'Save' }}
            </button>
          </div>
        </section>

        <div class="profile-grid">
          <div class="sidebar__panel"><h2>Posts</h2><p>{{ profile.post_count }}</p></div>
          <div class="sidebar__panel"><h2>Replies</h2><p>{{ profile.reply_count }}</p></div>
          <div class="sidebar__panel"><h2>Followers</h2><p>{{ profile.followers_count }}</p></div>
          <div class="sidebar__panel"><h2>Following</h2><p>{{ profile.following_count }}</p></div>
          <button class="sidebar__panel profile-stat" type="button" :disabled="!canViewActivity" @click="loadActivity('views')">
            <h2>Views</h2><p>{{ activityCounts.views }}</p>
          </button>
          <button class="sidebar__panel profile-stat" type="button" :disabled="!canViewActivity" @click="loadActivity('likes')">
            <h2>Likes</h2><p>{{ activityCounts.likes }}</p>
          </button>
          <button class="sidebar__panel profile-stat" type="button" :disabled="!canViewActivity" @click="loadActivity('comments')">
            <h2>Comments</h2><p>{{ activityCounts.comments }}</p>
          </button>
          <div class="sidebar__panel"><h2>Received likes</h2><p>{{ profile.received_like_count }}</p></div>
        </div>

        <section v-if="activeActivity" class="timeline profile-activity">
          <header class="timeline__header">
            <div>
              <p class="timeline__eyebrow">Activity</p>
              <h2>{{ activityTitle(activeActivity) }}</h2>
            </div>
          </header>

          <p v-if="isActivityLoading" class="timeline__state">Loading records...</p>
          <p v-else-if="activities.length === 0" class="timeline__state">No records yet.</p>
          <div v-else class="timeline__list">
            <article v-for="activity in activities" :key="activity.id" class="activity-card">
              <div class="activity-card__meta">
                <span>{{ formatActivityTime(activity.updated_at) }}</span>
                <span v-if="activity.type === 'view' && activity.count > 1">x{{ activity.count }}</span>
              </div>
              <PostCard v-if="activity.post" :post="activity.post" />
              <RouterLink
                v-if="activity.comment"
                class="activity-card__comment"
                :to="`/posts/${activity.comment.post_id}`"
              >
                <p class="timeline__eyebrow">Comment</p>
                <MarkdownContent :content="activity.comment.content_markdown" />
              </RouterLink>
            </article>
          </div>
        </section>

        <section class="timeline profile-posts">
          <header class="timeline__header">
            <div>
              <p class="timeline__eyebrow">Recent posts</p>
              <h2>@{{ profile.username }}'s threads</h2>
            </div>
          </header>

          <p v-if="posts.length === 0" class="timeline__state">No public posts yet.</p>
          <div v-else class="timeline__list">
            <PostCard v-for="post in posts" :key="post.id" :post="post" />
          </div>
        </section>
      </template>
    </div>
  </section>
</template>
