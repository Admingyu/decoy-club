<script setup lang="ts">
import { computed, nextTick, ref } from 'vue'
import { createPost, uploadImage } from '../api/client'
import { useAuthStore } from '../stores/auth'
import { convertRichHtmlToMarkdown, getClipboardImageFiles } from './richPaste'

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

type TextareaInsertion = {
  textarea: HTMLTextAreaElement
  start: number
  end: number
}

async function onUploadChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file || !authStore.token) {
    return
  }

  await uploadAndInsertImages([file])
  input.value = ''
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

async function onPaste(event: ClipboardEvent) {
  const pastedImages = getClipboardImageFiles(event.clipboardData)
  if (pastedImages.length) {
    event.preventDefault()
    const textarea = event.target as HTMLTextAreaElement
    await uploadAndInsertImages(pastedImages, {
      textarea,
      start: textarea.selectionStart ?? content.value.length,
      end: textarea.selectionEnd ?? textarea.selectionStart ?? content.value.length,
    })
    return
  }

  const html = event.clipboardData?.getData('text/html')
  if (!html) {
    return
  }

  const markdown = convertRichHtmlToMarkdown(html)
  if (!markdown) {
    return
  }

  event.preventDefault()
  insertMarkdown(markdown, {
    textarea: event.target as HTMLTextAreaElement,
    start: (event.target as HTMLTextAreaElement).selectionStart ?? content.value.length,
    end: (event.target as HTMLTextAreaElement).selectionEnd ?? (event.target as HTMLTextAreaElement).selectionStart ?? content.value.length,
  })
}

async function uploadAndInsertImages(files: File[], insertion?: TextareaInsertion) {
  if (!authStore.token || files.length === 0) {
    return
  }

  uploadMessage.value = files.length === 1 ? 'Uploading image...' : 'Uploading images...'
  errorMessage.value = ''
  try {
    const uploadedFiles = []
    for (const file of files) {
      const response = await uploadImage(authStore.token, file)
      uploadedFiles.push(response.file)
    }

    imageUrls.value = [...imageUrls.value, ...uploadedFiles.map((file) => file.public_url)]
    insertMarkdown(uploadedFiles.map((file) => `![${file.file_name}](${file.public_url})`).join('\n'), insertion)
    uploadMessage.value = uploadedFiles.length === 1 ? 'Image uploaded and inserted into your post.' : 'Images uploaded and inserted into your post.'
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Failed to upload image'
    uploadMessage.value = ''
  }
}

function insertMarkdown(markdown: string, insertion?: TextareaInsertion) {
  const start = insertion?.start ?? content.value.length
  const end = insertion?.end ?? start
  const before = content.value.slice(0, start)
  const after = content.value.slice(end)
  const prefix = before && !before.endsWith('\n') ? '\n' : ''
  const suffix = after && !after.startsWith('\n') ? '\n' : ''
  content.value = `${before}${prefix}${markdown}${suffix}${after}`

  if (!insertion) {
    return
  }

  void nextTick(() => {
    const cursor = before.length + prefix.length + markdown.length
    insertion.textarea.setSelectionRange(cursor, cursor)
  })
}
</script>

<template>
  <section class="composer">
    <textarea
      v-model="content"
      class="composer__textarea"
      placeholder="Share an experiment, a result, or a prompt pattern..."
      rows="8"
      @paste="onPaste"
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
