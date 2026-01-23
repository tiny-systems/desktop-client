<script setup>
import { ref } from 'vue'
import { useFlowStore } from '../../stores/flow'

const props = defineProps({
  modelValue: Boolean
})

const emit = defineEmits(['update:modelValue', 'error'])

const flowStore = useFlowStore()
const importFlowJSON = ref('')
const parseError = ref('')

const closeModal = () => {
  emit('update:modelValue', false)
  importFlowJSON.value = ''
  parseError.value = ''
}

const importFlow = async () => {
  parseError.value = ''
  try {
    const elements = JSON.parse(importFlowJSON.value)
    if (!Array.isArray(elements)) {
      throw new Error('Expected an array of elements')
    }
    await flowStore.import(elements)
    closeModal()
  } catch (e) {
    parseError.value = e.message || 'Invalid JSON'
    emit('error', parseError.value)
  }
}
</script>

<template>
  <div
    v-if="modelValue"
    class="fixed inset-0 z-50 flex items-center justify-center p-4"
    @keydown.escape="closeModal"
  >
    <!-- Backdrop -->
    <div
      class="fixed inset-0 bg-gray-500/25 dark:bg-black/75 backdrop-blur-sm"
      @click="closeModal"
    ></div>

    <!-- Modal -->
    <div class="relative bg-white dark:bg-gray-900 dark:border dark:border-gray-700 rounded-lg shadow-xl w-full max-w-3xl p-4">
      <h3 class="text-base font-semibold text-gray-900 dark:text-gray-100 mb-3 text-center">
        Import Flow JSON
      </h3>

      <!-- Textarea -->
      <div class="mb-3">
        <textarea
          v-model="importFlowJSON"
          placeholder="Paste flow JSON here..."
          class="w-full h-56 px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 placeholder-gray-400 focus:ring-1 focus:ring-sky-500 focus:border-sky-500 font-mono"
          autofocus
        ></textarea>
      </div>

      <!-- Error message -->
      <div v-if="parseError" class="mb-3 text-sm text-red-500">
        {{ parseError }}
      </div>

      <!-- Buttons -->
      <div class="flex justify-end gap-2">
        <button
          @click="closeModal"
          class="px-3 py-2 text-xs font-medium text-gray-700 dark:text-gray-300 border border-gray-600 rounded-md hover:bg-gray-100 dark:hover:bg-gray-800"
        >
          Cancel
        </button>
        <button
          @click="importFlow"
          :disabled="!importFlowJSON.trim() || flowStore.loading"
          class="px-3 py-2 text-xs font-medium text-white bg-sky-600 rounded-md hover:bg-sky-700 disabled:opacity-50"
        >
          {{ flowStore.loading ? 'Importing...' : 'Import' }}
        </button>
      </div>
    </div>
  </div>
</template>
