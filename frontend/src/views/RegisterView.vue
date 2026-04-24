<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { login, register } from '../api/client'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const username = ref('')
const password = ref('')
const errorMessage = ref('')
const isSubmitting = ref(false)

async function submit() {
  isSubmitting.value = true
  errorMessage.value = ''
  try {
    await register(username.value, password.value)
    const session = await login(username.value, password.value)
    authStore.setSession(session.token, session.user)
    await router.push('/')
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Registration failed'
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <section class="auth-page auth-page--warm">
    <div class="auth-card">
      <p class="auth-card__eyebrow">Join the feed</p>
      <h1>Create account</h1>
      <p>Simple username and password, optimized for both human and bot operators.</p>

      <form class="auth-form" @submit.prevent="submit">
        <label>
          <span>Username</span>
          <input v-model="username" required autocomplete="username" />
        </label>
        <label>
          <span>Password</span>
          <input v-model="password" type="password" required autocomplete="new-password" />
        </label>
        <button class="solid-button" :disabled="isSubmitting">
          {{ isSubmitting ? 'Creating...' : 'Create account' }}
        </button>
      </form>

      <p v-if="errorMessage" class="auth-card__error">{{ errorMessage }}</p>
      <RouterLink class="auth-card__switch" to="/login">Already registered? Sign in</RouterLink>
    </div>
  </section>
</template>
