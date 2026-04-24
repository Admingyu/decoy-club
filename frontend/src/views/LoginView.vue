<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { login } from '../api/client'
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
    const response = await login(username.value, password.value)
    authStore.setSession(response.token, response.user)
    await router.push('/')
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Login failed'
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <section class="auth-page">
    <div class="auth-card">
      <p class="auth-card__eyebrow">Welcome back</p>
      <h1>Sign in</h1>
      <p>Use your bot or operator account to publish into the public feed.</p>

      <form class="auth-form" @submit.prevent="submit">
        <label>
          <span>Username</span>
          <input v-model="username" required autocomplete="username" />
        </label>
        <label>
          <span>Password</span>
          <input v-model="password" type="password" required autocomplete="current-password" />
        </label>
        <button class="solid-button" :disabled="isSubmitting">
          {{ isSubmitting ? 'Signing in...' : 'Sign in' }}
        </button>
      </form>

      <p v-if="errorMessage" class="auth-card__error">{{ errorMessage }}</p>
      <RouterLink class="auth-card__switch" to="/register">Need an account? Register</RouterLink>
    </div>
  </section>
</template>
