<script setup lang="ts">
import { computed, ref } from 'vue'
import { createPost, uploadImage } from '../api/client'
import { useAuthStore } from '../stores/auth'

const emit = defineEmits<{
  posted: []
}>()

const authStore = useAuthStore()
const content = ref('')
const imageUrls = ref<string[]>([])
const isSubmitting = ref(false)
const errorMessage = ref('')
const uploadMessage = ref('')

const canSubmit = computed(() => Boolean(authStore.token && content.value.trim()) && !isSubmitting.value)

async function onUploadChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file || !authStore.token) {
    return
  }

  uploadMessage.value = 'Uploading image...'
  errorMessage.value = ''
  try {
    const response = await uploadImage(authStore.token, file)
    imageUrls.value = [...imageUrls.value, response.file.public_url]
    content.value = `${content.value}${content.value ? '\n' : ''}![${response.file.file_name}](${response.file.public_url})`
    uploadMessage.value = 'Image uploaded and inserted into your post.'
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Failed to upload image'
    uploadMessage.value = ''
  } finally {
    input.value = ''
  }
}

async function submitPost() {
  if (!canSubmit.value || !authStore.token) {
    return
  }

  isSubmitting.value = true
  errorMessage.value = ''
  try {
    await createPost(authStore.token, content.value.trim(), imageUrls.value)
    content.value = ''
    imageUrls.value = []
    uploadMessage.value = ''
    emit('posted')
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Failed to publish post'
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <section class="composer">
    <textarea
      v-model="content"
      class="composer__textarea"
      placeholder="Share an experiment, a result, or a prompt pattern..."
      rows="8"
    />

    <div class="composer__actions">
      <label class="composer__upload" aria-label="Upload image" title="Upload image">
        <input type="file" accept="image/*" @change="onUploadChange" />
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <rect x="3" y="5" width="18" height="14" rx="2" />
          <circle cx="8.5" cy="10" r="1.5" />
          <path d="m21 15-4.5-4.5L9 18" />
        </svg>
      </label>
      <button
        class="composer__submit"
        :disabled="!canSubmit"
        :aria-label="isSubmitting ? 'Publishing post' : 'Publish post'"
        :title="isSubmitting ? 'Publishing post' : 'Publish post'"
        @click="submitPost"
      >
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path d="M22 2 11 13" />
          <path d="m22 2-7 20-4-9-9-4 20-7z" />
        </svg>
      </button>
    </div>

    <p v-if="uploadMessage" class="composer__hint composer__hint--success">{{ uploadMessage }}</p>
    <p v-if="errorMessage" class="composer__hint composer__hint--error">{{ errorMessage }}</p>
  </section>
</template>
