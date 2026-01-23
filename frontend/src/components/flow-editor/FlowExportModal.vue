<script setup>
import { ref, watch } from 'vue'
import { useFlowStore } from '../../stores/flow'

const props = defineProps({
  modelValue: Boolean
})

const emit = defineEmits(['update:modelValue', 'error'])

const flowStore = useFlowStore()
const exportJSON = ref('')
const saveError = ref('')
const copied = ref(false)

// Generate JSON when modal opens
watch(() => props.modelValue, (isOpen) => {
  if (isOpen) {
    const elements = flowStore.export()
    exportJSON.value = JSON.stringify(elements, null, 2)
    saveError.value = ''
    copied.value = false
  }
})

const closeModal = () => {
  emit('update:modelValue', false)
  exportJSON.value = ''
  saveError.value = ''
  copied.value = false
}

const exportToFile = async () => {
  saveError.value = ''
  try {
    const filename = `${flowStore.flowResourceName || 'flow'}.json`
    const savedPath = await window.go.main.App.SaveFile(filename, exportJSON.value)
    if (savedPath) {
      closeModal()
    }
  } catch (e) {
    saveError.value = e.message || 'Failed to save file'
    emit('error', saveError.value)
  }
}

const copyToClipboard = async () => {
  try {
    await navigator.clipboard.writeText(exportJSON.value)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch (e) {
    saveError.value = 'Failed to copy to clipboard'
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
      class="fixed inset-0 bg-gray-500 bg-opacity-25 dark:bg-black dark:bg-opacity-75 backdrop-blur-sm"
      @click="closeModal"
    ></div>

    <!-- Modal -->
    <div class="relative transform rounded-lg bg-white text-left shadow-xl transition-all sm:my-8 p-1 w-full max-w-3xl mx-auto dark:bg-black dark:border dark:border-gray-800 dark:text-gray-300">
      <h3 class="text-center sm:mt-3 font-medium text-gray-900 dark:text-gray-100">
        Export Flow JSON
      </h3>

      <!-- Textarea -->
      <div class="h-full">
        <textarea
          v-model="exportJSON"
          readonly
          class="mt-1 border-sky-600 h-56 max-w-full placeholder-gray-400 focus:ring-sky-600 appearance-none border rounded w-full py-3 px-3 text-gray-700 leading-tight transition duration-150 ease-in-out sm:text-sm sm:leading-5 dark:bg-gray-900 dark:text-gray-300"
        ></textarea>
      </div>

      <!-- Error message -->
      <div v-if="saveError" class="text-red-500 text-sm py-2 px-1">
        {{ saveError }}
      </div>

      <!-- Buttons -->
      <div class="flex justify-between p-3">
        <!-- Export to file button -->
        <button
          @click="exportToFile"
          type="button"
          class="text-gray-500 bg-white hover:bg-gray-100 focus:ring-4 focus:outline-none focus:ring-gray-200 rounded-md border border-gray-200 text-sm font-medium px-3 py-1 hover:text-gray-900 focus:z-10 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-500 dark:hover:text-white dark:hover:bg-gray-600 dark:focus:ring-gray-600"
        >
          Export File...
        </button>

        <div class="flex gap-2">
          <button
            @click="copyToClipboard"
            type="button"
            class="text-gray-500 bg-white hover:bg-gray-100 focus:ring-4 focus:outline-none focus:ring-gray-200 rounded-md border border-gray-200 text-sm font-medium px-3 py-1 hover:text-gray-900 focus:z-10 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-500 dark:hover:text-white dark:hover:bg-gray-600 dark:focus:ring-gray-600"
          >
            {{ copied ? 'Copied!' : 'Copy' }}
          </button>
          <button
            @click="closeModal"
            type="button"
            class="text-gray-500 bg-white hover:bg-gray-100 focus:ring-4 focus:outline-none focus:ring-gray-200 rounded-md border border-gray-200 text-sm font-medium px-3 py-1 hover:text-gray-900 focus:z-10 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-500 dark:hover:text-white dark:hover:bg-gray-600 dark:focus:ring-gray-600"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
