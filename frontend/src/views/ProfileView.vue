<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { fetchProfile, fetchUserPosts, followUser, unfollowUser, type ApiPost, type ApiProfile } from '../api/client'
import PostCard from '../components/PostCard.vue'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const authStore = useAuthStore()

const profile = ref<ApiProfile | null>(null)
const isLoading = ref(true)
const isUpdatingFollow = ref(false)
const errorMessage = ref('')
const posts = ref<ApiPost[]>([])

const username = computed(() => String(route.params.username ?? authStore.user?.username ?? ''))
const isOwnProfile = computed(() => username.value !== '' && username.value === String(authStore.user?.username ?? ''))

async function loadProfile() {
  if (!username.value) {
    profile.value = null
    isLoading.value = false
    return
  }

  isLoading.value = true
  errorMessage.value = ''
  try {
    const [profileResponse, postsResponse] = await Promise.all([
      fetchProfile(username.value, authStore.token || undefined),
      fetchUserPosts(username.value, authStore.token || undefined),
    ])
    profile.value = profileResponse.profile
    posts.value = postsResponse.posts
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

watch(() => route.params.username, () => {
  void loadProfile()
}, { immediate: true })

onMounted(loadProfile)
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

        <div class="profile-grid">
          <div class="sidebar__panel"><h2>Posts</h2><p>{{ profile.post_count }}</p></div>
          <div class="sidebar__panel"><h2>Replies</h2><p>{{ profile.reply_count }}</p></div>
          <div class="sidebar__panel"><h2>Followers</h2><p>{{ profile.followers_count }}</p></div>
          <div class="sidebar__panel"><h2>Following</h2><p>{{ profile.following_count }}</p></div>
          <div class="sidebar__panel"><h2>Received likes</h2><p>{{ profile.received_like_count }}</p></div>
          <div class="sidebar__panel"><h2>Given likes</h2><p>{{ profile.given_like_count }}</p></div>
        </div>

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
