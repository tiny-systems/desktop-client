<script setup>
import { ref, watch, nextTick } from 'vue'

const props = defineProps({
  modelValue: Boolean,
  contextName: String,
  namespace: String,
  projectName: String
})

const emit = defineEmits(['update:modelValue', 'error', 'success'])

const importJSON = ref('')
const parseError = ref('')
const loading = ref(false)
const textareaRef = ref(null)

// Focus textarea when modal opens
watch(() => props.modelValue, (newVal) => {
  if (newVal) {
    nextTick(() => {
      textareaRef.value?.focus()
    })
  }
})

const GoApp = window.go?.main?.App

const closeModal = () => {
  emit('update:modelValue', false)
  importJSON.value = ''
  parseError.value = ''
}

const importProject = async () => {
  parseError.value = ''
  loading.value = true
  try {
    // Validate JSON format
    const data = JSON.parse(importJSON.value)
    if (!data.version || !data.tinyFlows || !data.elements) {
      throw new Error('Invalid project export format. Expected version, tinyFlows, and elements fields.')
    }

    await GoApp.ImportProject(props.contextName, props.namespace, props.projectName, importJSON.value)
    emit('success')
    closeModal()
  } catch (e) {
    parseError.value = e?.message || (typeof e === 'string' ? e : 'Invalid JSON')
    emit('error', parseError.value)
  } finally {
    loading.value = false
  }
}

const importFromFile = async () => {
  parseError.value = ''
  try {
    const content = await GoApp.OpenFile()
    if (content) {
      importJSON.value = content
    }
  } catch (e) {
    parseError.value = e.message || 'Failed to open file'
  }
}
</script>

<template>
  <div
    v-if="modelValue"
    class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6 md:p-20"
    @keydown.escape="closeModal"
  >
    <!-- Backdrop -->
    <div
      class="fixed inset-0 bg-gray-500/25 dark:bg-black/75 backdrop-blur-sm"
      @click="closeModal"
    ></div>

    <!-- Modal -->
    <div class="relative transform rounded-lg bg-white text-left shadow-xl transition-all sm:my-8 p-1 w-full max-w-3xl mx-auto dark:bg-black dark:border dark:border-gray-800 dark:text-gray-300">
      <h3 class="text-center sm:mt-3 font-medium text-gray-900 dark:text-gray-100">
        Import Project JSON
      </h3>

      <!-- Textarea -->
      <div class="h-full">
        <textarea
          ref="textareaRef"
          v-model="importJSON"
          placeholder="Paste project JSON here..."
          class="mt-1 border-sky-600 h-56 max-w-full placeholder-gray-400 focus:ring-sky-600 appearance-none border rounded w-full py-3 px-3 text-gray-700 leading-tight transition duration-150 ease-in-out sm:text-sm sm:leading-5 dark:bg-gray-900 dark:text-gray-300"
        ></textarea>
      </div>

      <!-- Error message -->
      <div v-if="parseError" class="text-red-500 text-sm py-2 px-1">
        {{ parseError }}
      </div>

      <!-- Buttons -->
      <div class="flex justify-between p-3">
        <!-- Import from file button -->
        <button
          @click="importFromFile"
          type="button"
          class="text-gray-500 bg-white hover:bg-gray-100 focus:ring-4 focus:outline-none focus:ring-gray-200 rounded-md border border-gray-200 text-sm font-medium px-3 py-1 hover:text-gray-900 focus:z-10 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-500 dark:hover:text-white dark:hover:bg-gray-600 dark:focus:ring-gray-600"
        >
          Import File...
        </button>

        <div class="flex gap-2">
          <button
            @click="closeModal"
            type="button"
            class="text-gray-500 bg-white hover:bg-gray-100 focus:ring-4 focus:outline-none focus:ring-gray-200 rounded-md border border-gray-200 text-sm font-medium px-3 py-1 hover:text-gray-900 focus:z-10 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-500 dark:hover:text-white dark:hover:bg-gray-600 dark:focus:ring-gray-600"
          >
            Cancel
          </button>
          <button
            @click="importProject"
            type="button"
            :disabled="!importJSON.trim() || loading"
            class="text-gray-500 bg-white hover:bg-gray-100 focus:ring-4 focus:outline-none focus:ring-gray-200 rounded-md border border-gray-200 text-sm font-medium px-3 py-1 hover:text-gray-900 focus:z-10 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-500 dark:hover:text-white dark:hover:bg-gray-600 dark:focus:ring-gray-600 disabled:opacity-50"
          >
            {{ loading ? 'Importing...' : 'Import' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
