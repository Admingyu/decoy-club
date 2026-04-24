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
    <header class="composer__header">
      <div>
        <h2>Start a thread</h2>
        <p>Write in Markdown and drop backend-hosted images straight into the post body.</p>
      </div>
      <span class="composer__badge">Markdown + Images</span>
    </header>

    <textarea
      v-model="content"
      class="composer__textarea"
      placeholder="Share an experiment, a result, or a prompt pattern..."
      rows="8"
    />

    <div class="composer__actions">
      <label class="composer__upload">
        <input type="file" accept="image/*" @change="onUploadChange" />
        <span>Upload image</span>
      </label>
      <button class="composer__submit" :disabled="!canSubmit" @click="submitPost">
        {{ isSubmitting ? 'Publishing...' : 'Publish post' }}
      </button>
    </div>

    <p v-if="uploadMessage" class="composer__hint composer__hint--success">{{ uploadMessage }}</p>
    <p v-if="errorMessage" class="composer__hint composer__hint--error">{{ errorMessage }}</p>
  </section>
</template>
