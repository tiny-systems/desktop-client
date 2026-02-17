<script setup>
import { ref, watch, nextTick, onUnmounted } from 'vue'
import { EventsOn, EventsOff } from '../../../wailsjs/runtime/runtime'

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
const importMessage = ref('')
const importDone = ref(false)
const textareaRef = ref(null)

// Listen for progress events from Go backend
const startListening = () => {
  EventsOn('import:progress', (msg) => {
    importMessage.value = msg
  })
}

const stopListening = () => {
  EventsOff('import:progress')
}

onUnmounted(stopListening)

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
  importMessage.value = ''
  importDone.value = false
  stopListening()
}

const confirmSuccess = () => {
  emit('success')
  closeModal()
}

const copyErrors = async () => {
  try {
    await navigator.clipboard.writeText(parseError.value)
  } catch (e) {
    // fallback
  }
}

const importProject = async () => {
  parseError.value = ''
  importMessage.value = ''
  importDone.value = false
  loading.value = true
  startListening()
  try {
    // Validate JSON format
    const data = JSON.parse(importJSON.value)
    if (!data.version || !data.tinyFlows || !data.elements) {
      throw new Error('Invalid project export format. Expected version, tinyFlows, and elements fields.')
    }

    await GoApp.ImportProject(props.contextName, props.namespace, props.projectName, importJSON.value)
    importDone.value = true
    importMessage.value = importMessage.value || 'Import complete!'
  } catch (e) {
    parseError.value = e?.message || (typeof e === 'string' ? e : 'Invalid JSON')
    emit('error', parseError.value)
  } finally {
    loading.value = false
    stopListening()
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

      <!-- Success state -->
      <div v-if="importDone" class="px-3 py-6">
        <div class="flex flex-col items-center gap-3">
          <svg class="h-10 w-10 text-green-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <p class="text-sm text-gray-700 dark:text-gray-300">{{ importMessage }}</p>
        </div>
        <div class="flex justify-end mt-4">
          <button
            @click="confirmSuccess"
            type="button"
            class="text-white bg-sky-600 hover:bg-sky-700 focus:ring-4 focus:outline-none focus:ring-sky-300 rounded-md text-sm font-medium px-4 py-1.5 dark:bg-sky-700 dark:hover:bg-sky-600"
          >
            OK
          </button>
        </div>
      </div>

      <!-- Normal state -->
      <template v-else>
        <!-- Textarea -->
        <div class="h-full">
          <textarea
            ref="textareaRef"
            v-model="importJSON"
            placeholder="Paste project JSON here..."
            class="mt-1 border-sky-600 h-56 max-w-full placeholder-gray-400 focus:ring-sky-600 appearance-none border rounded w-full py-3 px-3 text-gray-700 leading-tight transition duration-150 ease-in-out sm:text-sm sm:leading-5 dark:bg-gray-900 dark:text-gray-300"
            :disabled="loading"
          ></textarea>
        </div>

        <!-- Error message -->
        <div v-if="parseError" class="px-1 py-2">
          <div class="flex items-start gap-2">
            <div class="flex-1 max-h-40 overflow-y-auto rounded border border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-950/30 p-2">
              <pre class="text-red-600 dark:text-red-400 text-xs whitespace-pre-wrap font-mono">{{ parseError }}</pre>
            </div>
            <button
              @click="copyErrors"
              type="button"
              title="Copy errors"
              class="shrink-0 p-1.5 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 border border-gray-200 dark:border-gray-700 rounded"
            >
              <svg class="h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" d="M15.666 3.888A2.25 2.25 0 0013.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v0a.75.75 0 01-.75.75H9.75a.75.75 0 01-.75-.75v0c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 01-2.25 2.25H6.75A2.25 2.25 0 014.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 011.927-.184" />
              </svg>
            </button>
          </div>
        </div>

        <!-- Progress display -->
        <div v-if="loading && importMessage" class="flex items-center py-2 px-1">
          <svg class="animate-spin h-4 w-4 mr-2 text-sky-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          <span class="text-sm text-gray-700 dark:text-gray-300">{{ importMessage }}</span>
        </div>

        <!-- Buttons -->
        <div class="flex justify-between p-3">
          <!-- Import from file button -->
          <button
            @click="importFromFile"
            type="button"
            :disabled="loading"
            class="text-gray-500 bg-white hover:bg-gray-100 focus:ring-4 focus:outline-none focus:ring-gray-200 rounded-md border border-gray-200 text-sm font-medium px-3 py-1 hover:text-gray-900 focus:z-10 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-500 dark:hover:text-white dark:hover:bg-gray-600 dark:focus:ring-gray-600 disabled:opacity-50"
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
      </template>
    </div>
  </div>
</template>
